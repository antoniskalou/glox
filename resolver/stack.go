package resolver

import (
	"iter"
	"slices"
)

type stack[T any] struct {
	items []T
}

func (s *stack[T]) push(item T) {
	s.items = append(s.items, item)
}

func (s *stack[T]) pop() (T, bool) {
	if s.isEmpty() {
		var zero T
		return zero, false
	}

	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *stack[T]) peek() (T, bool) {
	if s.isEmpty() {
		var zero T
		return zero, false
	}

	return s.items[len(s.items)-1], true
}

func (s *stack[T]) size() int {
	return len(s.items)
}

func (s *stack[T]) isEmpty() bool {
	return len(s.items) == 0
}

func (s *stack[T]) iter() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range slices.Backward(s.items) {
			if !yield(i, v) {
				return
			}
		}
	}
}
