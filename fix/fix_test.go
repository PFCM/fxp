package fix

import (
	"testing"
)

func TestS17SAdd(t *testing.T) {
	for _, c := range []struct {
		a, b S17
		out  S17
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, -1, -1},
		{1, -1, 0},
		{-10, 15, 5},
		{125, 10, 127},
		{-126, 10, -116},
		{-125, -10, -128},
	} {
		got := c.a.SAdd(c.b)
		if got != c.out {
			t.Errorf("%s SAdd %s = %s, want: %s", c.a, c.b, got, c.out)
		}
		got = c.b.SAdd(c.a)
		if got != c.out {
			t.Errorf("%s SAdd %s = %s, want: %s", c.b, c.a, got, c.out)
		}
	}
}

func TestS17SMul(t *testing.T) {
	s44 := func(f float64) S17 {
		return FromFloat(f)
	}
	for _, c := range []struct {
		a, b S17
		out  S17
	}{
		{0, s44(1), 0},
		{0, s44(-1), 0},
		{s44(0.5), s44(0.5), s44(0.25)},
		{s44(0.5), s44(-0.5), s44(-0.25)},
		{s44(1.0), s44(0.5), s44(0.4921875)}, // 1.0 is slightly truncated
	} {
		got := c.a.SMul(c.b)
		if got != c.out {
			t.Errorf("%s SMul %s = %s, want: %s", c.a, c.b, got, c.out)
		}
		got = c.b.SMul(c.a)
		if got != c.out {
			t.Errorf("%s SMul %s = %s, want: %s", c.b, c.a, got, c.out)
		}
	}
}

func TestFromFloat(t *testing.T) {
	for _, c := range []struct {
		in  float64
		out S17
	}{
		{1.0, MaxS17},
		{2.0, MaxS17},
		{-1.0, MinS17},
		{-2.0, MinS17},
	} {
		got := FromFloat(c.in)
		if got != c.out {
			t.Errorf("FromFloat(%f): %s: want: %s", c.in, got, c.out)
		}
	}
}

func TestS17Float32RoundTrip(t *testing.T) {
	for i := int(MinS17); i <= int(MaxS17); i++ {
		s := S17(i)
		got := FromFloat(Float[float32](s))
		if s != got {
			t.Errorf("%x: Float: %f, FromFloat: %x", s, Float[float64](s), got)
		}
	}
}
