package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	Q     int32 = 3329
	N     int   = 256
	ZETA  int32 = 17
	INV_N int32 = 3303
)

var Zetas = [128]int32{
	1, 1729, 2580, 3289, 2642, 630, 1897, 848,
	1062, 1919, 193, 797, 2786, 3260, 569, 1746,
	296, 2447, 1339, 1476, 3046, 56, 2240, 1333,
	1426, 2094, 535, 2882, 2393, 2879, 1974, 821,
	289, 331, 3253, 1756, 1197, 2304, 2277, 2055,
	650, 1977, 2513, 632, 2865, 33, 1320, 1915,
	2319, 1435, 807, 452, 1438, 2868, 1534, 2402,
	2647, 2617, 1481, 648, 2474, 3110, 1227, 910,
	17, 2761, 583, 2649, 1637, 723, 2288, 1100,
	1409, 2662, 3281, 233, 756, 2156, 3015, 3050,
	1703, 1651, 2789, 1789, 1847, 952, 1461, 2687,
	939, 2308, 2437, 2388, 733, 2337, 268, 641,
	1584, 2298, 2037, 3220, 375, 2549, 2090, 1645,
	1063, 319, 2773, 757, 2099, 561, 2466, 2594,
	2804, 1092, 403, 1026, 1143, 2150, 2775, 886,
	1722, 1212, 1874, 1029, 2110, 2935, 885, 2154,
}

func barrettReduce(x int32) int32 {
	const V = 20159
	t := (int64(x) * V) >> 26
	x = int32(int64(x) - t*int64(Q))
	if x >= Q {
		x -= Q
	}
	if x < 0 {
		x += Q
	}
	return x
}

func addMod(a, b int32) int32 {
	return barrettReduce(a + b)
}

func subMod(a, b int32) int32 {
	return barrettReduce(a - b)
}

func mulMod(a, b int32) int32 {
	return barrettReduce(a * b)
}

func NTTForward(poly [N]int32) [N]int32 {
	var result [N]int32
	copy(result[:], poly[:])
	k := 1
	length := N / 2
	for length >= 1 {
		for start := 0; start < N; start += 2 * length {
			zeta := Zetas[k]
			k++
			for j := start; j < start+length; j++ {
				t := mulMod(zeta, result[j+length])
				result[j+length] = subMod(result[j], t)
				result[j] = addMod(result[j], t)
			}
		}
		length >>= 1
	}
	return result
}

func NTTInverse(poly [N]int32) [N]int32 {
	var result [N]int32
	copy(result[:], poly[:])
	k := N/2 - 1
	length := 1
	for length < N {
		for start := 0; start < N; start += 2 * length {
			zeta := Zetas[k]
			k--
			for j := start; j < start+length; j++ {
				t := result[j+length]
				result[j+length] = subMod(result[j], t)
				result[j] = addMod(result[j], t)
				result[j+length] = mulMod(zeta, result[j+length])
			}
		}
		length <<= 1
	}
	for i := 0; i < N; i++ {
		result[i] = mulMod(result[i], INV_N)
	}
	return result
}

func NaivePolyMul(a, b [N]int32) [N]int32 {
	var result [N]int32
	for i := 0; i < N; i++ {
		if a[i] == 0 {
			continue
		}
		for j := 0; j < N-i; j++ {
			if b[j] == 0 {
				continue
			}
			result[i+j] = addMod(result[i+j], mulMod(a[i], b[j]))
		}
	}
	return result
}

func NTTPolyMul(a, b [N]int32) [N]int32 {
	aNTT := NTTForward(a)
	bNTT := NTTForward(b)
	var productNTT [N]int32
	for i := 0; i < N; i++ {
		productNTT[i] = mulMod(aNTT[i], bNTT[i])
	}
	return NTTInverse(productNTT)
}

func GenerateRandomPolynomial() [N]int32 {
	var poly [N]int32
	for i := 0; i < N; i++ {
		poly[i] = int32(rand.Intn(int(Q)))
	}
	return poly
}

