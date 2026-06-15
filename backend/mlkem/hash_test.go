package mlkem

import (
	"bytes"
	"testing"
)

func TestG(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	
	rho, sigma := G(seed)
	
	if len(rho) != 32 {
		t.Errorf("rho length = %d, want 32", len(rho))
	}
	if len(sigma) != 32 {
		t.Errorf("sigma length = %d, want 32", len(sigma))
	}
	
	// Test deterministic
	rho2, sigma2 := G(seed)
	if !bytes.Equal(rho, rho2) {
		t.Error("G not deterministic")
	}
	if !bytes.Equal(sigma, sigma2) {
		t.Error("G not deterministic")
	}
	
	// Test different seed
	seed2 := make([]byte, 32)
	seed2[0] = 1
	rho3, sigma3 := G(seed2)
	if bytes.Equal(rho, rho3) {
		t.Error("Different seeds produced same rho")
	}
	if bytes.Equal(sigma, sigma3) {
		t.Error("Different seeds produced same sigma")
	}
}

func TestXOF(t *testing.T) {
	rho := make([]byte, 32)
	for i := range rho {
		rho[i] = byte(i)
	}
	
	out1 := XOF(rho, 0, 0)
	out2 := XOF(rho, 0, 0)
	
	if !bytes.Equal(out1, out2) {
		t.Error("XOF not deterministic")
	}
	
	// Test different outputs for different parameters
	out3 := XOF(rho, 0, 1)
	if bytes.Equal(out1, out3) {
		t.Error("XOF same for different j")
	}
	
	out4 := XOF(rho, 1, 0)
	if bytes.Equal(out1, out4) {
		t.Error("XOF same for different i")
	}
	
	// Test different rho
	rho2 := make([]byte, 32)
	rho2[0] = 255
	out5 := XOF(rho2, 0, 0)
	if bytes.Equal(out1, out5) {
		t.Error("XOF same for different rho")
	}
	
	if len(out1) != 168 {
		t.Errorf("XOF output length = %d, want 168", len(out1))
	}
}

func TestPRF(t *testing.T) {
	sigma := make([]byte, 32)
	for i := range sigma {
		sigma[i] = byte(i)
	}
	
	out1 := PRF(sigma, 0)
	out2 := PRF(sigma, 0)
	
	if !bytes.Equal(out1, out2) {
		t.Error("PRF not deterministic")
	}
	
	out3 := PRF(sigma, 1)
	if bytes.Equal(out1, out3) {
		t.Error("PRF same for different n")
	}
	
	// Test different sigma
	sigma2 := make([]byte, 32)
	sigma2[0] = 255
	out4 := PRF(sigma2, 0)
	if bytes.Equal(out1, out4) {
		t.Error("PRF same for different sigma")
	}
	
	if len(out1) != 128 {
		t.Errorf("PRF output length = %d, want 128", len(out1))
	}
}

func TestKDF(t *testing.T) {
	secret := []byte("test shared secret")
	
	out1 := KDF(secret)
	out2 := KDF(secret)
	
	if !bytes.Equal(out1, out2) {
		t.Error("KDF not deterministic")
	}
	
	if len(out1) != 64 {
		t.Errorf("KDF output length = %d, want 64", len(out1))
	}
	
	out3 := KDF([]byte("different secret"))
	if bytes.Equal(out1, out3) {
		t.Error("KDF same for different inputs")
	}
}

func TestHashFunctionIntegration(t *testing.T) {
	// Test that G output works with PRF
	seed := make([]byte, 32)
	rho, sigma := G(seed)
	
	// PRF should accept sigma
	prfOut := PRF(sigma, 0)
	if len(prfOut) == 0 {
		t.Error("PRF returned empty output")
	}
	
	// XOF should accept rho
	xofOut := XOF(rho, 0, 0)
	if len(xofOut) == 0 {
		t.Error("XOF returned empty output")
	}
	
	// KDF should produce output
	kdfOut := KDF(prfOut[:32])
	if len(kdfOut) != 64 {
		t.Error("KDF output length incorrect")
	}
}
