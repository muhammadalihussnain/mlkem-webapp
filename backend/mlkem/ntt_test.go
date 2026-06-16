package mlkem

import (
	"math/rand"
	"testing"
)

// Fixed seeds for deterministic test runs.
const (
	seedRoundTrip  = 42
	seedPolyMul    = 99
	seedMatVecMul  = 7
	seedPolyToBytes = 13
)

// nttRoundTripTrials is the number of random polynomials tested in the NTT round-trip.
const nttRoundTripTrials = 100

// polyMulTrials is the number of random polynomial pairs tested against naive multiplication.
const polyMulTrials = 20

// polyToBytesTrials is the number of random polynomials tested in encode/decode round-trip.
const polyToBytesTrials = 10

// maxCoeff12Bit is the maximum coefficient value for 12-bit encoding tests.
const maxCoeff12Bit = 4096

// randomPoly returns a polynomial with uniformly random coefficients in [0, Q).
func randomPoly(r *rand.Rand) [N]int32 {
	var p [N]int32
	for i := range p {
		p[i] = int32(r.Intn(int(Q)))
	}
	return p
}

// naiveNegacyclicMul computes the negacyclic convolution of a and b in Z_Q[X]/(X^N + 1)
// using the schoolbook O(N^2) algorithm. Used as a reference for TestPolyMulMatchesNaive.
func naiveNegacyclicMul(a, b [N]int32) [N]int32 {
	var tmp [2 * N]int32
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			tmp[i+j] = (tmp[i+j] + a[i]*b[j]) % Q
		}
	}
	// Reduce mod X^N + 1: the term x^(N+k) wraps to -x^k.
	var result [N]int32
	for i := 0; i < N; i++ {
		result[i] = (tmp[i] - tmp[i+N]%Q + Q) % Q
	}
	return result
}

