package vector

import (
	"sync"
	"unsafe"

	"github.com/pfcm/fxp/fix"
)

// VMAS17 does a vector multiply-add with fix.S17s. It takes 4 vector arguments
// a, b, c and out, and computes out[i] = a[i]*b[i]+c[i] for each i up to n.
func VMAS17(a, b, c, out []fix.S17, n int) {
	vmas17_simple(a, b, c, out, n)
}

func vmas17_simple(a, b, c, out []fix.S17, n int) {
	for i := 0; i < n; i++ {
		out[i] = a[i].SMul(b[i]).SAdd(c[i])
	}
}

func vmas17_unsafe(a, b, c, out []fix.S17, n int) {
	ap := unsafe.Pointer(&a[0])
	bp := unsafe.Pointer(&b[0])
	cp := unsafe.Pointer(&c[0])
	outp := unsafe.Pointer(&out[0])
	for i := 0; i < n; i++ {
		x := (*fix.S17)(unsafe.Add(ap, i))
		y := (*fix.S17)(unsafe.Add(bp, i))
		z := (*fix.S17)(unsafe.Add(cp, i))
		o := (*fix.S17)(unsafe.Add(outp, i))
		*o = x.SMul(*y).SAdd(*z)
	}
}

func vmas17_wg(a, b, c, out []fix.S17, n int) {
	const batch = 256
	if n < batch {
		vmas17_simple(a, b, c, out, n)
		return
	}
	var wg sync.WaitGroup
	for i := 0; i < n; i += batch {
		i := i
		wg.Add(1)
		go func() {
			num := min(n-i, batch)
			vmas17_simple(a[i:], b[i:], c[i:], out[i:], num)
			wg.Done()
		}()
	}
	wg.Wait()
}

func _vmas17_serial(a, b, c, out []fix.S17, n int)
