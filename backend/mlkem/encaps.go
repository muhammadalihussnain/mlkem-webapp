package mlkem

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/sha3"
)

// msgLen is the length in bytes of the plaintext message embedded in the ciphertext.
// FIPS 203 fixes this at 32 bytes (one 256-bit block).
const msgLen = 32

// sharedSecretLen is the length of the shared secret output in bytes.
const sharedSecretLen = 32

// EncapsResult holds every intermediate and final value from encapsulation,
// suitable for step-by-step visualisation.
type EncapsResult struct {
	// Message is the random 32-byte plaintext embedded in the ciphertext.
	Message []byte
	// R is the randomness vector sampled from B_eta1 (polynomial domain).
	R [maxMatrixDim][N]int32
	// E1 is the error vector for u (polynomial domain).
	E1 [maxMatrixDim][N]int32
	// E2 is the error polynomial for v (polynomial domain).
	E2 [N]int32
	// U is the first ciphertext component (compressed vector).
	U [maxMatrixDim][N]int32
	// V is the second ciphertext component (compressed scalar polynomial).
	V [N]int32
	// Ciphertext is the encoded ciphertext bytes (u_compressed || v_compressed).
	Ciphertext []byte
	// SharedSecret is the 32-byte shared secret.
	SharedSecret []byte
}

// Encapsulate implements ML-KEM encapsulation (FIPS 203 §6.2, Algorithm 17,
// CPA-PKE encryption layer).
//
// It parses the public key, generates a random message m, then computes:
//
//	(r, e1, e2) from PRF(seed)
//	u = NTT⁻¹(Aᵀ · NTT(r)) + e1
//	v = NTT⁻¹(tᵀ · NTT(r)) + e2 + Decompress(m, 1)
//
// The ciphertext is Compress(u, du) || Compress(v, dv).
// The shared secret is KDF(m || H(pk)).
func Encapsulate(pk []byte, params *Params) (*EncapsResult, error) {
	expectedPK := params.K*polyByteLen + SeedLen
	if len(pk) != expectedPK {
		return nil, errors.New("mlkem: invalid public key length")
	}

	// Parse public key: t vector || rho.
	tBytes := pk[:params.K*polyByteLen]
	rho := pk[params.K*polyByteLen:]

	t := decodePolyVec(tBytes, params.K)

	// Reconstruct the public matrix A (transpose will be used).
	AHat := GenerateMatrixA(rho, params.K)

	// Generate a random 32-byte message.
	m := make([]byte, msgLen)
	if _, err := rand.Read(m); err != nil {
		return nil, errors.New("mlkem: failed to generate random message: " + err.Error())
	}

	// Derive a deterministic seed from m to generate r, e1, e2.
	// Using H(m) as the noise seed keeps encapsulation deterministic given m.
	noiseSeed := hashH(m)

	// Sample r, e1, e2 using CBD.
	var r, e1 [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		rBytes := prfWithLen(noiseSeed, byte(i), cbdInputSize(params.Eta1))
		r[i] = CBD(rBytes, params.Eta1)
	}
	for i := 0; i < params.K; i++ {
		e1Bytes := prfWithLen(noiseSeed, byte(params.K+i), cbdInputSize(params.Eta2))
		e1[i] = CBD(e1Bytes, params.Eta2)
	}
	e2Bytes := prfWithLen(noiseSeed, byte(2*params.K), cbdInputSize(params.Eta2))
	e2 := CBD(e2Bytes, params.Eta2)

	// NTT-transform r.
	var rHat [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		rHat[i] = NTTForward(r[i])
	}

	// u = INTT(Aᵀ · r_hat) + e1
	// For the transpose, swap the A[i][j] → A[j][i] indices.
	var uHat [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		for j := 0; j < params.K; j++ {
			acc := pointwiseMulNTT(AHat[j][i], rHat[j])
			for idx := 0; idx < N; idx++ {
				uHat[i][idx] = addmod(uHat[i][idx], acc[idx])
			}
		}
	}
	var u [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		u[i] = PolyAdd(NTTInverse(uHat[i]), e1[i])
	}

	// v = INTT(tᵀ · r_hat) + e2 + Decompress(m, 1)
	var vHat [N]int32
	for j := 0; j < params.K; j++ {
		acc := pointwiseMulNTT(t[j], rHat[j])
		for idx := 0; idx < N; idx++ {
			vHat[idx] = addmod(vHat[idx], acc[idx])
		}
	}
	// tᵀ is already in NTT domain; t[j] is stored in polynomial domain → NTT first.
	// Re-do: t is decoded from pk (polynomial domain), so we need NTT(t[j]).
	var tHat [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		tHat[i] = NTTForward(t[i])
	}
	vHat = [N]int32{}
	for j := 0; j < params.K; j++ {
		acc := pointwiseMulNTT(tHat[j], rHat[j])
		for idx := 0; idx < N; idx++ {
			vHat[idx] = addmod(vHat[idx], acc[idx])
		}
	}
	mPoly := msgToPoly(m)
	v := PolyAdd(PolyAdd(NTTInverse(vHat), e2), mPoly)

	// Compress and encode.
	uComp := compressPolyVec(u, params.K, params.Du)
	vComp := compressPoly(v, params.Dv)

	ct := encodePolyVecCompressed(uComp, params.K, params.Du)
	ct = append(ct, encodePoly(vComp, params.Dv)...)

	// Shared secret: KDF(m || H(pk)).
	ss := encapsSharedSecret(m, pk)

	return &EncapsResult{
		Message:      m,
		R:            r,
		E1:           e1,
		E2:           e2,
		U:            u,
		V:            v,
		Ciphertext:   ct,
		SharedSecret: ss,
	}, nil
}

// decodePolyVec deserialises k polynomials from a byte slice using 12-bit encoding.
func decodePolyVec(data []byte, k int) [maxMatrixDim][N]int32 {
	var vec [maxMatrixDim][N]int32
	for i := 0; i < k; i++ {
		vec[i] = BytesToPoly(data[i*polyByteLen : (i+1)*polyByteLen])
	}
	return vec
}

// msgToPoly encodes a 32-byte message as a polynomial with 1-bit coefficients
// scaled to {0, Q/2}: bit 0 → 0, bit 1 → round(Q/2).
func msgToPoly(m []byte) [N]int32 {
	var poly [N]int32
	for i := 0; i < N; i++ {
		bit := int32((m[i/bitsPerByte] >> uint(i%bitsPerByte)) & 1)
		// Decompress_1: bit * round(Q/2)
		poly[i] = decompress(bit, 1)
	}
	return poly
}

// polyToMsg recovers a 32-byte message from a polynomial by compressing each
// coefficient to 1 bit: coeff → round(2/Q * coeff) mod 2.
func polyToMsg(poly [N]int32) []byte {
	m := make([]byte, msgLen)
	for i := 0; i < N; i++ {
		bit := compress(poly[i], 1)
		m[i/bitsPerByte] |= byte(bit) << uint(i%bitsPerByte)
	}
	return m
}

// hashH returns the 32-byte SHA3-256 hash of data, used in the shared secret KDF.
func hashH(data []byte) []byte {
	h := sha3.New256()
	h.Write(data)
	return h.Sum(nil)
}

// encapsSharedSecret computes the shared secret as KDF(m || H(pk)).
func encapsSharedSecret(m, pk []byte) []byte {
	hpk := hashH(pk)
	input := append(m, hpk...)
	return KDF(input)[:sharedSecretLen]
}

// ConstantTimeEqual returns true if a and b have the same length and contents,
// using a constant-time comparison to prevent timing side-channels.
func ConstantTimeEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
