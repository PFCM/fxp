package fix

// sat.go contains general 8 bit fixed-point saturating arithmetic. Most
// functions have a few implementations, the top level functions use the best
// implementation according to the benchmarks in sat_test.go. Benchmarks were
// carried out on a first-gen m1 macbook pro, obviously they may not be the best
// option on all machines. The differences are usually quite small though.

// TODO:
//  - division? this is hard and is it really necessary?

// usmul is unsigned saturating multiply on fixed-point numbers. Assumes both
// inputs have the same number of fractional bits f.
func usmul(a, b, f uint8) uint8 {
	return usmulbigmin(a, b, f)
}

// usmulbig is the easiest way to implement usmul: extend to 16 bits
// temporarily. It's usually the fastest; most machines will be storing our
// uint8s in much larger registers anyway and the conversion is pretty much
// free.
func usmulbig(a, b, f uint8) uint8 {
	x := (uint16(a) * uint16(b)) >> f
	if x&0xff00 != 0 {
		return 0xff
	}
	return uint8(x)
}

// usmulbigmin is the same as usmulbig, but using the min builtin. In early
// testing this was slower than usmulbig, but with this current iteration it
// seems to be notably faster. This is a little bit surprising, but it's
// probably safe to assume the builtin min has a better chance of being
// optimised nicely than an explicit if.
func usmulbigmin(a, b, f uint8) uint8 {
	return uint8(min(0xff, (uint16(a)*uint16(b))>>f))
}

// usmulbigbranchless is an attempt to figure out why usmulbigmin might be
// faster than usmulbig in some situations. It still branches in the sense that
// there's an if statement, but the if uses patterns the compiler is generally
// better at optimising. This is heavily architecture dependent, but on ARM64
// with go 1.22.1 this particular form tends to use instructions like CSET or
// CSEL or variants rather than a jump, and yields identical performance to
// usmulbigmin.
func usmulbigbranchless(a, b, f uint8) uint8 {
	x := (uint16(a) * uint16(b)) >> f
	if x&0xff00 != 0 {
		x = 0xff
	}
	return uint8(x)

}

// usmulbits is an awful implementation of usmul that does the multiplication by
// hand to be able to pack the result into 2 uint8s instead of a uint16. It is
// evil and excelptionally slow compared to the others.
func usmulbits(a, b, f uint8) uint8 {
	const mask uint8 = 1<<4 - 1
	// split a = (a1 * 16 + a0)
	a0 := uint8(a) & mask
	a1 := uint8(a) >> 4
	// split b = (b1 * 16 + b0)
	b0 := uint8(b) & mask
	b1 := uint8(b) >> 4

	w0 := a0 * b0
	t := a1*b0 + w0>>4
	w1 := t & mask
	w2 := t >> 4
	w1 += a0 * b1
	hi := a1*b1 + w2 + w1>>4
	lo := uint8(a) * uint8(b)

	// now we have the outcome.
	if hi>>f != 0 {
		return 0xff
	}
	// We have 16 bits, the answer we want is {hi, lo} >> f.
	x := lo >> f // the highest 8-f, in the lowest positions.
	x |= hi << (8 - f)
	return x
}

// usadd is an unsigned saturating add. In fixed-point arithmetic the scaling
// factors are unchanged after addition, so this only needs to know the
// operands.
func usadd(a, b uint8) uint8 {
	return usaddpre(a, b)
}

// checks for overflow first.
func usaddpre(a, b uint8) uint8 {
	// it's ok to only check one side, if this is true, it's also true that
	// 0xff-b < a.
	if 0xff-a < b {
		return 0xff
	}
	return a + b
}

// check for overflow after the addition.
func usaddpost(a, b uint8) uint8 {
	x := a + b
	// as before, we only need to check once.
	if x < a {
		return 0xff
	}
	return x
}

// variant of usaddpost, that generates branch-free code on both x86_64 (uses
// sethi, or setnbe outside of the go assembler) and ARM64 (uses cset). It does
// mean more work later though, and it's surprisingly touchy: replacing 1 with
// 255 in the if (and avoiding the ^) makes it insert a jump. Probably doesn't
// make much difference, but an interesting variation nevertheless.
func usaddpostbranchless(a, b uint8) uint8 {
	x := a + b
	i := uint8(0)
	if x < a {
		i = 1
	}
	return x | ^(i - 1)
}

// ussub subtracts b from a, returning zero if the result would underflow.
// Like usadd there's nothing extra required to work with fixed points.
func ussub(a, b uint8) uint8 {
	return ussubmin(a, b)
}

