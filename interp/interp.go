// package interp provides helpers for interpolating fixed-point 8 bit samples.
package interp

import (
	"github.com/pfcm/fxp/fix"
)

// L does linear interpolation:
//
//	   L(a, b, c) = (1-c)*a + c*b
//		= a - c*a + c*b
//		= a + c*(b-a)
//
// Typically the last form is nice because it eliminates a multiplication,
// but because we have very limited precision it's better to use the top
// form, to avoid subtractions when we might have negative inputs.
func L(a, b, c fix.S17) fix.S17 {
	return a.SAdd(-c.SMul(a)).SAdd(c.SMul(b))
}
