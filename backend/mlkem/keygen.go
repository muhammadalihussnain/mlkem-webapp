package mlkem

import (
	"golang.org/x/crypto/sha3"
)

// maxMatrixDim is the maximum supported module rank k (ML-KEM-1024).
// Used to size the fixed-dimension matrix and vector arrays.
const maxMatrixDim = 4

// cbdInputSize returns the number of PRF output bytes required to sample one
// polynomial from B_eta: 2 * eta * N / bitsPerByte.
func cbdInputSize(eta int) int {
	return 2 * eta * N / bitsPerByte
}

// xofSampleLen is the number of bytes drawn from the XOF per (i,j) matrix entry.
// 672 bytes (4 × 168-byte Keccak blocks) gives a rejection-sampling failure
// probability below 2^{-100} for uniform coefficients in [0, Q).
const xofSampleLen = xofBlockSize * 4

// GenerateMatrixA samples the public matrix A ∈ (Z_Q[X]/(X^N+1))^{k×k}
// from a 32-byte seed ρ (rho) using SHAKE128, as defined in FIPS 203 §4.2.1.
//
// Entry A[i][j] is produced by feeding XOF(rho, i, j) into SampleNTT.
// The result is in NTT domain (each polynomial is already transformed).
// Only the first k×k entries of the returned [4][4] array are populated.
func GenerateMatrixA(rho []byte, k int) [maxMatrixDim][maxMatrixDim][N]int32 {
	var A [maxMatrixDim][maxMatrixDim][N]int32
	for i := 0; i < k; i++ {
		for j := 0; j < k; j++ {
			stream := xofStream(rho, byte(i), byte(j), xofSampleLen)
			A[i][j] = SampleNTT(stream)
		}
	}
	return A
}

// GenerateSecretAndError samples the secret vector s and error vector e
// from the centered binomial distribution, as defined in FIPS 203 §4.2.2.
//
//   - s[i] = CBD(PRF(sigma, i),        eta1)  for i in [0, k)
//   - e[i] = CBD(PRF(sigma, k+i),      eta2)  for i in [0, k)
//
// Both s and e are returned in the polynomial ring (not NTT domain).
// Only the first k entries of each returned [4] vector are populated.
func GenerateSecretAndError(sigma []byte, k, eta1, eta2 int) (s, e [maxMatrixDim][N]int32) {
	for i := 0; i < k; i++ {
		sBytes := prfWithLen(sigma, byte(i), cbdInputSize(eta1))
		s[i] = CBD(sBytes, eta1)
	}
	for i := 0; i < k; i++ {
		eBytes := prfWithLen(sigma, byte(k+i), cbdInputSize(eta2))
		e[i] = CBD(eBytes, eta2)
	}
	return s, e
}

// prfWithLen returns n bytes from SHAKE256 keyed with (sigma || counter).
// It is the variable-length generalisation of PRF (hash.go), used internally
// to produce exactly the number of bytes CBD requires for a given eta.
func prfWithLen(sigma []byte, counter byte, n int) []byte {
	h := sha3.NewShake256()
	h.Write(sigma)
	h.Write([]byte{counter})
	out := make([]byte, n)
	h.Read(out)
	return out
}

// xofStream returns n bytes from SHAKE128 keyed with (rho || i || j).
// It is a thin wrapper around sha3.NewShake128 that abstracts stream length
// from the callers, making it easy to increase xofSampleLen without touching
// GenerateMatrixA.
func xofStream(rho []byte, i, j byte, n int) []byte {
	h := sha3.NewShake128()
	h.Write(rho)
	h.Write([]byte{i, j})
	out := make([]byte, n)
	h.Read(out)
	return out
}
