// Code generated by by github.com/pfcm/fxp/fix/gen DO NOT EDIT.

package fix

import (
	"math/rand"
	"testing"
)

func TestS26FloatRoundTrip(t *testing.T) {
	for i := -128; i < 128; i++ {
		u := S26(i)
		f := S26ToFloat[float64](u)
		u2 := S26FromFloat(f)
		if u != u2 {
			t.Errorf("FromFloat(ToFloat(%x)) = %x", u, u2)
		}
	}
}

func TestS26Ops(t *testing.T) {
	// The actual implementations of these are tested pretty thoroughly
	// so all we need to do here it make sure that the generated code
	// appears to be using them correctly.
	var as, bs [256]S26
	for i := range as {
		as[i] = S26(i)
		bs[i] = S26(i)
	}
	shuffle := func() {
		rand.Shuffle(256, func(i, j int) { as[i], as[j] = as[j], as[i] })
		rand.Shuffle(256, func(i, j int) { bs[i], bs[j] = bs[j], bs[i] })
	}
	test := func(t *testing.T, name string, f func(S26, S26) S26, h func(float64, float64) float64) {
		t.Helper()
		for range 10 {
			shuffle()
			for i, a := range as {
				b := bs[i]
				af := S26ToFloat[float64](a)
				bf := S26ToFloat[float64](b)
				wantf := h(af, bf)
				want := S26FromFloat(wantf)
				got := f(a, b)
				// We don't necessarily expect them to be
				// identical just because of rounding etc.,
				// it's ok if it's only off by one step.
				d := max(got, want) - min(got, want)
				if d > 1 {
					t.Errorf("%v.%s(%v) (%x.%s(%x)) = %v (%x), want: %v (%x)", a, name, b, uint8(a), name, uint8(b), got, uint8(got), want, uint8(want))
				}
			}
		}
	}
	t.Run("SAdd", func(t *testing.T) {
		test(t, "SAdd", S26.SAdd, func(a, b float64) float64 {
			x := a + b
			if x < -2 {
				return -2
			}
			if x > 1.984375 {
				return 1.984375
			}
			return x
		})
	})
	t.Run("SSub", func(t *testing.T) {
		test(t, "SSub", S26.SSub, func(a, b float64) float64 {
			x := a - b
			if x < -2 {
				return -2
			}
			if x > 1.984375 {
				return 1.984375
			}
			return x
		})
	})
	t.Run("SMul", func(t *testing.T) {
		test(t, "SMul", S26.SMul, func(a, b float64) float64 {
			x := a * b
			if x < -2 {
				return -2
			}
			if x > 1.984375 {
				return 1.984375
			}
			return x
		})
	})
}
