package mlkem

// Zetas contains precomputed powers of the primitive root 17 in bit-reversed order:
// Zetas[k] = 17^bitrev7(k) mod 3329
var Zetas = [128]int32{
	1, 1729, 2580, 3289, 2642, 630, 1897, 848,
	1062, 1919, 193, 797, 2786, 3260, 569, 1746,
	296, 2447, 1339, 1476, 3046, 56, 2240, 1333,
	1426, 2094, 535, 2882, 2393, 2879, 1974, 821,
	289, 331, 3253, 1756, 1197, 2304, 2277, 2055,
	650, 1977, 2513, 632, 2865, 33, 1320, 1915,
	2319, 1435, 807, 452, 1438, 2868, 1534, 2402,
	2647, 2617, 1481, 648, 2474, 3110, 1227, 910,
	17, 2761, 583, 2649, 1637, 723, 2288, 1100,
	1409, 2662, 3281, 233, 756, 2156, 3015, 3050,
	1703, 1651, 2789, 1789, 1847, 952, 1461, 2687,
	939, 2308, 2437, 2388, 733, 2337, 268, 641,
	1584, 2298, 2037, 3220, 375, 2549, 2090, 1645,
	1063, 319, 2773, 757, 2099, 561, 2466, 2594,
	2804, 1092, 403, 1026, 1143, 2150, 2775, 886,
	1722, 1212, 1874, 1029, 2110, 2935, 885, 2154,
}

// NTTForward performs the in-place forward NTT as defined in FIPS 203 §4.3.
// Coefficients are reduced modulo Q throughout.
func NTTForward(poly [256]int32) [256]int32 {
	f := poly
	k := 1
	for length := 128; length >= 2; length >>= 1 {
		for start := 0; start < 256; start += 2 * length {
			zeta := Zetas[k]
			k++
			for j := start; j < start+length; j++ {
				t := mulmod(zeta, f[j+length])
				f[j+length] = submod(f[j], t)
				f[j] = addmod(f[j], t)
			}
		}
	}
	return f
}

// NTTInverse performs the inverse NTT as defined in FIPS 203 §4.3.
// After the butterfly network, each coefficient is multiplied by 3303 = 128^-1 mod 3329.
func NTTInverse(poly [256]int32) [256]int32 {
	f := poly
	k := 127
	for length := 2; length <= 128; length <<= 1 {
		for start := 0; start < 256; start += 2 * length {
			zeta := Zetas[k]
			k--
			for j := start; j < start+length; j++ {
				t := f[j]
				f[j] = addmod(t, f[j+length])
				f[j+length] = mulmod(zeta, submod(f[j+length], t))
			}
		}
	}
	// 3303 = 128^{-1} mod 3329
	const invN int32 = 3303
	for i := range f {
		f[i] = mulmod(f[i], invN)
	}
	return f
}

// BaseMul multiplies two degree-1 polynomials in Z_Q[X]/(X^2 - zeta).
// Operands are pairs (a[2i], a[2i+1]) and (b[2i], b[2i+1]).
// Result: c0 = a0*b0 + zeta*a1*b1,  c1 = a0*b1 + a1*b0
func BaseMul(a, b [256]int32, zeta int32) [256]int32 {
	var r [256]int32
	for i := 0; i < 128; i++ {
		a0, a1 := a[2*i], a[2*i+1]
		b0, b1 := b[2*i], b[2*i+1]
		r[2*i] = addmod(mulmod(a0, b0), mulmod(mulmod(zeta, a1), b1))
		r[2*i+1] = addmod(mulmod(a0, b1), mulmod(a1, b0))
	}
	return r
}

// PolyMul multiplies two polynomials using NTT + pointwise BaseMul + INTT.
// This computes the product in Z_Q[X]/(X^256 + 1).
//
// After 7 NTT layers (length 128→2), the 128 coefficient-pairs (f[2i], f[2i+1])
// each live in Z_Q[X]/(X^2 - zeta_i).  The zeta for pair i comes from the
// "virtual" 8th layer: zeta_i = Zetas[64 + i/2] with sign controlled by parity,
// but per FIPS 203 the standard is to use Zetas[64 + i] for i in 0..63 for pairs
// 0..63, and the negated form for pairs 64..127.  In practice the simplest
// correct formulation is: pair i uses Zetas[64 + (i >> 1)] with alternating sign,
// i.e., zeta = (-1)^i * Zetas[64 + (i>>1)].
func PolyMul(a, b [256]int32) [256]int32 {
	aHat := NTTForward(a)
	bHat := NTTForward(b)

	var cHat [256]int32
	for i := 0; i < 128; i++ {
		// zeta for pair i: Zetas[64 + i>>1], negated for odd i
		zeta := Zetas[64+i/2]
		if i%2 == 1 {
			zeta = Q - zeta
		}
		a0, a1 := aHat[2*i], aHat[2*i+1]
		b0, b1 := bHat[2*i], bHat[2*i+1]
		cHat[2*i] = addmod(mulmod(a0, b0), mulmod(mulmod(zeta, a1), b1))
		cHat[2*i+1] = addmod(mulmod(a0, b1), mulmod(a1, b0))
	}

	return NTTInverse(cHat)
}

// addmod returns (a + b) mod Q, keeping the result in [0, Q).
func addmod(a, b int32) int32 {
	r := a + b
	if r >= Q {
		r -= Q
	}
	return r
}

// submod returns (a - b) mod Q, keeping the result in [0, Q).
func submod(a, b int32) int32 {
	r := a - b
	if r < 0 {
		r += Q
	}
	return r
}

// mulmod returns (a * b) mod Q using Barrett reduction.
func mulmod(a, b int32) int32 {
	return barrettReduce(a * b)
}
