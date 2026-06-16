package mlkem

// bitsPerByte is the number of bits in a byte.
const bitsPerByte = 8

// xofBlockSize is the SHAKE128 output block size (one Keccak rate block).
// Using a full block minimises XOF calls during rejection sampling.
const xofBlockSize = 168

// xofExpandFactor is the number of XOF blocks buffered by SampleNTT.
// Two blocks (336 bytes) are sufficient to fill N coefficients in the
// common case without triggering a second XOF read.
const xofExpandFactor = 2

// xofBufSize is the total byte buffer size pre-allocated for SampleNTT.
const xofBufSize = xofBlockSize * xofExpandFactor

// sampleGroupSize is the number of bytes consumed per iteration in SampleNTT.
// Each 3-byte group yields two 12-bit candidates.
const sampleGroupSize = 3

// sampleNTTMask is the 12-bit mask applied to rejection-sampling candidates.
// Values in [Q, 4095] after masking are discarded per FIPS 203 §4.2.1.
const sampleNTTMask int32 = 0x0FFF

// CBD samples a polynomial from the centered binomial distribution B_eta,
// as defined in FIPS 203 §4.2.2 Algorithm 2 (SamplePolyCBD).
//
// For each coefficient index i:
//
//	a = popcount of eta bits at bit-position 2*eta*i
//	b = popcount of eta bits at bit-position 2*eta*i + eta
//	f[i] = (a - b) mod Q  ∈ {0, 1, …, Q-eta} ∪ {Q-eta+1, …, Q-1}
//
// The signed difference (a - b) lies in [-eta, +eta]; negative values are
// lifted into [0, Q) by adding Q.
//
// bytes must have length exactly 2*eta*N/bitsPerByte.
func CBD(bytes []byte, eta int) [N]int32 {
	var poly [N]int32
	bitsPerCoeff := 2 * eta // bits consumed per output coefficient

	for i := 0; i < N; i++ {
		bitBase := i * bitsPerCoeff

		var a, b int32
		for j := 0; j < eta; j++ {
			a += int32(readBit(bytes, bitBase+j))
			b += int32(readBit(bytes, bitBase+eta+j))
		}

		diff := a - b
		if diff < 0 {
			diff += int32(Q)
		}
		poly[i] = diff
	}
	return poly
}

// readBit returns the value of the bit at position pos within data (0 or 1).
// Bits are ordered LSB-first within each byte, matching FIPS 203 bit conventions.
func readBit(data []byte, pos int) byte {
	return (data[pos/bitsPerByte] >> uint(pos%bitsPerByte)) & 1
}

// SampleNTT performs rejection sampling on a SHAKE128 byte stream to produce
// a polynomial with coefficients uniformly distributed in [0, Q),
// as defined in FIPS 203 §4.2.1 Algorithm 1 (SampleNTT).
//
// The stream is consumed in sampleGroupSize (3-byte) groups. Each group yields
// two 12-bit candidates:
//
//	c1 = (b0 | (b1 << 8)) & 0x0FFF
//	c2 = (b1 >> 4) | (b2 << 4)
//
// A candidate is accepted only if it is strictly less than Q.
//
// The caller is responsible for providing a stream long enough to produce N
// accepted coefficients. GenerateMatrixA uses an XOF that produces arbitrarily
// long output, so exhaustion does not occur in practice.
func SampleNTT(stream []byte) [N]int32 {
	var poly [N]int32
	count := 0

	for i := 0; i+sampleGroupSize <= len(stream) && count < N; i += sampleGroupSize {
		b0 := int32(stream[i])
		b1 := int32(stream[i+1])
		b2 := int32(stream[i+2])

		c1 := (b0 | (b1 << 8)) & sampleNTTMask
		c2 := (b1 >> 4) | (b2 << 4)

		if c1 < int32(Q) {
			poly[count] = c1
			count++
		}
		if count < N && c2 < int32(Q) {
			poly[count] = c2
			count++
		}
	}
	return poly
}
