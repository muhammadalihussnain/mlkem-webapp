package mlkem

import (
	"bytes"
	"testing"
)

// testSeed is a fixed 32-byte seed used for deterministic key generation in tests.
var testSeed = func() []byte {
	s := make([]byte, SeedLen)
	for i := range s {
		s[i] = byte(i + 1)
	}
	return s
}()

// TestEncapsDecapsRoundTrip verifies that Encapsulate followed by Decapsulate
// produces the same shared secret for all three ML-KEM parameter sets.
func TestEncapsDecapsRoundTrip(t *testing.T) {
	for _, flavor := range []string{Flavor512, Flavor768, Flavor1024} {
		t.Run(flavor, func(t *testing.T) {
			params, err := NewParams(flavor)
			if err != nil {
				t.Fatalf("NewParams: %v", err)
			}

			kg, err := KeyGen(params, testSeed)
			if err != nil {
				t.Fatalf("KeyGen: %v", err)
			}

			enc, err := Encapsulate(kg.PublicKey, params)
			if err != nil {
				t.Fatalf("Encapsulate: %v", err)
			}

			ss, err := Decapsulate(kg.PrivateKey, enc.Ciphertext, kg.PublicKey, params)
			if err != nil {
				t.Fatalf("Decapsulate: %v", err)
			}

			if !bytes.Equal(enc.SharedSecret, ss) {
				t.Errorf("shared secret mismatch:\n  encaps: %x\n  decaps: %x",
					enc.SharedSecret, ss)
			}
		})
	}
}

// TestEncapsCiphertextSize verifies that the ciphertext length matches
// the FIPS 203 specification for each parameter set.
func TestEncapsCiphertextSize(t *testing.T) {
	cases := []struct {
		flavor string
		ctSize int
	}{
		{Flavor512, ctSize512},
		{Flavor768, ctSize768},
		{Flavor1024, ctSize1024},
	}

	for _, tc := range cases {
		t.Run(tc.flavor, func(t *testing.T) {
			params, _ := NewParams(tc.flavor)
			kg, err := KeyGen(params, testSeed)
			if err != nil {
				t.Fatalf("KeyGen: %v", err)
			}

			enc, err := Encapsulate(kg.PublicKey, params)
			if err != nil {
				t.Fatalf("Encapsulate: %v", err)
			}

			if len(enc.Ciphertext) != tc.ctSize {
				t.Errorf("ciphertext size: want %d got %d", tc.ctSize, len(enc.Ciphertext))
			}
		})
	}
}

// TestEncapsSharedSecretLen verifies the shared secret is always sharedSecretLen bytes.
func TestEncapsSharedSecretLen(t *testing.T) {
	params, _ := NewParams(Flavor512)
	kg, _ := KeyGen(params, testSeed)
	enc, err := Encapsulate(kg.PublicKey, params)
	if err != nil {
		t.Fatalf("Encapsulate: %v", err)
	}
	if len(enc.SharedSecret) != sharedSecretLen {
		t.Errorf("shared secret length: want %d got %d", sharedSecretLen, len(enc.SharedSecret))
	}
}

// TestDecapsWrongKeyGivesDifferentSecret verifies that decapsulating with a
// different private key yields a different shared secret.
func TestDecapsWrongKeyGivesDifferentSecret(t *testing.T) {
	params, _ := NewParams(Flavor512)

	// Correct key pair.
	seed1 := make([]byte, SeedLen)
	for i := range seed1 {
		seed1[i] = byte(i + 1)
	}
	kg1, _ := KeyGen(params, seed1)

	// Wrong key pair.
	seed2 := make([]byte, SeedLen)
	for i := range seed2 {
		seed2[i] = byte(i + 17) // different seed
	}
	kg2, _ := KeyGen(params, seed2)

	enc, err := Encapsulate(kg1.PublicKey, params)
	if err != nil {
		t.Fatalf("Encapsulate: %v", err)
	}

	// Decapsulate with the wrong secret key (but correct ciphertext length).
	ssWrong, err := Decapsulate(kg2.PrivateKey, enc.Ciphertext, kg1.PublicKey, params)
	if err != nil {
		t.Fatalf("Decapsulate with wrong key: %v", err)
	}

	if bytes.Equal(enc.SharedSecret, ssWrong) {
		t.Error("wrong private key produced the same shared secret — implementation is broken")
	}
}

