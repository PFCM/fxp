// package dct implements a discrete cosine transform (DCT-II and DCT-III) on
// fix.S17s.
package dct

import (
	"math"

	"github.com/pfcm/fxp/fix"
)

// Matrix builds a DCT matrix. The result is square with size n, and each element
// [i][j] = cos(pi/n * (j + 1/2) * i).
// TODO: we almost definitely don't need the whole matrix.
func Matrix(n int) [][]fix.S17 {
	out := make([][]fix.S17, n)
	for i := range out {
		out[i] = make([]fix.S17, n)
	}
	var piOverN = math.Pi / float64(n)
	for i := range out {
		for j := range out[i] {
			f := math.Cos(piOverN * (float64(j) + 0.5) * float64(i))
			out[i][j] = fix.FromFloat(f)
		}
	}
	return out
}

// TODO: use an algorithm that isn't O(N^2)
func Transform(in, out []fix.S17, mat [][]fix.S17) {
	// This is just a matrix multiply.
	for i := range out {
		var acc fix.S17 = 0
		for j, c := range mat[i] {
			acc = acc.SAdd(c.SMul(in[j]))
		}
		out[i] = acc
	}
}
