package mlkem

import (
	"bytes"
	"testing"
)

// TestKeyGenRejectsShortSeed verifies KeyGen returns an error for seeds shorter than 32 bytes.
func TestKeyGenRejectsShortSeed(t *testing.T) {
	params, _ := NewParams(Flavor512)
	_, err := KeyGen(params, make([]byte, SeedLen-1))
	if err == nil {
		t.Fatal("expected error for short seed, got nil")
	}
}

// TestKeyGenRejectsLongSeed verifies KeyGen returns an error for seeds longer than 32 bytes.
func TestKeyGenRejectsLongSeed(t *testing.T) {
	params, _ := NewParams(Flavor512)
	_, err := KeyGen(params, make([]byte, SeedLen+1))
	if err == nil {
		t.Fatal("expected error for long seed, got nil")
	}
}

// TestKeyGenDeterministic verifies that the same seed always produces the same keys.
func TestKeyGenDeterministic(t *testing.T) {
	for _, flavor := range []string{Flavor512, Flavor768, Flavor1024} {
		t.Run(flavor, func(t *testing.T) {
			params, _ := NewParams(flavor)
			seed := make([]byte, SeedLen)
			for i := range seed {
				seed[i] = byte(i + 1)
			}

			r1, err := KeyGen(params, seed)
			if err != nil {
				t.Fatalf("KeyGen: %v", err)
			}
			r2, err := KeyGen(params, seed)
			if err != nil {
				t.Fatalf("KeyGen second: %v", err)
			}

			if !bytes.Equal(r1.PublicKey, r2.PublicKey) {
				t.Error("PublicKey not deterministic")
			}
			if !bytes.Equal(r1.PrivateKey, r2.PrivateKey) {
				t.Error("PrivateKey not deterministic")
			}
		})
	}
}

// TestKeyGenOutputSizes verifies pk and sk byte lengths match the FIPS 203 specification.
func TestKeyGenOutputSizes(t *testing.T) {
	cases := []struct {
		flavor string
		pkSize int
		skSize int
	}{
		{Flavor512, pkSize512, skSize512},
		{Flavor768, pkSize768, skSize768},
		{Flavor1024, pkSize1024, skSize1024},
	}

	seed := make([]byte, SeedLen)
	for i := range seed {
		seed[i] = byte(i + 3)
	}

	for _, tc := range cases {
		t.Run(tc.flavor, func(t *testing.T) {
			params, _ := NewParams(tc.flavor)
			r, err := KeyGen(params, seed)
			if err != nil {
				t.Fatalf("KeyGen: %v", err)
			}

			if len(r.PublicKey) != tc.pkSize {
				t.Errorf("pk size: want %d got %d", tc.pkSize, len(r.PublicKey))
			}
			if len(r.PrivateKey) != tc.skSize {
				t.Errorf("sk size: want %d got %d", tc.skSize, len(r.PrivateKey))
			}
		})
	}
}

// TestKeyGenResultFields verifies that all KeyGenResult fields are populated.
func TestKeyGenResultFields(t *testing.T) {
	params, _ := NewParams(Flavor512)
	seed := make([]byte, SeedLen)
	for i := range seed {
		seed[i] = byte(i + 7)
	}

	r, err := KeyGen(params, seed)
	if err != nil {
		t.Fatalf("KeyGen: %v", err)
	}

	if !bytes.Equal(r.Seed, seed) {
		t.Error("Seed not preserved in result")
	}
	if len(r.Rho) != SeedLen {
		t.Errorf("Rho length: want %d got %d", SeedLen, len(r.Rho))
	}
	if len(r.Sigma) != SeedLen {
		t.Errorf("Sigma length: want %d got %d", SeedLen, len(r.Sigma))
	}

	// Verify t coefficients are in [0, Q).
	for i := 0; i < params.K; i++ {
		for j, c := range r.T[i] {
			if c < 0 || c >= int32(Q) {
				t.Errorf("T[%d][%d]=%d out of [0,Q)", i, j, c)
			}
		}
	}
}

// TestEncodePolyVec verifies that EncodePolyVec produces k*polyByteLen bytes.
func TestEncodePolyVec(t *testing.T) {
	for _, k := range []int{k512, k768, k1024} {
		var vec [maxMatrixDim][N]int32
		out := EncodePolyVec(vec, k)
		want := k * polyByteLen
		if len(out) != want {
			t.Errorf("k=%d: EncodePolyVec length want %d got %d", k, want, len(out))
		}
	}
}
