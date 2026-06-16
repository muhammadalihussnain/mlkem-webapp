package mlkem

// N is the degree of the polynomial ring Z_Q[X]/(X^N + 1) used in ML-KEM.
const N = 256

// Q is the modulus for ML-KEM, a prime satisfying Q ≡ 1 (mod 256) per FIPS 203.
const Q int32 = 3329

// barrettV is the Barrett reduction constant: floor(2^barrettShift / Q) + 1.
// Used to approximate division by Q via a multiply-and-shift.
const barrettV = int64(20159)

// barrettShift is the bit-shift used in Barrett reduction (2^26 approximation).
const barrettShift = 26

// barrettReduce reduces x modulo Q using Barrett reduction, returning a value in [0, Q).
// Handles both positive and negative inputs.
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
