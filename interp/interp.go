// package interp provides helpers for interpolating fixed-point 8 bit samples.
package interp

import "github.com/pfcm/fxp/fix"

// L does linear interpolation: L(a, b, c) = c*a + (1-c)*b
//
//	= c*a + b-c*b
//	= c*(a-b)+b
func L(a, b, c fix.S17) fix.S17 {
	d := a.SAdd(-b)
	return d.SMul(c).SAdd(b)
}
