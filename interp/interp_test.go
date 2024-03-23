package interp

import (
	"testing"

	"github.com/pfcm/fxp/fix"
)

func TestL(t *testing.T) {
	f := func(f float32) fix.S17 {
		return fix.FromFloat(f)
	}
	for _, c := range []struct {
		a, b, c fix.S17
		out     fix.S17
	}{{
		a:   f(0.5),
		b:   f(0),
		c:   f(1.0),
		out: f(0.4921875),
	}, {
		a:   f(0.5),
		b:   f(-0.5),
		c:   f(0.5),
		out: f(0),
	}} {
		got := L(c.a, c.b, c.c)
		if got != c.out {
			t.Errorf("L(%v, %v, %v) = %v, want: %v", c.a, c.b, c.c, got, c.out)
		}
	}

}
