package buffer

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "\tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("TODO: %s:%d:\n", path.Base(fn), fl) //TODOOK
}

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO) //TODOOK
}

// ============================================================================

func Test(t *testing.T) {
	a := [1 << 10]*[]byte{}
	m := map[*[]byte]struct{}{}
	for i := range a {
		p := Get(i)
		if _, ok := m[p]; ok {
			t.Fatal(i)
		}

		a[i] = p
		m[p] = struct{}{}
		b := *p
		for j := range b {
			b[j] = 123
		}
	}
	for i := range a {
		Put(a[i])
	}
	for i := range a {
		p := Get(i)
		if _, ok := m[p]; !ok {
			t.Fatal(i)
		}

		delete(m, p)
		b := *p
		if g, e := len(b), i; g != e {
			t.Fatal(g, e)
		}

		for j, v := range b[:cap(b)] {
			if v != 0 {
				t.Fatal(i, j, v)
			}
		}
	}
}

func test(t testing.TB, allocs, goroutines int) {
	ready := make(chan int, goroutines)
	run := make(chan int)
	done := make(chan int, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			a := rand.Perm(allocs)
			ready <- 1
			<-run
			for _, v := range a {
				p := Get(v)
				b := *p
				if g, e := len(b), v; g != e {
					t.Error(g, e)
					break
				}

				for i, d := range b[:cap(b)] {
					if d != 0 {
						t.Error(v, i, d)
						break
					}
				}

				for i := range b {
					b[i] = 123
				}
				Put(p)
			}
			done <- 1
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-ready
	}
	close(run)
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func Test2(t *testing.T) {
	test(t, 1<<14, 16)
}

func Benchmark1(b *testing.B) {
	const (
		allocs     = 1000
		goroutines = 100
	)
	for i := 0; i < b.N; i++ {
		test(b, allocs, goroutines)
	}
	b.SetBytes(goroutines * (allocs*allocs + allocs) / 2)
}
