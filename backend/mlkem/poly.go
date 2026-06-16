package mlkem

// polyByteLen is the number of bytes required to encode one polynomial
// with 12-bit coefficients: 256 coefficients * 12 bits / 8 = 384 bytes.
const polyByteLen = N * 12 / 8

// mask12Bit masks the lower 12 bits of a coefficient value.
const mask12Bit = 0x0FFF

// mask8Bit masks the lower 8 bits.
const mask8Bit = 0xFF

// mask4BitLow masks the lower 4 bits.
const mask4BitLow = 0x0F

// mask4BitHigh masks the upper 4 bits of a byte.
const mask4BitHigh = 0xF0

// PolyAdd returns the element-wise sum of polynomials a and b, reduced mod Q.
func PolyAdd(a, b [N]int32) [N]int32 {
	var result [N]int32
	for i := range result {
		result[i] = addmod(a[i], b[i])
	}
	return result
}

// PolySub returns the element-wise difference (a - b) mod Q.
func PolySub(a, b [N]int32) [N]int32 {
	var result [N]int32
	for i := range result {
		result[i] = submod(a[i], b[i])
	}
	return result
}

// MatVecMul computes the matrix-vector product A·s in the NTT domain,
// where A is a k×k matrix of polynomials and s is a k-vector of polynomials.
// Only the first k rows and columns of the fixed-size arrays are used.
// The result is a k-vector; entries at indices k..3 are left as zero.
func MatVecMul(A [4][4][N]int32, s [4][N]int32, k int) [4][N]int32 {
	var result [4][N]int32
	for i := 0; i < k; i++ {
		for j := 0; j < k; j++ {
			prod := PolyMul(A[i][j], s[j])
			for t := 0; t < N; t++ {
				result[i][t] = addmod(result[i][t], prod[t])
			}
		}
	}
	return result
}

// PolyToBytes serialises a polynomial into a byte slice using 12-bit packed encoding.
// Each pair of coefficients (even index 2i, odd index 2i+1) is packed into 3 bytes:
//
//	byte[3i+0] = low 8 bits of coeff[2i]
//	byte[3i+1] = high 4 bits of coeff[2i]  |  low 4 bits of coeff[2i+1] << 4
//	byte[3i+2] = high 8 bits of coeff[2i+1]
func PolyToBytes(poly [N]int32) []byte {
	out := make([]byte, polyByteLen)
	for i := 0; i < N; i += 2 {
		c0 := poly[i] & mask12Bit
		c1 := poly[i+1] & mask12Bit
		idx := 3 * (i / 2)
		out[idx] = byte(c0 & mask8Bit)
		out[idx+1] = byte((c0>>8)&mask4BitLow) | byte((c1&mask4BitLow)<<4)
		out[idx+2] = byte(c1 >> 4)
	}
	return out
}

// BytesToPoly deserialises a byte slice produced by PolyToBytes back into a polynomial.
func BytesToPoly(data []byte) [N]int32 {
	var poly [N]int32
	for i := 0; i < N; i += 2 {
		idx := 3 * (i / 2)
		poly[i] = int32(data[idx]) | (int32(data[idx+1]&mask4BitLow) << 8)
		poly[i+1] = int32(data[idx+1]&mask4BitHigh)>>4 | int32(data[idx+2])<<4
	}
	return poly
}