// TestNTTRoundTrip verifies that NTTInverse(NTTForward(p)) == p for random polynomials.
func TestNTTRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(seedRoundTrip))
	for trial := 0; trial < nttRoundTripTrials; trial++ {
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

// TestPolyMulMatchesNaive verifies that PolyMul produces the same result as the
// schoolbook negacyclic convolution for random polynomial pairs.
func TestPolyMulMatchesNaive(t *testing.T) {
	r := rand.New(rand.NewSource(seedPolyMul))
	for trial := 0; trial < polyMulTrials; trial++ {
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

// TestMatVecMulDimensions verifies that MatVecMul produces the correct result
// for k=2, 3, and 4, and that rows beyond k remain zero.
func TestMatVecMulDimensions(t *testing.T) {
	r := rand.New(rand.NewSource(seedMatVecMul))

	for _, k := range []int{k512, k768, k1024} {
		var A [4][4][N]int32
		var s [4][N]int32

		for i := 0; i < k; i++ {
			s[i] = randomPoly(r)
			for j := 0; j < k; j++ {
				A[i][j] = randomPoly(r)
			}
		}

		result := MatVecMul(A, s, k)

		// Verify each output row equals the dot product of the corresponding matrix row with s.
		for i := 0; i < k; i++ {
			var expected [N]int32
			for j := 0; j < k; j++ {
				prod := PolyMul(A[i][j], s[j])
				for idx := 0; idx < N; idx++ {
					expected[idx] = addmod(expected[idx], prod[idx])
				}
			}
			for idx := 0; idx < N; idx++ {
				if result[i][idx] != expected[idx] {
					t.Fatalf("k=%d row=%d idx=%d: want %d got %d",
						k, i, idx, expected[idx], result[i][idx])
				}
			}
		}

		// Rows beyond k must be untouched (zero).
		for i := k; i < 4; i++ {
			for idx := 0; idx < N; idx++ {
				if result[i][idx] != 0 {
					t.Fatalf("k=%d: row %d (>= k) should be zero at idx %d, got %d",
						k, i, idx, result[i][idx])
				}
			}
		}
	}
}

// TestPolyAdd verifies element-wise addition with wrap-around modulo Q.
func TestPolyAdd(t *testing.T) {
	var a, b [N]int32
	for i := range a {
		a[i] = int32(i)
		b[i] = int32(N - i)
	}
	result := PolyAdd(a, b)
	expected := int32(N) % Q
	for i := range result {
		if result[i] != expected {
			t.Fatalf("PolyAdd[%d]: want %d got %d", i, expected, result[i])
		}
	}
}

// TestPolySub verifies element-wise subtraction with modular wrap-around.
func TestPolySub(t *testing.T) {
	var a, b [N]int32
	for i := range a {
		a[i] = int32(i * 2)
		b[i] = int32(i)
	}
	result := PolySub(a, b)
	for i := range result {
		if result[i] != int32(i) {
			t.Fatalf("PolySub[%d]: want %d got %d", i, i, result[i])
		}
	}
}

// TestBaseMul verifies a single pair computation of the exported BaseMul function.
func TestBaseMul(t *testing.T) {
	var a, b [N]int32
	// Use small values so the expected result is easy to compute by hand.
	a[0], a[1] = 2, 3
	b[0], b[1] = 4, 5
	zeta := Zetas[1] // a well-defined, non-trivial zeta

	r := BaseMul(a, b, zeta)

	// c0 = a0*b0 + zeta*a1*b1 mod Q
	wantC0 := addmod(mulmod(a[0], b[0]), mulmod(mulmod(zeta, a[1]), b[1]))
	// c1 = a0*b1 + a1*b0 mod Q
	wantC1 := addmod(mulmod(a[0], b[1]), mulmod(a[1], b[0]))

	if r[0] != wantC0 {
		t.Errorf("BaseMul[0] = %d, want %d", r[0], wantC0)
	}
	if r[1] != wantC1 {
		t.Errorf("BaseMul[1] = %d, want %d", r[1], wantC1)
	}
}

// TestBarrettReduce verifies barrettReduce for boundary and negative inputs.
func TestBarrettReduce(t *testing.T) {
	cases := []struct {
		input    int32
		expected int32
	}{
		{0, 0},
		{Q - 1, Q - 1},   // max in-range value
		{Q, 0},            // exact modulus
		{Q + 1, 1},        // one above modulus
		{2 * Q, 0},        // double modulus
		{-1, Q - 1},       // small negative
		{-Q, 0},           // negative modulus: triggers x >= Q branch after Barrett
	}
	for _, c := range cases {
		got := barrettReduce(c.input)
		if got != c.expected {
			t.Errorf("barrettReduce(%d) = %d, want %d", c.input, got, c.expected)
		}
	}
}

// TestPolyToBytesBytesToPoly verifies that PolyToBytes and BytesToPoly are inverses.
func TestPolyToBytesBytesToPoly(t *testing.T) {
	r := rand.New(rand.NewSource(seedPolyToBytes))
	for trial := 0; trial < polyToBytesTrials; trial++ {
		var poly [N]int32
		for i := range poly {
			poly[i] = int32(r.Intn(maxCoeff12Bit))
		}
		encoded := PolyToBytes(poly)
		if len(encoded) != polyByteLen {
			t.Fatalf("PolyToBytes length = %d, want %d", len(encoded), polyByteLen)
		}
		recovered := BytesToPoly(encoded)
		for i := range poly {
			if poly[i] != recovered[i] {
				t.Fatalf("trial %d index %d: want %d got %d", trial, i, poly[i], recovered[i])
			}
		}
	}
}

// TestMulMontgomery exercises the mulMontgomery helper for correctness.
func TestMulMontgomery(t *testing.T) {
	cases := []struct {
		a, b, want int32
	}{
		{0, 0, 0},
		{1, 1, 1},
		{2, 3, 6},
		{Q - 1, 1, Q - 1},
		{Q, 1, 0},        // Q ≡ 0 mod Q
		{100, 200, (100 * 200) % Q},
	}
	for _, c := range cases {
		got := mulMontgomery(c.a, c.b)
		if got != c.want {
			t.Errorf("mulMontgomery(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
