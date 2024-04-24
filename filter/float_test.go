package filter

import (
	"sort"
	"testing"
)

func TestUF8ToFloat(t *testing.T) {
	type pair struct {
		u uf8
		f float64
	}
	var values [256]pair
	for i := range values {
		values[i] = pair{uf8(i), uf8ToFloat[float64](uf8(i))}
	}
	sort.Slice(values[:], func(i, j int) bool {
		vi, vj := values[i], values[j]
		if vi.f != vj.f {
			return vi.f < vj.f
		}
		return vi.u < vj.u
	})
	for _, p := range values {
		t.Errorf("%2x: %f", p.u, p.f)
	}
}
