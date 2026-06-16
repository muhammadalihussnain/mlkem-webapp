package mlkem

import (
	"math"
	"math/rand"
	"testing"
)

// Test seeds and trial counts — no magic numbers inline.
const (
	seedCBD       = 1234
	seedSampleNTT = 5678

	cbdTrials       = 1000  // number of polynomials sampled to test mean ≈ 0
	sampleNTTTrials = 100   // number of SampleNTT calls checked for range

	// meanTolerance is the maximum allowed absolute deviation of the sample mean
	// from zero.  Over cbdTrials*N samples the standard error is very small;
	// 0.05 is a conservative bound.
	meanTolerance = 0.05

	// xofRhoLen is the length of a test ρ seed in bytes (matches FIPS 203 §3.3).
	xofRhoLen = 32
)

// cbdInputLen returns the required PRF output length for CBD with parameter eta:
// 2 * eta * N / bitsPerByte bytes.
func cbdInputLen(eta int) int {
	return 2 * eta * N / bitsPerByte
}

// TestCBDRangeEta2 verifies that every coefficient produced by CBD with eta=2
// lies in the signed range [-eta, +eta], represented in [0, Q) after lifting.
func TestCBDRangeEta2(t *testing.T) {
	testCBDRange(t, eta1_768) // eta=2
}

// TestCBDRangeEta3 verifies the same property for eta=3 (used in ML-KEM-512).
func TestCBDRangeEta3(t *testing.T) {
	testCBDRange(t, eta1_512) // eta=3
}

// testCBDRange is the shared helper: samples one polynomial per RNG-derived
// input and checks that every coefficient is in the valid lifted range.
func testCBDRange(t *testing.T, eta int) {
	t.Helper()
	r := rand.New(rand.NewSource(seedCBD))
	inputLen := cbdInputLen(eta)

	for trial := 0; trial < cbdTrials; trial++ {
		buf := make([]byte, inputLen)
		r.Read(buf) //nolint:errcheck // rand.Read never errors

		poly := CBD(buf, eta)
		for i, c := range poly {
			// After lifting, a coefficient from [-eta, +eta] maps to:
			//   positive values: [0, eta]
			//   negative values: [Q-eta, Q-1]
			if c < 0 || c >= int32(Q) {
				t.Fatalf("trial %d eta=%d: coeff[%d]=%d out of [0,Q)", trial, eta, i, c)
			}
			// Check that the "signed" value is within [-eta, +eta].
			signed := c
			if signed > int32(Q)/2 {
				signed -= int32(Q)
			}
			if signed < -int32(eta) || signed > int32(eta) {
				t.Fatalf("trial %d eta=%d: coeff[%d] signed value %d outside [-%d, +%d]",
					trial, eta, i, signed, eta, eta)
			}
		}
	}
}

// TestCBDMeanNearZero verifies that the sample mean of CBD coefficients
// (interpreted as signed values) is approximately zero, confirming the
// distribution is centred.
func TestCBDMeanNearZero(t *testing.T) {
	const eta = eta1_768
	r := rand.New(rand.NewSource(seedCBD))
	inputLen := cbdInputLen(eta)

	var sum float64
	total := 0

	for trial := 0; trial < cbdTrials; trial++ {
		buf := make([]byte, inputLen)
		r.Read(buf) //nolint:errcheck

		poly := CBD(buf, eta)
		for _, c := range poly {
			signed := c
			if signed > int32(Q)/2 {
				signed -= int32(Q)
			}
			sum += float64(signed)
			total++
		}
	}

	mean := sum / float64(total)
	if math.Abs(mean) > meanTolerance {
		t.Errorf("CBD mean = %.6f, want |mean| < %.2f (distribution not centred)", mean, meanTolerance)
	}
}

// TestSampleNTTRange verifies that every coefficient produced by SampleNTT
// lies in [0, Q-1] = [0, 3328].
func TestSampleNTTRange(t *testing.T) {
	r := rand.New(rand.NewSource(seedSampleNTT))

	for trial := 0; trial < sampleNTTTrials; trial++ {
		stream := make([]byte, xofBufSize)
		r.Read(stream) //nolint:errcheck

		poly := SampleNTT(stream)
		for i, c := range poly {
			if c < 0 || c >= int32(Q) {
				t.Fatalf("trial %d: SampleNTT coeff[%d]=%d outside [0, Q-1]", trial, i, c)
			}
		}
	}
}

