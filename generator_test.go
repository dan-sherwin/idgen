package idgen

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestNewDefaultsAndFormatWidth(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if g.width != 8 {
		t.Fatalf("default width = %d, want 8", g.width)
	}
	if g.ob.DomainBits() != g.bits {
		t.Fatalf("obfuscator bits %d != generator bits %d", g.ob.DomainBits(), g.bits)
	}

	// Generate one and ensure formatted width is fixed
	raw := g.Generate()
	id := g.Format(raw)
	if len(id) != 8 {
		t.Fatalf("formatted id length = %d, want 8, id=%q", len(id), id)
	}

	// Round-trip Format -> Parse
	back, err := g.Parse(id)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if back != raw {
		t.Fatalf("round-trip mismatch: got %d, want %d", back, raw)
	}
}

func TestFormatParsePropertyRandom(t *testing.T) {
	g, err := New(WithWidth(8))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	bits := g.bits
	mask := (uint64(1) << bits) - 1
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 2000; i++ {
		v := rng.Uint64() & mask
		id := g.Format(int64(v))
		if len(id) != 8 {
			t.Fatalf("iter %d: len(id)=%d want 8", i, len(id))
		}
		back, err := g.Parse(id)
		if err != nil {
			t.Fatalf("iter %d: parse error: %v", i, err)
		}
		if uint64(back)&mask != v {
			t.Fatalf("iter %d: round-trip mismatch got=%d want=%d", i, back, v)
		}
	}
}

func TestGenerateMonotonicAndPaced(t *testing.T) {
	// Use 1ms pace (default). Verify monotonic and time-based pacing.
	g, err := New(WithPace(time.Millisecond))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	start := time.Now()
	const n = 4
	vals := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		vals = append(vals, g.Generate())
	}
	elapsed := time.Since(start)
	for i := 1; i < len(vals); i++ {
		if vals[i] <= vals[i-1] {
			t.Fatalf("not strictly increasing at %d: %d <= %d", i, vals[i], vals[i-1])
		}
	}
	// Expect at least (n-1) * 1ms total due to enforced pacing. Give a small scheduling slack.
	min := (n - 1) * int(time.Millisecond)
	if int(elapsed) < min {
		t.Fatalf("elapsed %v < expected minimum %v", elapsed, time.Duration(min))
	}
}

func TestGenerateConcurrentUniqueness(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	// Launch several goroutines to call Generate once each.
	workers := 8
	if workers < runtime.GOMAXPROCS(0) {
		workers = runtime.GOMAXPROCS(0)
	}
	vals := make([]int64, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			vals[i] = g.Generate()
		}(i)
	}
	wg.Wait()

	sorted := append([]int64(nil), vals...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			t.Fatalf("duplicate generate() values: %v", sorted)
		}
	}
}

func TestTimestampConversions(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	raw := g.Generate()
	id := g.Format(raw)
	back, err := g.Parse(id)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if back != raw {
		t.Fatalf("round-trip raw mismatch: got %d want %d", back, raw)
	}

	when1 := g.TimestampFromRaw(raw)
	when2, err := g.TimestampFromID(id)
	if err != nil {
		t.Fatalf("TimestampFromID error: %v", err)
	}
	if !when1.Equal(when2) {
		t.Fatalf("timestamp mismatch: %v vs %v", when1, when2)
	}

	// Sanity: timestamps should be >= epoch and close to now
	epoch := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if when1.Before(epoch) {
		t.Fatalf("timestamp before epoch: %v < %v", when1, epoch)
	}
}

func TestOptionsValidation(t *testing.T) {
	// Invalid pace
	if _, err := New(WithPace(0)); err == nil {
		t.Fatal("expected error for pace=0")
	}
	// Invalid bits
	if _, err := New(WithBits(0)); err == nil {
		t.Fatal("expected error for bits=0")
	}
	if _, err := New(WithBits(64)); err == nil {
		t.Fatal("expected error for bits>63")
	}
	// Width too small for bits
	if _, err := New(WithWidth(2), WithBits(40)); err == nil {
		t.Fatal("expected error for width too small for bits")
	}
	// Obfuscator bits mismatch
	ob, err := NewFeistel(20, 3)
	if err != nil {
		t.Fatalf("NewFeistel error: %v", err)
	}
	if _, err := New(WithWidth(8), WithBits(41), WithObfuscation(ob)); err == nil {
		t.Fatal("expected error for obfuscator domain mismatch")
	}
}

func TestParseErrors(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if _, err := g.Parse(""); err == nil {
		t.Fatal("expected error on empty string")
	}
	if _, err := g.Parse("!!!"); err == nil {
		t.Fatal("expected error on invalid base36 string")
	}
}
