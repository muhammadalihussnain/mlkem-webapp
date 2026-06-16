package mlkem

import (
	"errors"

	"golang.org/x/crypto/sha3"
)

// maxMatrixDim is the maximum supported module rank k (ML-KEM-1024).
// Used to size the fixed-dimension matrix and vector arrays.
const maxMatrixDim = 4

// MaxMatrixDim is the exported form of maxMatrixDim for use in other packages.
const MaxMatrixDim = maxMatrixDim

// seedLen is the required byte length of the seed passed to KeyGen.
// FIPS 203 §4.1 specifies a 32-byte seed for key generation.
const seedLen = 32

// SeedLen is the exported form of seedLen.
const SeedLen = seedLen

// xofSampleLen is the number of bytes drawn from the XOF per (i,j) matrix entry.
// 672 bytes (4 × 168-byte Keccak blocks) gives a rejection-sampling failure
// probability below 2^{-100} for uniform coefficients in [0, Q).
const xofSampleLen = xofBlockSize * 4

// KeyGenResult holds every intermediate and final value produced during ML-KEM
// key generation, suitable for step-by-step visualisation.
type KeyGenResult struct {
	// Inputs
	Seed []byte // 32-byte seed passed to G

	// Derived seeds (output of G)
	Rho   []byte // 32-byte public matrix seed
	Sigma []byte // 32-byte secret/noise seed

	// Sampled values (all in polynomial domain unless noted)
	A [maxMatrixDim][maxMatrixDim][N]int32 // public matrix, NTT domain
	S [maxMatrixDim][N]int32               // secret vector, polynomial domain
	E [maxMatrixDim][N]int32               // error vector, polynomial domain

	// Computed public value
	T [maxMatrixDim][N]int32 // t = NTT^{-1}(A_hat · NTT(s)) + e, polynomial domain

	// Encoded keys
	PublicKey  []byte // pk = encode(t) || rho
	PrivateKey []byte // sk = encode(s)
}

// KeyGen performs ML-KEM key generation for the given parameter set and 32-byte seed.
// It implements FIPS 203 §4.1 Algorithm 12 (ML-KEM.KeyGen_internal).
//
// Steps:
//  1. (rho, sigma) = G(seed)
//  2. A_hat        = GenerateMatrixA(rho, k)           — NTT domain
//  3. (s, e)       = GenerateSecretAndError(sigma, k, eta1, eta2)
//  4. s_hat        = NTT(s[i]) for each i
//  5. t            = NTT^{-1}(A_hat · s_hat) + e
//  6. pk           = encode(t) || rho
//  7. sk           = encode(s)
func KeyGen(params *Params, seed []byte) (*KeyGenResult, error) {
	if len(seed) != seedLen {
		return nil, errors.New("mlkem: seed must be exactly 32 bytes")
	}

	// Step 1 — derive public and secret seeds.
	rho, sigma := G(seed)

	// Step 2 — sample the public matrix A in NTT domain.
	A := GenerateMatrixA(rho, params.K)

	// Step 3 — sample secret s and error e from B_eta.
	s, e := GenerateSecretAndError(sigma, params.K, params.Eta1, params.Eta2)

	// Step 4 — transform s into NTT domain.
	var sHat [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		sHat[i] = NTTForward(s[i])
	}

	// Step 5 — compute t = A·s_hat in NTT domain, then inverse NTT, then add e.
	// MatVecMulNTT operates in NTT domain (both A and sHat are already transformed).
	tHat := MatVecMulNTT(A, sHat, params.K)
	var t [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		tPoly := NTTInverse(tHat[i])
		t[i] = PolyAdd(tPoly, e[i])
	}

	// Step 6 — encode public key: encode(t[0]) || … || encode(t[k-1]) || rho.
	pk := EncodePolyVec(t, params.K)
	pk = append(pk, rho...)

	// Step 7 — encode secret key: encode(s[0]) || … || encode(s[k-1]).
	sk := EncodePolyVec(s, params.K)

	return &KeyGenResult{
		Seed:       seed,
		Rho:        rho,
		Sigma:      sigma,
		A:          A,
		S:          s,
		E:          e,
		T:          t,
		PublicKey:  pk,
		PrivateKey: sk,
	}, nil
}

