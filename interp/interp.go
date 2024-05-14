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
// Typically the last form is nice because it eliminates a multiplication, but
// because we have very limited precision and saturating arithmetic it makes
// more sense to use the top form so we don't saturate prematurely.
// TODO: U08 for c?
func L(a, b, c fix.S17) fix.S17 {
	return a.SAdd(-c.SMul(a)).SAdd(c.SMul(b))
}