// TestEncapsInvalidPublicKey verifies that a malformed public key returns an error.
func TestEncapsInvalidPublicKey(t *testing.T) {
	params, _ := NewParams(Flavor512)
	_, err := Encapsulate([]byte("too short"), params)
	if err == nil {
		t.Error("expected error for invalid public key length, got nil")
	}
}

// TestDecapsInvalidSKLength verifies that a wrong-length secret key returns an error.
func TestDecapsInvalidSKLength(t *testing.T) {
	params, _ := NewParams(Flavor512)
	kg, _ := KeyGen(params, testSeed)
	enc, _ := Encapsulate(kg.PublicKey, params)

	_, err := Decapsulate([]byte("short"), enc.Ciphertext, kg.PublicKey, params)
	if err == nil {
		t.Error("expected error for invalid sk length, got nil")
	}
}

// TestDecapsInvalidCiphertextLength verifies that a wrong-length ciphertext returns an error.
func TestDecapsInvalidCiphertextLength(t *testing.T) {
	params, _ := NewParams(Flavor512)
	kg, _ := KeyGen(params, testSeed)

	_, err := Decapsulate(kg.PrivateKey, []byte("short"), kg.PublicKey, params)
	if err == nil {
		t.Error("expected error for invalid ciphertext length, got nil")
	}
}

// TestEncapsDeterministicWithSameMessage verifies that two encapsulations
// with the same public key produce different ciphertexts (random message).
func TestEncapsDifferentEachCall(t *testing.T) {
	params, _ := NewParams(Flavor512)
	kg, _ := KeyGen(params, testSeed)

	enc1, _ := Encapsulate(kg.PublicKey, params)
	enc2, _ := Encapsulate(kg.PublicKey, params)

	// Two calls should almost certainly produce different ciphertexts.
	if bytes.Equal(enc1.Ciphertext, enc2.Ciphertext) {
		t.Error("two Encapsulate calls produced identical ciphertexts — RNG may be broken")
	}
}

// TestMsgPolyRoundTrip verifies that msgToPoly → polyToMsg is the identity
// on random 32-byte messages.
func TestMsgPolyRoundTrip(t *testing.T) {
	msg := make([]byte, msgLen)
	for i := range msg {
		msg[i] = byte(i * 7)
	}
	poly := msgToPoly(msg)
	recovered := polyToMsg(poly)
	if !bytes.Equal(msg, recovered) {
		t.Errorf("msg round-trip failed:\n  want: %x\n   got: %x", msg, recovered)
	}
}

// TestConstantTimeEqual verifies the exported comparison helper.
func TestConstantTimeEqual(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2, 3}
	c := []byte{1, 2, 4}

	if !ConstantTimeEqual(a, b) {
		t.Error("expected equal slices to compare equal")
	}
	if ConstantTimeEqual(a, c) {
		t.Error("expected different slices to compare unequal")
	}
	if ConstantTimeEqual(a, []byte{1, 2}) {
		t.Error("expected different-length slices to compare unequal")
	}
}

// TestCompressDecompressRoundTrip verifies that decompress(compress(x, d), d) ≈ x
// within the expected rounding error for representative bit-widths.
func TestCompressDecompressRoundTrip(t *testing.T) {
	for _, d := range []int{1, 4, 10, 11} {
		maxErr := int32(Q / (1 << (d + 1)))
		for x := int32(0); x < Q; x += 13 { // sample every 13th value for speed
			c := compress(x, d)
			r := decompress(c, d)
			diff := r - x
			if diff < 0 {
				diff = -diff
			}
			// Wrap-around case.
			if diff > Q/2 {
				diff = Q - diff
			}
			if diff > maxErr+1 { // +1 for integer rounding
				t.Errorf("d=%d x=%d: compress→decompress error %d exceeds bound %d",
					d, x, diff, maxErr)
			}
		}
	}
}