func ussubbranch(a, b uint8) uint8 {
	// This one is pretty straightforward to check.
	if a < b {
		return 0
	}
	return a - b
}

// could certainly still jump, but less likely. Doesn't seem to make any
// difference though.
func ussubbranchless(a, b uint8) uint8 {
	x := a - b
	i := uint8(0)
	if x > a {
		i = 1
	}
	return x & (i - 1)
}

// this one is cute.
func ussubmin(a, b uint8) uint8 {
	return a - min(a, b)
}

const (
	maxInt8 int8 = 0x7f
	minInt8 int8 = -0x80
)

// ssadd is a signed saturating add.
func ssadd(a, b int8) int8 {
	return ssaddbranchless(a, b)
}

// ssaddbranch is a simple starting point for ssadd implementations,
// checking for overflow before carrying out the addition.
func ssaddbranch(a, b int8) int8 {
	// These checks are not great.
	if a > 0 && b > 0 && a > maxInt8-b {
		return maxInt8
	}
	if a < 0 && b < 0 && a < minInt8-b {
		return minInt8
	}
	return a + b
}

func ssaddbranchless(a, b int8) int8 {
	x := a + b
	same := uint8(^(a ^ b)) >> 7 // 1 if a and b are the same sign
	s := uint8(^(x ^ a)) >> 7    // 1 if a and x are the same sign
	// maxInt8 or minInt8 depending on the sign bit of a
	r := int8(uint8(a)>>7 + 0x7f)
	//	if same != 0 && s == 0 {
	if (s^same)&same != 0 {
		x = r
	}
	return x
}

// ssaddbig just uses bigger numbers so the overflow checks are easy.
func ssaddbig(a, b int8) int8 {
	x := int16(a) + int16(b)
	return int8(max(min(x, 0x7f), -0x80))
}

// sssub is signed, saturating subtraction.
func sssub(a, b int8) int8 {
	return sssubbig(a, b)
}

// sssub by going in and out of int16s.
func sssubbig(a, b int8) int8 {
	return int8(min(max(int16(a)-int16(b), -0x80), 0x7f))
}

// sssub re-using ssadd
func sssubadd(a, b int8) int8 {
	// If b = -0x80 (-128), -b (128) is not representable. The Go spec
	// (https://go.dev/ref/spec#Operators) says -b = 0-b, which is computed
	// with overflow, meaning -(-128) = -128. This would give some confusing
	// results if we just did ssadd(a, -b), so we have to check for this
	// particular overflow first. This feels like it probably removes most
	// of the benefits of re-using ssadd (if nothing else by causing trouble
	// with the inlining), but that's what benchmarks are for.
	if b == -0x80 {
		if a >= 0 {
			// a - (-128) > 127, saturate
			return maxInt8
		}
		// Can't overflow (but 0x80 is > maxInt8)
		return a + int8(0x7f) + 1
	}
	return ssadd(a, -b)
}

// sssub re-using ssadd, but less branchy. It does indeed generate branchless
// code, but it also ends up significantly slower, I suspect it is not being
// inlined.
func sssubaddlessbranch(a, b int8) int8 {
	offset := int8(0)
	if b == -0x80 {
		// 1 if a is negative.
		offset = 1 & (a >> 7)
		b = -0x7f
	}

	return ssadd(a, -b) + offset
}

func sssubdirect(a, b int8) int8 {
	x := a - b
	// The result overflowed if:
	// - a is negative, b is positive, x is positive
	// - a is positive, b is negative, x is negative
	// if both signs are the same, overflow is impossible.
	abDifferent := 1 & ((a ^ b) >> 7)
	axDifferent := 1 & ((a ^ x) >> 7)
	if abDifferent&axDifferent != 0 {
		// overflow happened
		x = int8(uint8(a)>>7 + 0x7f)
	}
	return x
}

// ssmul is a signed, saturating fixed point multiply.
func ssmul(a, b int8, f uint8) int8 {
	return ssmulbig(a, b, f)
}

func ssmulbig(a, b int8, f uint8) int8 {
	return int8(min(max((int16(a)*int16(b))>>f, -0x80), 0x7f))
}

// ssmulbigpost is the branchiest way to implement ssmulbig, so see if thre's
// any difference.
func ssmulbigbranch(a, b int8, f uint8) int8 {
	r := int16(a) * int16(b) >> f
	if r > 0x7f {
		return 0x7f
	}
	if r < -0x80 {
		return -0x80
	}
	return int8(r)
}
