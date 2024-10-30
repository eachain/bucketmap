package bucketmap

import (
	"math/rand"
	"sync"

	"github.com/eachain/unsafehash"
)

type bucket[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

// Map is like a Go map[K]V but is safe for concurrent use
// by multiple goroutines without additional locking or coordination.
//
// The Map type splits keys to different buckets.
// It like a simple Go map[K]V when buckets size is 1.
type Map[K comparable, V any] struct {
	buckets []bucket[K, V]
	hash    unsafehash.HashFunc[K]
}

// Make makes a Map with default 31 buckets.
func Make[K comparable, V any](buckets ...int) *Map[K, V] {
	n := 31
	if len(buckets) > 0 && buckets[0] > 0 {
		n = buckets[0]
	}
	var hash unsafehash.HashFunc[K]
	if n == 1 {
		hash = func(k K) uint64 { return 0 }
	} else {
		hash = unsafehash.Map[K]()
	}
	return &Map[K, V]{
		buckets: make([]bucket[K, V], n),
		hash:    hash,
	}
}

func (m *Map[K, V]) get(key K) *bucket[K, V] {
	return &m.buckets[m.hash(key)%uint64(len(m.buckets))]
}

// Load returns the value stored in the map for a key,
// or zero value if no value is present.
// The ok result indicates whether value was found in the map.
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	bkt := m.get(key)
	bkt.RLock()
	value, ok = bkt.m[key]
	bkt.RUnlock()
	return
}

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) {
	bkt := m.get(key)
	bkt.Lock()
	if bkt.m == nil {
		bkt.m = make(map[K]V)
	}
	bkt.m[key] = value
	bkt.Unlock()
}

// Delete deletes the value for a key.
func (m *Map[K, V]) Delete(key K) {
	bkt := m.get(key)
	bkt.Lock()
	delete(bkt.m, key)
	bkt.Unlock()
}

// Clear deletes all the entries, resulting in an empty Map.
func (m *Map[K, V]) Clear() {
	for i := 0; i < len(m.buckets); i++ {
		b := &m.buckets[i]
		b.Lock()
		clear(b.m)
		b.Unlock()
	}
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	bkt := m.get(key)
	bkt.Lock()
	value, loaded = bkt.m[key]
	if loaded {
		delete(bkt.m, key)
	}
	bkt.Unlock()
	return
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	bkt := m.get(key)
	bkt.Lock()
	actual, loaded = bkt.m[key]
	if !loaded {
		if bkt.m == nil {
			bkt.m = make(map[K]V)
		}
		bkt.m[key] = value
		actual = value
	}
	bkt.Unlock()
	return
}

// LoadOrStoreFunc returns the existing value for the key if present.
// Otherwise, it stores and returns the given value which is returned by newValue func.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[K, V]) LoadOrStoreFunc(key K, newValue func() V) (actual V, loaded bool) {
	bkt := m.get(key)
	bkt.Lock()
	defer bkt.Unlock()
	actual, loaded = bkt.m[key]
	if !loaded {
		if bkt.m == nil {
			bkt.m = make(map[K]V)
		}
		value := newValue()
		bkt.m[key] = value
		actual = value
	}
	return
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	bkt := m.get(key)
	bkt.Lock()
	defer bkt.Unlock()
	previous, loaded = bkt.m[key]
	if bkt.m == nil {
		bkt.m = make(map[K]V)
	}
	bkt.m[key] = value
	return
}

// Iter returns an iterator over key-value pairs in the Map.
func (m *Map[K, V]) Iter() func(yield func(K, V) bool) {
	order := make([]int, len(m.buckets))
	for i := range order {
		order[i] = i
	}
	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	return func(yield func(K, V) bool) {
		for _, i := range order {
			broken := false
			bkt := &m.buckets[i]
			f := func(k K, v V) bool {
				bkt.RUnlock()
				defer bkt.RLock()
				return yield(k, v)
			}
			func() {
				bkt.RLock()
				defer bkt.RUnlock()

				for k, v := range bkt.m {
					if !f(k, v) {
						broken = true
						return
					}
				}
			}()
			if broken {
				break
			}
		}
	}
}
