package mlkem

import (
	"math/rand"
	"testing"
)

// randomPoly returns a polynomial with random coefficients in [0, Q).
func randomPoly(r *rand.Rand) [256]int32 {
	var p [256]int32
	for i := range p {
		p[i] = int32(r.Intn(int(Q)))
	}
	return p
}

// naiveNegacyclicMul computes the negacyclic convolution of a and b in Z_Q[X]/(X^256+1).
func naiveNegacyclicMul(a, b [256]int32) [256]int32 {
	var tmp [512]int32
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			tmp[i+j] = (tmp[i+j] + a[i]*b[j]) % Q
		}
	}
	// Reduce mod X^256 + 1: x^(256+k) = -x^k
	var result [256]int32
	for i := 0; i < 256; i++ {
		result[i] = (tmp[i] - tmp[i+256]%Q + Q) % Q
	}
	return result
}

// TestNTTRoundTrip verifies that NTTInverse(NTTForward(p)) == p for random polynomials.
func TestNTTRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for trial := 0; trial < 100; trial++ {
		p := randomPoly(r)
		recovered := NTTInverse(NTTForward(p))
		for i := range p {
			if p[i] != recovered[i] {
				t.Fatalf("trial %d: round-trip mismatch at index %d: want %d, got %d",
					trial, i, p[i], recovered[i])
			}
		}
	}
}

// TestPolyMulMatchesNaive verifies PolyMul agrees with naive negacyclic convolution.
func TestPolyMulMatchesNaive(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	for trial := 0; trial < 20; trial++ {
		a := randomPoly(r)
		b := randomPoly(r)
		got := PolyMul(a, b)
		want := naiveNegacyclicMul(a, b)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("trial %d: PolyMul mismatch at index %d: want %d, got %d",
					trial, i, want[i], got[i])
			}
		}
	}
}

// TestMatVecMulDimensions checks that MatVecMul returns correct dimension results for k=2,3,4.
func TestMatVecMulDimensions(t *testing.T) {
	r := rand.New(rand.NewSource(7))

	for _, k := range []int{2, 3, 4} {
		var A [4][4][256]int32
		var s [4][256]int32

		// Fill the used portion with random data
		for i := 0; i < k; i++ {
			s[i] = randomPoly(r)
			for j := 0; j < k; j++ {
				A[i][j] = randomPoly(r)
			}
		}

		result := MatVecMul(A, s, k)

		// Verify result has k non-trivially-zero rows by checking each
		// against a manual dot product of row 0.
		for i := 0; i < k; i++ {
			expected := [256]int32{}
			for j := 0; j < k; j++ {
				prod := PolyMul(A[i][j], s[j])
				for t2 := 0; t2 < 256; t2++ {
					expected[t2] = addmod(expected[t2], prod[t2])
				}
			}
			for idx := 0; idx < 256; idx++ {
				if result[i][idx] != expected[idx] {
					t.Fatalf("k=%d row=%d idx=%d: want %d got %d",
						k, i, idx, expected[idx], result[i][idx])
				}
			}
		}

		// Rows beyond k must be zero (untouched)
		for i := k; i < 4; i++ {
			for idx := 0; idx < 256; idx++ {
				if result[i][idx] != 0 {
					t.Fatalf("k=%d: row %d (beyond k) should be zero, got %d at idx %d",
						k, i, result[i][idx], idx)
				}
			}
		}
	}
}

// TestPolyAdd and TestPolySub basic correctness
func TestPolyAdd(t *testing.T) {
	var a, b [256]int32
	for i := 0; i < 256; i++ {
		a[i] = int32(i)
		b[i] = int32(256 - i)
	}
	result := PolyAdd(a, b)
	for i := 0; i < 256; i++ {
		expected := int32(256) % Q
		if result[i] != expected {
			t.Fatalf("PolyAdd[%d]: want %d got %d", i, expected, result[i])
		}
	}
}

func TestPolySub(t *testing.T) {
	var a, b [256]int32
	for i := 0; i < 256; i++ {
		a[i] = int32(i * 2)
		b[i] = int32(i)
	}
	result := PolySub(a, b)
	for i := 0; i < 256; i++ {
		if result[i] != int32(i) {
			t.Fatalf("PolySub[%d]: want %d got %d", i, i, result[i])
		}
	}
}

// TestBaseMul verifies the exported BaseMul matches inline computation.
func TestBaseMul(t *testing.T) {
	var a, b [256]int32
	a[0], a[1] = 2, 3
	b[0], b[1] = 4, 5
	zeta := int32(17)
	r := BaseMul(a, b, zeta)
	// Pair 0: c0 = 2*4 + 17*3*5 = 8 + 255 = 263, c1 = 2*5 + 3*4 = 10+12 = 22
	if r[0] != 263 {
		t.Errorf("BaseMul[0] = %d, want 263", r[0])
	}
	if r[1] != 22 {
		t.Errorf("BaseMul[1] = %d, want 22", r[1])
	}
}

func TestBarrettReduce(t *testing.T) {
	cases := [][2]int32{{0, 0}, {3328, 3328}, {3329, 0}, {3330, 1}, {6658, 0}, {-1, 3328}}
	for _, c := range cases {
		got := barrettReduce(c[0])
		if got != c[1] {
			t.Errorf("barrettReduce(%d) = %d, want %d", c[0], got, c[1])
		}
	}
}