// TestSampleNTTFullyPopulated verifies that SampleNTT always fills all N
// coefficients when given a sufficiently long stream (xofBufSize bytes of
// uniformly random data provides enough candidates with overwhelming probability).
func TestSampleNTTFullyPopulated(t *testing.T) {
	r := rand.New(rand.NewSource(seedSampleNTT))
	stream := make([]byte, xofBufSize)
	r.Read(stream) //nolint:errcheck

	poly := SampleNTT(stream)

	// A zero polynomial is astronomically unlikely from random input;
	// if all coefficients are zero the sampling loop didn't run properly.
	allZero := true
	for _, c := range poly {
		if c != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("SampleNTT returned all-zero polynomial — likely did not sample enough coefficients")
	}
}

// TestGenerateMatrixADimensions verifies that GenerateMatrixA populates exactly
// k×k entries for each supported flavor and leaves entries beyond k at zero.
func TestGenerateMatrixADimensions(t *testing.T) {
	rho := make([]byte, xofRhoLen)
	// Use a fixed non-zero seed so the matrix is non-trivial.
	for i := range rho {
		rho[i] = byte(i + 1)
	}

	for _, tc := range []struct {
		flavor string
		k      int
	}{
		{Flavor512, k512},
		{Flavor768, k768},
		{Flavor1024, k1024},
	} {
		t.Run(tc.flavor, func(t *testing.T) {
			A := GenerateMatrixA(rho, tc.k)

			// Entries within k×k must be non-trivially populated.
			for i := 0; i < tc.k; i++ {
				for j := 0; j < tc.k; j++ {
					allZero := true
					for _, c := range A[i][j] {
						if c != 0 {
							allZero = false
							break
						}
					}
					if allZero {
						t.Errorf("A[%d][%d] is all-zero for k=%d — SampleNTT may have failed",
							i, j, tc.k)
					}
				}
			}

			// Entries outside k×k must be zero (untouched).
			for i := tc.k; i < maxMatrixDim; i++ {
				for j := tc.k; j < maxMatrixDim; j++ {
					for idx, c := range A[i][j] {
						if c != 0 {
							t.Errorf("A[%d][%d][%d]=%d should be zero (outside k=%d)",
								i, j, idx, c, tc.k)
						}
					}
				}
			}
		})
	}
}

// TestGenerateSecretAndErrorRange verifies that all coefficients in the secret
// vector s and error vector e lie in the valid CBD range for their respective eta.
func TestGenerateSecretAndErrorRange(t *testing.T) {
	sigma := make([]byte, xofRhoLen)
	for i := range sigma {
		sigma[i] = byte(i + 7)
	}

	for _, tc := range []struct {
		flavor string
		k      int
		eta1   int
	}{
		{Flavor512, k512, eta1_512},
		{Flavor768, k768, eta1_768},
		{Flavor1024, k1024, eta1_1024},
	} {
		t.Run(tc.flavor, func(t *testing.T) {
			s, e := GenerateSecretAndError(sigma, tc.k, tc.eta1, eta2)

			checkVecRange := func(name string, vec [maxMatrixDim][N]int32, eta int) {
				for i := 0; i < tc.k; i++ {
					for idx, c := range vec[i] {
						if c < 0 || c >= int32(Q) {
							t.Errorf("%s[%d][%d]=%d out of [0,Q)", name, i, idx, c)
						}
						signed := c
						if signed > int32(Q)/2 {
							signed -= int32(Q)
						}
						if signed < -int32(eta) || signed > int32(eta) {
							t.Errorf("%s[%d][%d] signed=%d outside [-%d,+%d]",
								name, i, idx, signed, eta, eta)
						}
					}
				}
			}

			checkVecRange("s", s, tc.eta1)
			checkVecRange("e", e, eta2)
		})
	}
}
