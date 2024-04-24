package flo

import (
	"sort"
	"testing"
)

func TestUF35ToFloat(t *testing.T) {
	type pair struct {
		u UF35
		f float64
	}
	pairs := make([]pair, 256)
	for i := 0; i < 256; i++ {
		pairs[i] = pair{
			u: UF35(i),
			f: UF35ToFloat[float64](UF35(i)),
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].f != pairs[j].f {
			return pairs[i].f < pairs[j].f
		}
		return pairs[i].u < pairs[j].u
	})
	const split, bias = 5, 3
	for _, p := range pairs {
		mask := UF35(0xff >> (8 - split))
		s := float64(p.u&mask)/float64(1<<split) + 0.125
		e := int64(p.u>>split) - bias
		t.Errorf("%02x: %f (%f x 2^%d)", p.u, p.f, s, e)
	}
}
