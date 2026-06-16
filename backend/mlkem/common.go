// Package mlkem implements the Module Lattice-based Key Encapsulation Mechanism
// as specified in FIPS 203. This file defines all shared constants used across
// the package, centralised here to avoid any cross-file visibility issues.
package mlkem

// ── Ring parameters ────────────────────────────────────────────────────────────

// N is the degree of the polynomial ring Z_Q[X]/(X^N + 1) used in ML-KEM.
const N = 256

// Q is the NTT prime modulus satisfying Q ≡ 1 (mod 256) per FIPS 203.
const Q int32 = 3329

// ── Barrett reduction constants ────────────────────────────────────────────────

// barrettV is the Barrett reduction multiplier: floor(2^barrettShift / Q) + 1.
const barrettV = int64(20159)

// barrettShift is the right-shift used in Barrett reduction.
const barrettShift = 26

// ── Parameter set identifiers (FIPS 203 §2) ───────────────────────────────────

// Supported ML-KEM security level identifiers.
const (
	Flavor512  = "512"
	Flavor768  = "768"
	Flavor1024 = "1024"
)

// ── Module-lattice dimensions ──────────────────────────────────────────────────

// k512, k768, k1024 are the module ranks for each ML-KEM security level.
const (
	k512  = 2
	k768  = 3
	k1024 = 4
)

// ── Noise distribution widths (FIPS 203 Table 1) ──────────────────────────────

// eta1_* is the centered binomial distribution width used during key generation.
// eta2 is the width used during encapsulation and is shared across all levels.
const (
	eta1_512  = 3
	eta1_768  = 2
	eta1_1024 = 2
	eta2      = 2
)

// ── Compression parameters (FIPS 203 Table 1) ─────────────────────────────────

// du and dv are the bit-widths for compressing ciphertext components u and v.
const (
	du512  = 10
	dv512  = 4
	du768  = 10
	dv768  = 4
	du1024 = 11
	dv1024 = 5
)

// delta is the rounding constant 2^12, shared across all parameter sets.
const delta = 1 << 12

// ── Key and ciphertext sizes in bytes (FIPS 203 Table 2) ──────────────────────

const (
	pkSize512  = 800
	skSize512  = 768
	ctSize512  = 768
	pkSize768  = 1184
	skSize768  = 1152
	ctSize768  = 1088
	pkSize1024 = 1568
	skSize1024 = 1536
	ctSize1024 = 1568
)

// ── Arithmetic helpers ─────────────────────────────────────────────────────────

// barrettReduce reduces x modulo Q using Barrett reduction, returning a value
// in [0, Q). Handles both positive and negative inputs.
func barrettReduce(x int32) int32 {
	t := (int64(x) * barrettV) >> barrettShift
	x = int32(int64(x) - t*int64(Q))
	if x >= Q {
		x -= Q
	}
	if x < 0 {
		x += Q
	}
	return x
}

// mulMontgomery returns (a * b) mod Q via Barrett reduction.
func mulMontgomery(a, b int32) int32 {
	return barrettReduce(a * b)
}