func ComparePolynomials(a, b [N]int32) bool {
	for i := 0; i < N; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func PrintPolynomial(name string, poly [N]int32, n int) {
	fmt.Printf("%s[0..%d] = ", name, n-1)
	for i := 0; i < n && i < N; i++ {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", poly[i])
	}
	if n < N {
		fmt.Printf(", ...")
	}
	fmt.Println()
}

func VerifyRootProperties() bool {
	fmt.Println("\n=== Verifying NTT Root Properties ===")
	pow := int32(1)
	for i := 0; i < N/2; i++ {
		pow = mulMod(pow, ZETA)
	}
	fmt.Printf("ζ^(%d) mod %d = %d (expected %d)\n", N/2, Q, pow, Q-1)
	if pow != Q-1 {
		fmt.Println("FAILED")
		return false
	}
	fmt.Println("PASS: ζ^(N/2) ≡ -1 mod Q")

	pow = int32(1)
	for i := 0; i < N; i++ {
		pow = mulMod(pow, ZETA)
	}
	fmt.Printf("ζ^%d mod %d = %d (expected 1)\n", N, Q, pow)
	if pow != 1 {
		fmt.Println("FAILED")
		return false
	}
	fmt.Println("PASS: ζ^N ≡ 1 mod Q")
	return true
}

func VerifyNTTRoundTrip() bool {
	fmt.Println("\n=== Verifying NTT Round-trip ===")
	testPoly := GenerateRandomPolynomial()
	PrintPolynomial("Original", testPoly, 6)
	transformed := NTTForward(testPoly)
	recovered := NTTInverse(transformed)
	PrintPolynomial("Recovered", recovered, 6)
	if ComparePolynomials(testPoly, recovered) {
		fmt.Println("PASS: NTT round-trip")
		return true
	}
	fmt.Println("FAIL: NTT round-trip")
	return false
}

func VerifyNTTPolyMul(numTests int) bool {
	fmt.Printf("\n=== Verifying NTT Multiplication (%d tests) ===\n", numTests)
	allPassed := true
	for test := 0; test < numTests; test++ {
		a := GenerateRandomPolynomial()
		b := GenerateRandomPolynomial()
		naiveResult := NaivePolyMul(a, b)
		nttResult := NTTPolyMul(a, b)
		if !ComparePolynomials(naiveResult, nttResult) {
			fmt.Printf("FAIL: Test %d\n", test+1)
			PrintPolynomial("  Naive", naiveResult, 8)
			PrintPolynomial("  NTT  ", nttResult, 8)
			allPassed = false
			break
		}
		if (test+1)%5 == 0 {
			fmt.Printf("  %d tests passed\n", test+1)
		}
	}
	if allPassed {
		fmt.Printf("PASS: All %d tests\n", numTests)
	} else {
		fmt.Println("FAIL: Some tests failed")
	}
	return allPassed
}

func BenchmarkMultiplication(iterations int) {
	fmt.Println("\n=== Performance Benchmark ===")
	var naiveTime, nttTime time.Duration
	for i := 0; i < iterations; i++ {
		a := GenerateRandomPolynomial()
		b := GenerateRandomPolynomial()
		start := time.Now()
		NaivePolyMul(a, b)
		naiveTime += time.Since(start)
		start = time.Now()
		NTTPolyMul(a, b)
		nttTime += time.Since(start)
	}
	fmt.Printf("Naive: %v (avg %v)\n", naiveTime, naiveTime/time.Duration(iterations))
	fmt.Printf("NTT:   %v (avg %v)\n", nttTime, nttTime/time.Duration(iterations))
	if nttTime > 0 {
		fmt.Printf("Speedup: %.2fx\n", float64(naiveTime)/float64(nttTime))
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("============================================================")
	fmt.Println("    ML-KEM NTT Polynomial Multiplication Verification")
	fmt.Println("============================================================")
	fmt.Printf("Modulus Q = %d, N = %d, ζ = %d\n", Q, N, ZETA)
	fmt.Println()

	rootOK := VerifyRootProperties()
	roundTripOK := VerifyNTTRoundTrip()
	multiplicationOK := VerifyNTTPolyMul(10)
	BenchmarkMultiplication(50)

	fmt.Println("\n============================================================")
	fmt.Println("                    FINAL SUMMARY")
	fmt.Println("============================================================")
	fmt.Printf("Root Properties:    %s\n", map[bool]string{true: "PASS", false: "FAIL"}[rootOK])
	fmt.Printf("NTT Round-trip:     %s\n", map[bool]string{true: "PASS", false: "FAIL"}[roundTripOK])
	fmt.Printf("NTT Multiplication: %s\n", map[bool]string{true: "PASS", false: "FAIL"}[multiplicationOK])

	if rootOK && roundTripOK && multiplicationOK {
		fmt.Println("\nALL TESTS PASSED!")
		fmt.Println("INTT(NTT(A) * NTT(B)) = Naive multiplication")
	} else {
		fmt.Println("\nSome tests failed. Please review.")
	}
}
