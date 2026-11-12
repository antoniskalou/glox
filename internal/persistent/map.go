package persistent

import (
	"cmp"
	"iter"
	"math/rand"
	"sync/atomic"
)

type Map[K cmp.Ordered, V any] struct {
	root *mapNode[K, V]
}

func (*Map[K, V]) less(l, r K) bool {
	return l < r
}

type mapNode[K cmp.Ordered, V any] struct {
	key         K
	value       *refValue[K, V]
	weight      uint64
	refCount    int32
	left, right *mapNode[K, V]
}

type refValue[K cmp.Ordered, V any] struct {
	refCount int32
	value    V
	release  func(key K, val V)
}

func newNodeWithRef[K cmp.Ordered, V any](
	key K,
	val V,
	release func(key K, val V),
) *mapNode[K, V] {
	return &mapNode[K, V]{
		key: key,
		value: &refValue[K, V]{
			value:    val,
			release:  release,
			refCount: 1,
		},
		refCount: 1,
		weight:   rand.Uint64(),
	}
}

func (node *mapNode[K, V]) shallowCloneWithRef() *mapNode[K, V] {
	atomic.AddInt32(&node.value.refCount, 1)
	return &mapNode[K, V]{
		key:      node.key,
		value:    node.value,
		weight:   node.weight,
		refCount: 1,
	}
}

func (node *mapNode[K, V]) incref() *mapNode[K, V] {
	if node != nil {
		atomic.AddInt32(&node.refCount, 1)
	}
	return node
}

func (node *mapNode[K, V]) decref() {
	if node == nil {
		return
	}

	if atomic.AddInt32(&node.refCount, -1) == 0 {
		if node.value.release != nil {
			node.value.release(node.key, node.value.value)
		}
		node.value = nil
	}
	node.left.decref()
	node.right.decref()
}

func (m Map[K, V]) Clone() Map[K, V] {
	return Map[K, V]{
		root: m.root.incref(),
	}
}

func (m Map[K, V]) Destroy() {
	m.Clear()
}

func (m Map[K, V]) Clear() {
	m.root.decref()
	m.root = nil
}

func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.root.forEach(func(k K, _ V) bool {
			return yield(k)
		})
	}
}

func (m *Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.root.forEach(yield)
	}
}

func (node *mapNode[K, V]) forEach(yield func(key K, val V) bool) bool {
	return node.left.forEach(yield) &&
		yield(node.key, node.value.value) &&
		node.right.forEach(yield)
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	node := m.root
	for node != nil {
		if key < node.key {
			node = node.left
		} else if key > node.key {
			node = node.right
		} else {
			return node.value.value, true
		}
	}

	var zero V
	return zero, false
}

func (m *Map[K, V]) SetAll(other Map[K, V]) {
	root := m.root
	m.root = union(root, other.root, m.less, true)
	root.decref()
}

func (m *Map[K, V]) Set(key K, val V, release func(key K, val V)) {
	first := m.root
	second := newNodeWithRef(key, val, release)
	m.root = union(first, second, m.less, true)
	first.decref()
	second.decref()
}

func (m *Map[K, V]) Delete(key K) bool {
	root := m.root
	left, mid, right := split(root, key, m.less, true)
	if mid == nil {
		return false
	}
	m.root = merge(left, right)
	left.decref()
	mid.decref()
	right.decref()
	root.decref()
	return true
}

// returns a new tree which is a union of the first and second one
func union[K cmp.Ordered, V any](
	first, second *mapNode[K, V],
	less func(K, K) bool,
	overwrite bool,
) *mapNode[K, V] {
	if first == nil {
		return second.incref()
	}

	if second == nil {
		return first.incref()
	}

	if first.weight < second.weight {
		second, first, overwrite = first, second, !overwrite
	}

	left, mid, right := split(second, first.key, less, false)
	var result *mapNode[K, V]
	if overwrite && mid != nil {
		result = mid.shallowCloneWithRef()
	} else {
		result = first.shallowCloneWithRef()
	}
	result.weight = first.weight
	result.left = union(first.left, left, less, overwrite)
	result.right = union(first.right, right, less, overwrite)
	left.decref()
	mid.decref()
	right.decref()
	return result
}

// split the tree midway by the key into three different ones
func split[K cmp.Ordered, V any](
	n *mapNode[K, V],
	key K,
	less func(K, K) bool,
	requireMid bool,
) (left, mid, right *mapNode[K, V]) {
	if n == nil {
		return nil, nil, nil
	}

	if less(n.key, key) {
		left, mid, right := split(n.right, key, less, requireMid)
		if requireMid && mid == nil {
			return nil, nil, nil
		}
		newN := n.shallowCloneWithRef()
		newN.left = n.left.incref()
		newN.right = left
		return newN, mid, right
	} else if less(key, n.key) {
		left, mid, right := split(n.left, key, less, requireMid)
		if requireMid && mid == nil {
			return nil, nil, nil
		}
		newN := n.shallowCloneWithRef()
		newN.left = right
		newN.right = n.right.incref()
		return left, mid, newN
	}
	mid = n.shallowCloneWithRef()
	return n.left.incref(), mid, n.right.incref()
}

func merge[K cmp.Ordered, V any](left, right *mapNode[K, V]) *mapNode[K, V] {
	switch {
	case left == nil:
		return right.incref()
	case right == nil:
		return left.incref()
	case left.weight > right.weight:
		root := left.shallowCloneWithRef()
		root.left = left.left.incref()
		root.right = merge(left.right, right)
		return root
	default:
		root := right.shallowCloneWithRef()
		root.left = merge(left, right.left)
		root.right = right.right.incref()
		return root
	}
}
