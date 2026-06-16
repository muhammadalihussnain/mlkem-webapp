package mlkem

// compress reduces a coefficient x ∈ [0, Q) to d bits using the formula:
//
//	Compress_d(x) = round(2^d / Q * x) mod 2^d
//
// as defined in FIPS 203 §4.2.1.
func compress(x int32, d int) int32 {
	// Multiply by 2^d, add Q/2 for rounding, divide by Q, mask to d bits.
	shifted := (int64(x) << d) + int64(Q)/2
	return int32(shifted/int64(Q)) & ((1 << d) - 1)
}

// decompress expands a d-bit value y back to [0, Q) using:
//
//	Decompress_d(y) = round(Q / 2^d * y)
//
// as defined in FIPS 203 §4.2.1.
func decompress(y int32, d int) int32 {
	// Multiply by Q, add 2^(d-1) for rounding, divide by 2^d.
	return int32((int64(y)*int64(Q) + (1 << (d - 1))) >> d)
}

// compressPoly applies Compress_d to every coefficient of poly.
func compressPoly(poly [N]int32, d int) [N]int32 {
	var out [N]int32
	for i, c := range poly {
		out[i] = compress(c, d)
	}
	return out
}

// decompressPoly applies Decompress_d to every coefficient of poly.
func decompressPoly(poly [N]int32, d int) [N]int32 {
	var out [N]int32
	for i, c := range poly {
		out[i] = decompress(c, d)
	}
	return out
}

// compressPolyVec applies Compress_d to every polynomial in the first k entries of vec.
func compressPolyVec(vec [maxMatrixDim][N]int32, k, d int) [maxMatrixDim][N]int32 {
	var out [maxMatrixDim][N]int32
	for i := 0; i < k; i++ {
		out[i] = compressPoly(vec[i], d)
	}
	return out
}

// decompressPolyVec applies Decompress_d to every polynomial in the first k entries of vec.
func decompressPolyVec(vec [maxMatrixDim][N]int32, k, d int) [maxMatrixDim][N]int32 {
	var out [maxMatrixDim][N]int32
	for i := 0; i < k; i++ {
		out[i] = decompressPoly(vec[i], d)
	}
	return out
}

// encodePoly packs a polynomial compressed to d bits into bytes.
// Each coefficient uses exactly d bits; the output is (N*d/8) bytes.
func encodePoly(poly [N]int32, d int) []byte {
	out := make([]byte, N*d/bitsPerByte)
	bitPos := 0
	for _, c := range poly {
		for b := 0; b < d; b++ {
			if (c>>b)&1 == 1 {
				out[bitPos/bitsPerByte] |= 1 << (uint(bitPos) % bitsPerByte)
			}
			bitPos++
		}
	}
	return out
}

// decodePoly unpacks a byte slice into a polynomial with d-bit coefficients.
func decodePoly(data []byte, d int) [N]int32 {
	var poly [N]int32
	bitPos := 0
	for i := range poly {
		var val int32
		for b := 0; b < d; b++ {
			bit := int32(data[bitPos/bitsPerByte]>>(uint(bitPos)%bitsPerByte)) & 1
			val |= bit << b
			bitPos++
		}
		poly[i] = val
	}
	return poly
}

// encodePolyVecCompressed packs k polynomials each compressed to d bits.
func encodePolyVecCompressed(vec [maxMatrixDim][N]int32, k, d int) []byte {
	out := make([]byte, 0, k*N*d/bitsPerByte)
	for i := 0; i < k; i++ {
		out = append(out, encodePoly(vec[i], d)...)
	}
	return out
}

// decodePolyVecCompressed unpacks k polynomials each with d-bit coefficients.
func decodePolyVecCompressed(data []byte, k, d int) [maxMatrixDim][N]int32 {
	var vec [maxMatrixDim][N]int32
	stride := N * d / bitsPerByte
	for i := 0; i < k; i++ {
		vec[i] = decodePoly(data[i*stride:(i+1)*stride], d)
	}
	return vec
}
