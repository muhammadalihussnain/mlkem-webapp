package mlkem

import (
	"errors"
)

// Decapsulate implements ML-KEM decapsulation (FIPS 203 §6.3, Algorithm 18,
// CPA-PKE decryption layer).
//
// It parses the private key (s vector) and ciphertext (u, v), then recovers
// the plaintext message:
//
//	v' = NTT⁻¹(sᵀ · NTT(u))
//	m  = Compress(v - v', 1)        — i.e. round to nearest {0, Q/2}
//
// The shared secret is recomputed as KDF(m || H(pk)), matching Encapsulate.
//
// pk must be supplied alongside sk so the shared secret derivation can
// include H(pk), matching the encapsulation side.
func Decapsulate(sk, ct, pk []byte, params *Params) ([]byte, error) {
	expectedSK := params.K * polyByteLen
	if len(sk) != expectedSK {
		return nil, errors.New("mlkem: invalid private key length")
	}

	uLen := params.K * N * params.Du / bitsPerByte
	vLen := N * params.Dv / bitsPerByte
	if len(ct) != uLen+vLen {
		return nil, errors.New("mlkem: invalid ciphertext length")
	}

	// Decode secret key.
	s := decodePolyVec(sk, params.K)

	// Decode ciphertext: u_compressed || v_compressed.
	uComp := decodePolyVecCompressed(ct[:uLen], params.K, params.Du)
	vComp := decodePoly(ct[uLen:], params.Dv)

	// Decompress u and v back to polynomial domain.
	u := decompressPolyVec(uComp, params.K, params.Du)
	v := decompressPoly(vComp, params.Dv)

	// NTT-transform each u[i] and s[i].
	var uHat [maxMatrixDim][N]int32
	var sHat [maxMatrixDim][N]int32
	for i := 0; i < params.K; i++ {
		uHat[i] = NTTForward(u[i])
		sHat[i] = NTTForward(s[i])
	}

	// v' = NTT⁻¹(sᵀ · u_hat)
	var vpHat [N]int32
	for j := 0; j < params.K; j++ {
		acc := pointwiseMulNTT(sHat[j], uHat[j])
		for idx := 0; idx < N; idx++ {
			vpHat[idx] = addmod(vpHat[idx], acc[idx])
		}
	}
	vp := NTTInverse(vpHat)

	// Recover message: m = Compress(v - v', 1)
	diff := PolySub(v, vp)
	m := polyToMsg(diff)

	// Recompute shared secret to match encapsulation.
	ss := encapsSharedSecret(m, pk)
	return ss, nil
}
