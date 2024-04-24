// package buffer provides some audio buffer primitives.
package buffer

import (
	"fmt"
	"sync"

	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/interp"
)

// Ring is an interpolating ring buffer.
// TODO: figure out an api for a multi-tap
type Ring struct {
	Buf    []fix.S17
	Writep float32
	Readp  float32
}

// newRing allocates a new ring buffer with the given number of samples. The
// write head and the read head start off in the same position to provide
// maximum delay; so be sure to read before writing. Or move the heads around,
// I'm not your mum.
func NewRing(size int) *Ring {
	return &Ring{
		Buf: make([]fix.S17, size),
	}
}

// Write writes a chunk of samples to the buffer at the current position of the
// write head using the provided rate of input samples to buffer
// samples. Updates the write head.
func (r *Ring) Write(in, rates []fix.S17) {
	// rate := rates[0]
	// all := true
	// for _, s := range rates[1:] {
	// 	if s != rate {
	// 		all = false
	// 		break
	// 	}
	// }
	// if all {
	// 	r.writeConstantRate(in, fix.InterpretAsRat44(rate))
	// 	return
	// }
	// TODO: this does not feel good.
	for i, rate := range rates {
		// rate = (rate << 4) | (rate >> 4)
		step := fix.InterpretAsRat44(rate)
		nextWritep := r.Writep + step
		newSamples := int(nextWritep) - int(r.Writep)
		switch newSamples {
		case 0:
			// nothing to do.
		// case 1:
		// 	// interpolate a new sample based on the two on either
		// 	// side of i.
		// 	c := (nextWritep - float32(int(nextWritep))) / step
		// 	a := in[i]
		// 	if i > 0 {
		// 		a = interp.L(in[i], in[i-1], fix.FromFloat(c))
		// 	}
		// 	r.Buf[int(nextWritep)%len(r.Buf)] = a
		default:
			// > 1, rate is small so we have to write multiple buffer
			// samples per input sample.
			for wp := int(r.Writep) + 1; float32(wp) <= nextWritep; wp++ {
				c := (float32(wp) - r.Writep) / step
				a := in[i]
				if i > 0 {
					a = interp.L(in[i], in[i-1], fix.FromFloat(c))
				}
				r.Buf[wp%len(r.Buf)] = a
			}
		}
		r.Writep = nextWritep
		if n := float32(len(r.Buf)); r.Writep >= n {
			r.Writep -= n
		}
	}
}

func (r *Ring) writeConstantRate(in []fix.S17, rate float32) {
	src := resample(in, rate)
	defer putB(src)
	if len(src) > len(r.Buf) {
		panic(fmt.Errorf("input %d larger than Buffer %d", len(src), len(r.Buf)))
	}
	wp := int(r.Writep) % len(r.Buf) // definitely not correct
	copied := copy(r.Buf[wp:], src)
	if copied < len(src) {
		// we couldn't fit it all on the end.
		wp = copy(r.Buf, src[copied:])
	} else {
		wp += copied
	}
	r.Writep = float32(wp)
}

// Read reads a chunk of samples from the buffer at the current read
// head. Advances the read head. The slice of rates is reinterpreted
// as fix.Rat44s.
func (r *Ring) Read(out, rates []fix.S17) {
	if len(out) > len(r.Buf) {
		panic(fmt.Errorf("output %d larger than buffer %d", len(out), len(r.Buf)))
	}
	for i := range out {
		out[i] = readAt(r.Buf, r.Readp)
		r.Readp += fix.InterpretAsRat44(rates[i])
		if n := float32(len(r.Buf)); r.Readp >= n {
			r.Readp -= n
		}
	}
}

func readAt(src []fix.S17, pos float32) fix.S17 {
	j, k := int(pos), int(pos+1)%len(src)
	c := fix.FromFloat(pos - float32(j))
	return interp.L(src[j%len(src)], src[k], c)
}

var pool = sync.Pool{
	New: func() any {
		b := make([]fix.S17, 4096)
		return &b
	},
}

func getB(size int) []fix.S17 {
	b := *(pool.Get().(*[]fix.S17))
	if cap(b) < size {
		b = make([]fix.S17, size)
	}
	return b[:size]
}

func putB(b []fix.S17) {
	pool.Put(&b)
}

// TODO: probably needs a slightly fractional start position, unlikely to really matter.
func resample(src []fix.S17, rate float32) []fix.S17 {
	var (
		size  = int(float32(len(src)) * rate)
		inc   = 1.0 / rate
		b     = getB(size)
		phase = float32(0)
	)
	for i := range b {
		// TODO: we should be wrapping here
		// but what else do we do at the end?
		b[i] = readAt(src, phase)
		phase += inc
	}
	return b
}