// MatVecMulNTT computes A·s where both A and s are already in NTT domain.
// It performs pointwise polynomial multiplication and returns the result in NTT domain.
func MatVecMulNTT(A [maxMatrixDim][maxMatrixDim][N]int32, s [maxMatrixDim][N]int32, k int) [maxMatrixDim][N]int32 {
	var result [maxMatrixDim][N]int32
	for i := 0; i < k; i++ {
		for j := 0; j < k; j++ {
			// Pointwise NTT-domain multiplication: multiply coefficient by coefficient
			// using BaseMul for each of the nttSize pairs.
			acc := pointwiseMulNTT(A[i][j], s[j])
			for idx := 0; idx < N; idx++ {
				result[i][idx] = addmod(result[i][idx], acc[idx])
			}
		}
	}
	return result
}

// pointwiseMulNTT multiplies two polynomials that are already in NTT domain
// by applying BaseMul to each of the 128 coefficient pairs.
func pointwiseMulNTT(a, b [N]int32) [N]int32 {
	var r [N]int32
	for i := 0; i < nttSize; i++ {
		zeta := Zetas[nttBaseMulOffset+i/2]
		if i%2 == 1 {
			zeta = Q - zeta
		}
		a0, a1 := a[2*i], a[2*i+1]
		b0, b1 := b[2*i], b[2*i+1]
		r[2*i] = addmod(mulmod(a0, b0), mulmod(mulmod(zeta, a1), b1))
		r[2*i+1] = addmod(mulmod(a0, b1), mulmod(a1, b0))
	}
	return r
}

// EncodePolyVec serialises the first k polynomials of vec into a byte slice
// using 12-bit packed encoding (polyByteLen bytes per polynomial).
func EncodePolyVec(vec [maxMatrixDim][N]int32, k int) []byte {
	out := make([]byte, 0, k*polyByteLen)
	for i := 0; i < k; i++ {
		out = append(out, PolyToBytes(vec[i])...)
	}
	return out
}

// cbdInputSize returns the number of PRF output bytes required to sample one
// polynomial from B_eta: 2 * eta * N / bitsPerByte.
func cbdInputSize(eta int) int {
	return 2 * eta * N / bitsPerByte
}

// GenerateMatrixA samples the public matrix A ∈ (Z_Q[X]/(X^N+1))^{k×k}
// from a 32-byte seed ρ (rho) using SHAKE128, as defined in FIPS 203 §4.2.1.
//
// Entry A[i][j] is produced by feeding XOF(rho, i, j) into SampleNTT.
// The result is in NTT domain (each polynomial is already transformed).
// Only the first k×k entries of the returned array are populated.
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
//   - s[i] = CBD(PRF(sigma, i),   eta1)  for i in [0, k)
//   - e[i] = CBD(PRF(sigma, k+i), eta2)  for i in [0, k)
//
// Both s and e are returned in the polynomial ring (not NTT domain).
// Only the first k entries of each returned vector are populated.
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
// Variable-length generalisation of PRF (hash.go) used to produce exactly
// the byte count CBD requires for a given eta.
func prfWithLen(sigma []byte, counter byte, n int) []byte {
	h := sha3.NewShake256()
	h.Write(sigma)
	h.Write([]byte{counter})
	out := make([]byte, n)
	h.Read(out)
	return out
}

// xofStream returns n bytes from SHAKE128 keyed with (rho || i || j).
func xofStream(rho []byte, i, j byte, n int) []byte {
	h := sha3.NewShake128()
	h.Write(rho)
	h.Write([]byte{i, j})
	out := make([]byte, n)
	h.Read(out)
	return out
}
