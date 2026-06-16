package mlkem

// nttSize is the number of zeta values (N/2) used in the NTT butterfly network.
const nttSize = N / 2

// nttBaseMulOffset is the index into Zetas where the base-multiplication zetas begin.
// After 7 NTT layers the last "virtual" layer starts at Zetas[nttBaseMulOffset].
const nttBaseMulOffset = nttSize / 2

// invN128 is the modular inverse of 128 mod Q: 128 * 3303 ≡ 1 (mod 3329).
// Applied to every coefficient at the end of NTTInverse to normalise the transform.
const invN128 int32 = 3303

// Zetas contains precomputed twiddle factors for the NTT, defined in FIPS 203 §4.3:
//
//	Zetas[k] = 17^bitrev7(k) mod 3329
//
// 17 is a primitive 256th root of unity modulo Q = 3329.
var Zetas = [nttSize]int32{
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

// NTTForward computes the forward Number Theoretic Transform of poly,
// as specified in FIPS 203 §4.3 Algorithm 9 (NTT).
// All arithmetic is performed modulo Q; the result is in NTT domain.
func NTTForward(poly [N]int32) [N]int32 {
	f := poly
	k := 1
	for length := nttSize; length >= 2; length >>= 1 {
		for start := 0; start < N; start += 2 * length {
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

// NTTInverse computes the inverse NTT of poly,
// as specified in FIPS 203 §4.3 Algorithm 10 (NTT^{-1}).
// Applies the normalisation factor invN128 = 128^{-1} mod Q to each coefficient.
func NTTInverse(poly [N]int32) [N]int32 {
	f := poly
	k := nttSize - 1
	for length := 2; length <= nttSize; length <<= 1 {
		for start := 0; start < N; start += 2 * length {
			zeta := Zetas[k]
			k--
			for j := start; j < start+length; j++ {
				t := f[j]
				f[j] = addmod(t, f[j+length])
				f[j+length] = mulmod(zeta, submod(f[j+length], t))
			}
		}
	}
	for i := range f {
		f[i] = mulmod(f[i], invN128)
	}
	return f
}

// BaseMul multiplies two degree-1 polynomials in Z_Q[X]/(X^2 - zeta).
//
// Each polynomial is represented as a pair of coefficients:
//
//	a = (a[2i], a[2i+1]),  b = (b[2i], b[2i+1])
//
// The result satisfies:
//
//	c[2i]   = a0*b0 + zeta*a1*b1  (mod Q)
//	c[2i+1] = a0*b1 + a1*b0       (mod Q)
func BaseMul(a, b [N]int32, zeta int32) [N]int32 {
	var r [N]int32
	for i := 0; i < nttSize; i++ {
		a0, a1 := a[2*i], a[2*i+1]
		b0, b1 := b[2*i], b[2*i+1]
		r[2*i] = addmod(mulmod(a0, b0), mulmod(mulmod(zeta, a1), b1))
		r[2*i+1] = addmod(mulmod(a0, b1), mulmod(a1, b0))
	}
	return r
}

// PolyMul multiplies two polynomials in Z_Q[X]/(X^256 + 1) using NTT-based multiplication:
//  1. Forward NTT both operands.
//  2. Pointwise multiply each of the 128 coefficient-pairs using BaseMul.
//  3. Inverse NTT the result.
//
// The zeta for pair i is Zetas[nttBaseMulOffset + i/2], negated for odd i,
// reflecting the structure of the 8th (virtual) NTT layer.
func PolyMul(a, b [N]int32) [N]int32 {
	aHat := NTTForward(a)
	bHat := NTTForward(b)

	var cHat [N]int32
	for i := 0; i < nttSize; i++ {
		zeta := Zetas[nttBaseMulOffset+i/2]
		if i%2 == 1 {
			// Odd pairs use the negated zeta (conjugate factor in the ring).
			zeta = Q - zeta
		}
		a0, a1 := aHat[2*i], aHat[2*i+1]
		b0, b1 := bHat[2*i], bHat[2*i+1]
		cHat[2*i] = addmod(mulmod(a0, b0), mulmod(mulmod(zeta, a1), b1))
		cHat[2*i+1] = addmod(mulmod(a0, b1), mulmod(a1, b0))
	}

	return NTTInverse(cHat)
}

// addmod returns (a + b) mod Q in [0, Q).
// Assumes a, b are already in [0, Q).
func addmod(a, b int32) int32 {
	r := a + b
	if r >= Q {
		r -= Q
	}
	return r
}

// submod returns (a - b) mod Q in [0, Q).
// Assumes a, b are already in [0, Q).
func submod(a, b int32) int32 {
	r := a - b
	if r < 0 {
		r += Q
	}
	return r
}

// mulmod returns (a * b) mod Q via Barrett reduction.
func mulmod(a, b int32) int32 {
	return barrettReduce(a * b)
}
