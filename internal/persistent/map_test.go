package persistent

import "testing"

func TestSimpleMap(t *testing.T) {
	parent := new(Map[int, int])

	// missing key
	if v, ok := parent.Get(0); ok {
		t.Errorf("expected key %d to be missing, got ok", v)
	}

	parent.Set(0, 1, nil)
	if _, ok := parent.Get(0); !ok {
		t.Error("expected key 0 to be be present")
	}

	cloned := parent.Clone()
	cloned.Set(0, 2, nil)
	// cloned updated
	if v, _ := cloned.Get(0); v != 2 {
		t.Errorf("expected key 0 to be be 2, got %d", v)
	}
	// original stays the same
	if v, _ := parent.Get(0); v != 1 {
		t.Errorf("original map mutated")
	}

	// clearning the clone doesn't affect the parent
	cloned.Destroy()
	if v, _ := parent.Get(0); v != 1 {
		t.Error("destroyed clone should not affect parent")
	}
}
