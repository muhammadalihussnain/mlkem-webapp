package mlkem

import "fmt"

// Params holds the full set of ML-KEM parameters for a given security level.
// All field values are derived from FIPS 203 Table 1 and Table 2.
type Params struct {
	// N is the polynomial degree (always 256 for ML-KEM).
	N int
	// Q is the NTT prime modulus (always 3329 for ML-KEM).
	Q int
	// K is the module rank (2, 3, or 4).
	K int
	// Eta1 is the noise width for key generation sampling.
	Eta1 int
	// Eta2 is the noise width for encapsulation sampling.
	Eta2 int
	// Delta is the rounding constant used in compression.
	Delta int
	// Du is the bit-width for compressing ciphertext vector u.
	Du int
	// Dv is the bit-width for compressing ciphertext scalar v.
	Dv int
	// PkSize is the public key size in bytes.
	PkSize int
	// SkSize is the secret key size in bytes.
	SkSize int
	// CtSize is the ciphertext size in bytes.
	CtSize int
}

// NewParams returns the Params for the given ML-KEM flavor (Flavor512, Flavor768,
// or Flavor1024). Returns an error for unrecognised flavors.
func NewParams(flavor string) (*Params, error) {
	switch flavor {
	case Flavor512:
		return &Params{
			N: N, Q: int(Q), K: k512,
			Eta1: eta1_512, Eta2: eta2, Delta: delta,
			Du: du512, Dv: dv512,
			PkSize: pkSize512, SkSize: skSize512, CtSize: ctSize512,
		}, nil

	case Flavor768:
		return &Params{
			N: N, Q: int(Q), K: k768,
			Eta1: eta1_768, Eta2: eta2, Delta: delta,
			Du: du768, Dv: dv768,
			PkSize: pkSize768, SkSize: skSize768, CtSize: ctSize768,
		}, nil

	case Flavor1024:
		return &Params{
			N: N, Q: int(Q), K: k1024,
			Eta1: eta1_1024, Eta2: eta2, Delta: delta,
			Du: du1024, Dv: dv1024,
			PkSize: pkSize1024, SkSize: skSize1024, CtSize: ctSize1024,
		}, nil

	default:
		return nil, fmt.Errorf("invalid ML-KEM flavor %q: expected one of %q, %q, %q",
			flavor, Flavor512, Flavor768, Flavor1024)
	}
}
