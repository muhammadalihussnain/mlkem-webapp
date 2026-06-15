package mlkem

import (
	"golang.org/x/crypto/sha3"
)

// G implements SHA3-512 as specified in FIPS 203
func G(seed []byte) (rho, sigma []byte) {
	hash := sha3.Sum512(seed)
	return hash[:32], hash[32:64]
}

// XOF implements SHAKE128 for matrix A sampling
func XOF(rho []byte, i, j byte) []byte {
	domainSep := make([]byte, len(rho)+2)
	copy(domainSep, rho)
	domainSep[len(rho)] = i
	domainSep[len(rho)+1] = j
	
	shake := sha3.NewShake128()
	shake.Write(domainSep)
	
	output := make([]byte, 168)
	shake.Read(output)
	return output
}

// PRF implements SHAKE256 for generating randomness for CBD
func PRF(sigma []byte, n byte) []byte {
	domainSep := make([]byte, len(sigma)+1)
	copy(domainSep, sigma)
	domainSep[len(sigma)] = n
	
	shake := sha3.NewShake256()
	shake.Write(domainSep)
	
	output := make([]byte, 128)
	shake.Read(output)
	return output
}

// KDF implements SHAKE256 for key derivation
func KDF(sharedSecret []byte) []byte {
	shake := sha3.NewShake256()
	shake.Write(sharedSecret)
	output := make([]byte, 64)
	shake.Read(output)
	return output
}
