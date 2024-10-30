package bucketmap

import "testing"

func TestMap(t *testing.T) {
	m := Make[int, string]()
	if value, ok := m.Load(123); ok {
		t.Fatalf("load 123: %v", value)
	}

	m.Store(123, "abc")
	if value, ok := m.Load(123); !ok {
		t.Fatalf("load 123: not exists")
	} else if value != "abc" {
		t.Fatalf("load 123: %v", value)
	}

	m.Delete(123)
	if value, ok := m.Load(123); ok {
		t.Fatalf("load 123: %v", value)
	}

	m.Store(123, "abc")
	m.Clear()
	if value, ok := m.Load(123); ok {
		t.Fatalf("load 123: %v", value)
	}

	m.Store(123, "abc")
	if value, loaded := m.LoadAndDelete(123); !loaded {
		t.Fatalf("load 123: not exists")
	} else if value != "abc" {
		t.Fatalf("load 123: %v", value)
	}

	if value, loaded := m.LoadOrStore(123, "abc"); loaded {
		t.Fatalf("load 123: %v", value)
	} else if value != "abc" {
		t.Fatalf("load 123: %v", value)
	}

	if value, loaded := m.LoadOrStore(123, "abc"); !loaded {
		t.Fatalf("load 123: not exists")
	} else if value != "abc" {
		t.Fatalf("load 123: %v", value)
	}

	if previous, loaded := m.Swap(123, "def"); !loaded {
		t.Fatalf("load 123: not exists")
	} else if previous != "abc" {
		t.Fatalf("load 123: %v", previous)
	}

	if value, loaded := m.LoadOrStoreFunc(123, func() string { return "abc" }); !loaded {
		t.Fatalf("load 123: not exists")
	} else if value != "def" {
		t.Fatalf("load 123: %v", value)
	}

	m.Delete(123)
	if value, loaded := m.LoadOrStoreFunc(123, func() string { return "abc" }); loaded {
		t.Fatalf("load 123: %v", value)
	} else if value != "abc" {
		t.Fatalf("load 123: %v", value)
	}

	m.Iter()(func(key int, value string) bool {
		m.Delete(key)
		return true
	})
}
