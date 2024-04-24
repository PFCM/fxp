package vector

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/pfcm/fxp/fix"
)

var sizes = []int{1, 3, 53, 127, 1009, 100007}

var vmas17s = []struct {
	name string
	f    func(a, b, c, out []fix.S17, n int)
}{
	{"simple", vmas17_simple},
	{"unsafe", vmas17_unsafe},
	{"____wg", vmas17_wg},
	{"___asm", _vmas17_serial},
}

func BenchmarkVMAS17(b *testing.B) {
	for _, size := range sizes {
		size := size
		b.Run(fmt.Sprintf("%6d", size), func(b *testing.B) {
			var (
				x   = randS17s(size)
				y   = randS17s(size)
				z   = randS17s(size)
				out = randS17s(size)
			)
			for _, f := range vmas17s {
				f := f
				b.Run(f.name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						f.f(x, y, z, out, size)
					}
				})
			}
		})
	}
}

func TestVMAS17(t *testing.T) {
	for _, size := range sizes {
		for _, f := range vmas17s {
			if f.name == "simple" {
				continue
			}
			t.Run(fmt.Sprintf("%s/%d", f.name, size), func(t *testing.T) {
				var (
					a    = newTestDatum("a", randS17s, size)
					b    = newTestDatum("b", randS17s, size)
					c    = newTestDatum("c", randS17s, size)
					out  = newTestDatum("out", randS17s, size)
					want = make([]fix.S17, size)
				)
				vmas17_simple(a.get(), b.get(), c.get(), want, size)

				f.f(a.get(), b.get(), c.get(), out.get(), size)

				a.check(t, a.get())
				b.check(t, b.get())
				c.check(t, c.get())
				out.check(t, want)
			})
		}
	}
}

const pad = 13

type testDatum[T comparable] struct {
	name      string
	b         []T
	pre, post []T
}

func newTestDatum[T comparable](name string, init func(int) []T, size int) testDatum[T] {
	b := init(size + pad*2)
	pre, post := make([]T, pad), make([]T, pad)
	copy(pre, b)
	copy(post, b[size+pad:])
	return testDatum[T]{name: name, b: b, pre: pre, post: post}
}

func (t testDatum[T]) get() []T {
	return t.b[pad : len(t.b)-pad]
}

func (td testDatum[T]) check(t *testing.T, want []T) {
	t.Helper()
	if diff := cmp.Diff(td.b[pad:len(td.b)-pad], want); diff != "" {
		t.Errorf("arg %q: unexpected diff (-got,+want):\n%v", td.name, diff)
	}
	if diff := cmp.Diff(td.b[:pad], td.pre); diff != "" {
		t.Errorf("arg %q: diff before slice (-got,+want):\n%v", td.name, diff)
	}
	if diff := cmp.Diff(td.b[len(td.b)-pad:], td.post); diff != "" {
		t.Errorf("arg %q: diff after slice (-got,+want):\n%v", td.name, diff)
	}
}

func randS17s(n int) []fix.S17 {
	b := make([]byte, n)
	rand.Read(b)
	out := make([]fix.S17, n)
	for i := range b {
		out[i] = fix.S17(b[i])
	}
	return out
}
