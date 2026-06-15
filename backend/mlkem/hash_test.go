package mlkem

import (
	"bytes"
	"testing"
	
	"golang.org/x/crypto/sha3"
)

func TestG(t *testing.T) {
	// Test with known seed
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	
	rho, sigma := G(seed)
	
	// Check lengths
	if len(rho) != 32 {
		t.Errorf("rho length = %d, want 32", len(rho))
	}
	if len(sigma) != 32 {
		t.Errorf("sigma length = %d, want 32", len(sigma))
	}
	
	// Test deterministic output
	rho2, sigma2 := G(seed)
	if !bytes.Equal(rho, rho2) {
		t.Error("G not deterministic for same input")
	}
	if !bytes.Equal(sigma, sigma2) {
		t.Error("G not deterministic for same input")
	}
	
	// Test different seed produces different output
	seed2 := make([]byte, 32)
	seed2[0] = 1
	rho3, sigma3 := G(seed2)
	
	if bytes.Equal(rho, rho3) {
		t.Error("Different seeds produced same rho")
	}
	if bytes.Equal(sigma, sigma3) {
		t.Error("Different seeds produced same sigma")
	}
	
	// Verify total output is 64 bytes
	hash := sha3.Sum512(seed)
	if len(hash) != 64 {
		t.Errorf("SHA3-512 output length = %d, want 64", len(hash))
	}
}

func TestXOF(t *testing.T) {
	rho := make([]byte, 32)
	for i := range rho {
		rho[i] = byte(i)
	}
	
	// Test deterministic output
	output1 := XOF(rho, 0, 0)
	output2 := XOF(rho, 0, 0)
	
	if !bytes.Equal(output1, output2) {
		t.Error("XOF not deterministic for same inputs")
	}
	
	// Test domain separation works
	output3 := XOF(rho, 0, 1)
	if bytes.Equal(output1, output3) {
		t.Error("XOF same output for different j")
	}
	
	output4 := XOF(rho, 1, 0)
	if bytes.Equal(output1, output4) {
		t.Error("XOF same output for different i")
	}
	
	// Test different rho
	rho2 := make([]byte, 32)
	rho2[0] = 255
	output5 := XOF(rho2, 0, 0)
	if bytes.Equal(output1, output5) {
		t.Error("XOF same output for different rho")
	}
	
	// Check output length
	if len(output1) != 168 {
		t.Errorf("XOF output length = %d, want 168", len(output1))
	}
}

func TestPRF(t *testing.T) {
	sigma := make([]byte, 32)
	for i := range sigma {
		sigma[i] = byte(i)
	}
	
	// Test deterministic output
	output1 := PRF(sigma, 0)
	output2 := PRF(sigma, 0)
	
	if !bytes.Equal(output1, output2) {
		t.Error("PRF not deterministic for same inputs")
	}
	
	// Test domain separation with different n
	output3 := PRF(sigma, 1)
	if bytes.Equal(output1, output3) {
		t.Error("PRF same output for different n")
	}
	
	// Test different sigma
	sigma2 := make([]byte, 32)
	sigma2[0] = 255
	output4 := PRF(sigma2, 0)
	if bytes.Equal(output1, output4) {
		t.Error("PRF same output for different sigma")
	}
	
	// Check output length
	if len(output1) != 128 {
		t.Errorf("PRF output length = %d, want 128", len(output1))
	}
}

func TestKDF(t *testing.T) {
	sharedSecret := []byte("test shared secret")
	
	output1 := KDF(sharedSecret)
	output2 := KDF(sharedSecret)
	
	// Test deterministic
	if !bytes.Equal(output1, output2) {
		t.Error("KDF not deterministic")
	}
	
	// Check length (64 bytes for shared secret)
	if len(output1) != 64 {
		t.Errorf("KDF output length = %d, want 64", len(output1))
	}
	
	// Test different input gives different output
	output3 := KDF([]byte("different secret"))
	if bytes.Equal(output1, output3) {
		t.Error("KDF same output for different inputs")
	}
}

// Test that all hash functions work together
func TestHashIntegration(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	
	rho, sigma := G(seed)
	
	// Use rho with XOF
	xofOut := XOF(rho, 0, 0)
	if len(xofOut) == 0 {
		t.Error("XOF returned empty output")
	}
	
	// Use sigma with PRF
	prfOut := PRF(sigma, 0)
	if len(prfOut) == 0 {
		t.Error("PRF returned empty output")
	}
	
	// Use KDF
	kdfOut := KDF(prfOut[:32])
	if len(kdfOut) != 64 {
		t.Error("KDF output length incorrect")
	}
}
