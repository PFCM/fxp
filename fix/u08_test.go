// Code generated by by github.com/pfcm/fxp/fix/gen DO NOT EDIT.

package fix

import (
	"math/rand"
	"testing"
)

func TestU08FloatRoundTrip(t *testing.T) {
	for i := 0; i < 256; i++ {
		u := U08(i)
		f := U08ToFloat[float64](u)
		u2 := U08FromFloat(f)
		if u != u2 {
			t.Errorf("FromFloat(ToFloat(%x)) = %x", u, u2)
		}
	}
}

func TestU08Ops(t *testing.T) {
	// The actual implementations of these are tested pretty thoroughly
	// so all we need to do here it make sure that the generated code
	// appears to be using them correctly.
	var as, bs [256]U08
	for i := range as {
		as[i] = U08(i)
		bs[i] = U08(i)
	}
	shuffle := func() {
		rand.Shuffle(256, func(i, j int) { as[i], as[j] = as[j], as[i] })
		rand.Shuffle(256, func(i, j int) { bs[i], bs[j] = bs[j], bs[i] })
	}
	test := func(t *testing.T, name string, f func(U08, U08) U08, h func(float64, float64) float64) {
		t.Helper()
		for range 10 {
			shuffle()
			for i, a := range as {
				b := bs[i]
				af := U08ToFloat[float64](a)
				bf := U08ToFloat[float64](b)
				wantf := h(af, bf)
				want := U08FromFloat(wantf)
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
		test(t, "SAdd", U08.SAdd, func(a, b float64) float64 {
			x := a + b
			if x < 0 {
				return 0
			}
			if x > 0.99609375 {
				return 0.99609375
			}
			return x
		})
	})
	t.Run("SSub", func(t *testing.T) {
		test(t, "SSub", U08.SSub, func(a, b float64) float64 {
			x := a - b
			if x < 0 {
				return 0
			}
			if x > 0.99609375 {
				return 0.99609375
			}
			return x
		})
	})
	t.Run("SMul", func(t *testing.T) {
		test(t, "SMul", U08.SMul, func(a, b float64) float64 {
			x := a * b
			if x < 0 {
				return 0
			}
			if x > 0.99609375 {
				return 0.99609375
			}
			return x
		})
	})
}
