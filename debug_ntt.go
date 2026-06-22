package main

import (
	"fmt"
	"github.com/muhammadalihussnain/mlkem-webapp/backend/mlkem"
)

func main() {
	// Test 1: Polynomial with a single non-zero coefficient at position 0
	var poly1 [256]int32
	poly1[0] = 1
	
	fmt.Println("Test 1: poly with poly[0]=1")
	fmt.Printf("Original[0:5] = %v\n", poly1[0:5])
	
	fwd1 := mlkem.NTTForward(poly1)
	fmt.Printf("After Forward NTT[0:5] = %v\n", fwd1[0:5])
	
	inv1 := mlkem.NTTInverse(fwd1)
	fmt.Printf("After Inverse NTT[0:5] = %v\n", inv1[0:5])
	
	// Check if round-trip works
	match := true
	mismatches := 0
	for i := 0; i < 256; i++ {
		if poly1[i] != inv1[i] {
			match = false
			mismatches++
			if mismatches <= 10 {
				fmt.Printf("Mismatch at %d: orig=%d, recovered=%d\n", i, poly1[i], inv1[i])
			}
		}
	}
	fmt.Printf("Round-trip match: %v (mismatches: %d)\n\n", match, mismatches)
	
	// Test 2: Zero check
	var zero [256]int32
	fwdZero := mlkem.NTTForward(zero)
	invZero := mlkem.NTTInverse(fwdZero)
	zeroMatch := true
	for i := 0; i < 256; i++ {
		if invZero[i] != 0 {
			zeroMatch = false
			if i < 5 {
				fmt.Printf("Zero mismatch at %d: got %d\n", i, invZero[i])
			}
		}
	}
	fmt.Printf("Zero polynomial round-trip: %v\n", zeroMatch)
}
