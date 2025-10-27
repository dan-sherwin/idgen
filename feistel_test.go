package idgen

import (
	"math/rand"
	"testing"
	"time"
)

func TestFeistelBijectionSmallK(t *testing.T) {
	k := uint(12)
	ob, err := NewFeistel(k, 4)
	if err != nil {
		t.Fatalf("NewFeistel error: %v", err)
	}
	if ob.DomainBits() != k {
		t.Fatalf("DomainBits=%d want %d", ob.DomainBits(), k)
	}
	mask := (uint64(1) << k) - 1

	seen := make([]bool, 1<<12)
	for x := uint64(0); x <= mask; x++ {
		y := ob.Obfuscate(x)
		if y > mask {
			t.Fatalf("y out of range: %d > mask", y)
		}
		// Uniqueness (permutation) check
		if seen[y] {
			t.Fatalf("collision at y=%d", y)
		}
		seen[y] = true
		x2 := ob.Deobfuscate(y)
		if x2 != x {
			t.Fatalf("round-trip failed: got %d want %d", x2, x)
		}
	}
}

func TestFeistelRoundTripRandom41(t *testing.T) {
	k := uint(41)
	ob, err := NewFeistel(k, 4)
	if err != nil {
		t.Fatalf("NewFeistel error: %v", err)
	}
	mask := (uint64(1) << k) - 1
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 2000; i++ {
		x := rng.Uint64() & mask
		y := ob.Obfuscate(x)
		x2 := ob.Deobfuscate(y)
		if x2 != x {
			t.Fatalf("iter %d: round-trip mismatch got=%d want=%d", i, x2, x)
		}
	}
}
