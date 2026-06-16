package mlkem

// PolyAdd adds two polynomials element-wise
func PolyAdd(a, b [256]int32) [256]int32 {
    var result [256]int32
    for i := 0; i < 256; i++ {
        result[i] = (a[i] + b[i]) % Q
    }
    return result
}

// PolySub subtracts polynomial b from a element-wise
func PolySub(a, b [256]int32) [256]int32 {
    var result [256]int32
    for i := 0; i < 256; i++ {
        result[i] = (a[i] - b[i] + Q) % Q
    }
    return result
}

// MatVecMul multiplies a k×k matrix of polynomials with a k-vector of polynomials
func MatVecMul(A [4][4][256]int32, s [4][256]int32, k int) [4][256]int32 {
    var result [4][256]int32

    for i := 0; i < k; i++ {
        for j := 0; j < 256; j++ {
            result[i][j] = 0
        }

        for j := 0; j < k; j++ {
            prod := PolyMul(A[i][j], s[j])
            for t := 0; t < 256; t++ {
                result[i][t] = (result[i][t] + prod[t]) % Q
            }
        }
    }

    return result
}

// PolyToBytes converts a polynomial to byte array (compressed 12-bit coefficients)
func PolyToBytes(poly [256]int32) []byte {
    bytes := make([]byte, 384)
    for i := 0; i < 256; i++ {
        if i%2 == 0 {
            bytes[3*(i/2)] = byte(poly[i] & 0xFF)
            bytes[3*(i/2)+1] = byte((poly[i] >> 8) & 0x0F)
        } else {
            bytes[3*(i/2)+1] |= byte((poly[i] & 0x0F) << 4)
            bytes[3*(i/2)+2] = byte((poly[i] >> 4) & 0xFF)
        }
    }
    return bytes
}

// BytesToPoly converts a byte array to polynomial (12-bit coefficients)
func BytesToPoly(data []byte) [256]int32 {
    var poly [256]int32
    for i := 0; i < 256; i++ {
        if i%2 == 0 {
            poly[i] = int32(data[3*(i/2)]) | (int32(data[3*(i/2)+1]&0x0F) << 8)
        } else {
            poly[i] = int32(data[3*(i/2)+1]&0xF0) >> 4
            poly[i] |= int32(data[3*(i/2)+2]) << 4
        }
    }
    return poly
}