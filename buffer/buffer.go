// Package buffer implements a pool of pointers to byte slices.
//
// Example usage pattern
//
//	p := buffer.Get(size)
//	b := *p	// Now you can use b in any way you need.
//	...
//	// When b will not be used anymore
//	buffer.Put(p)
//	...
//	// If b or p are not going out of scope soon, optionally
//	b = nil
//	p = nil
//
// Otherwise the pool cannot release the buffer on garbage collection.
//
// Do not do
//
//	p := buffer.Get(size)
//	b := *p
//	...
//	buffer.Put(&b)
//
// or
//
//	b := *buffer.Get(size)
//	...
//	buffer.Put(&b)
package buffer

import (
	"sync"

	"github.com/cznic/mathutil"
)

var (
	m    [63]sync.Pool
	null []byte
)

func init() {
	for i := range m {
		size := 1 << uint(i)
		m[i] = sync.Pool{New: func() interface{} {
			// 0:     1 -      1
			// 1:    10 -     10
			// 2:    11 -    100
			// 3:   101 -   1000
			// 4:  1001 -  10000
			// 5: 10001 - 100000
			b := make([]byte, size)
			return &b
		}}
	}
}

// Get returns a pointer to a byte slice of len size. The pointed to byte slice
// is zeroed up to its cap. Get panics for size < 0.
//
// Get is safe for concurrent use by multiple goroutines.
func Get(size int) *[]byte {
	var index int
	switch {
	case size < 0:
		panic("buffer.Get: negative size")
	case size == 0:
		return &null
	case size > 1:
		index = mathutil.Log2Uint64(uint64(size-1)) + 1
	}
	p := m[index].Get().(*[]byte)
	b := *p
	for i := range b[:cap(b)] {
		b[i] = 0
	}
	*p = b[:size]
	return p
}

// Put puts a pointer to a byte slice into a pool for possible later reuse by
// Get.
//
// Put is safe for concurrent use by multiple goroutines.
func Put(p *[]byte) {
	if p == nil {
		return
	}

	b := *p
	size := cap(b)
	if size == 0 {
		return
	}

	b = b[:size]
	*p = b
	m[mathutil.Log2Uint64(uint64(size))].Put(p)
}
