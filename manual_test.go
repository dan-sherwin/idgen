//go:build manual

package idgen

import (
	"flag"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// Flags for manual exploration. Use: go test -tags=manual -run TestManualOptions -v -- [flags]
var (
	fEpoch      = flag.String("epoch", "2025-01-01T00:00:00Z", "Epoch in RFC3339 (UTC recommended), e.g., 2025-01-01T00:00:00Z")
	fPace       = flag.Duration("pace", time.Millisecond, "Min spacing between IDs (e.g., 1ms, 250us, 2s)")
	fWidth      = flag.Int("width", 12, "Fixed base36 width (e.g., 8 → ~41 bits)")
	fBits       = flag.Uint("bits", 0, "Domain bits (optional). If 0, derived from width")
	fRounds     = flag.Int("rounds", 4, "Feistel rounds (>=2)")
	fCount      = flag.Int("n", 10, "How many IDs to generate sequentially")
	fConcurrent = flag.Int("concurrent", 0, "If >0, also spawn this many goroutines to Generate() once each")
	fShowRaw    = flag.Bool("show_raw", false, "Print raw numeric value alongside formatted ID")
	fShowTS     = flag.Bool("show_ts", true, "Print timestamp decoded from ID")
)

func TestManualOptions(t *testing.T) {
	// Parse epoch
	epStr := strings.TrimSpace(*fEpoch)
	epoch, err := time.Parse(time.RFC3339, epStr)
	if err != nil {
		t.Fatalf("invalid epoch %q: %v", epStr, err)
	}
	// Build options
	opts := []Option{WithEpoch(epoch), WithPace(*fPace), WithWidth(*fWidth)}
	if *fBits != 0 {
		opts = append(opts, WithBits(*fBits))
	}
	// Use Feistel with the generator's bits; we can only construct it after New if bits derived from width.
	// So we first build with defaults to compute bits; then, if rounds differ from default, rebuild with explicit obfuscator.

	// First pass: create with current opts to get computed bits
	g1, err := New(opts...)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	// If rounds requested differ from default 4, create a Feistel with same domain bits
	if *fRounds != 4 {
		ob, err := NewFeistel(g1.bits, *fRounds)
		if err != nil {
			t.Fatalf("NewFeistel(bits=%d, rounds=%d) error: %v", g1.bits, *fRounds, err)
		}
		opts2 := append(opts, WithBits(g1.bits), WithObfuscation(ob))
		g1, err = New(opts2...)
		if err != nil {
			t.Fatalf("New(with custom feistel) failed: %v", err)
		}
	}
	g := g1

	t.Logf("idgen manual: epoch=%s pace=%s width=%d bits=%d rounds=%d", epoch.UTC().Format(time.RFC3339), g.pace, g.width, g.bits, *fRounds)

	// Sequential generation
	type row struct {
		idx  int
		raw  int64
		id   string
		when time.Time
	}
	rows := make([]row, 0, *fCount)
	start := time.Now()
	for i := 0; i < *fCount; i++ {
		raw := g.Generate()
		id := g.Format(raw)
		back, err := g.Parse(id)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", id, err)
		}
		if back != raw {
			t.Fatalf("round-trip mismatch at i=%d: back=%d raw=%d", i, back, raw)
		}
		var when time.Time
		if *fShowTS {
			when = g.TimestampFromRaw(raw)
		}
		rows = append(rows, row{idx: i, raw: raw, id: id, when: when})
	}
	elapsed := time.Since(start)

	// Print table
	for _, r := range rows {
		switch {
		case *fShowRaw && *fShowTS:
			t.Logf("%3d  id=%s  raw=%d  ts=%s", r.idx, r.id, r.raw, r.when.Format(time.RFC3339Nano))
		case *fShowRaw:
			t.Logf("%3d  id=%s  raw=%d", r.idx, r.id, r.raw)
		case *fShowTS:
			t.Logf("%3d  id=%s  ts=%s", r.idx, r.id, r.when.Format(time.RFC3339Nano))
		default:
			t.Logf("%3d  id=%s", r.idx, r.id)
		}
	}
	// Pacing expectation for sequential part
	// Expect at least (n-1) * pace, with some tolerance.
	expectedMin := time.Duration(max(0, *fCount-1)) * g.pace
	if elapsed < expectedMin {
		t.Logf("note: elapsed %v < expected pacing floor %v (system scheduling may vary)", elapsed, expectedMin)
	} else {
		t.Logf("sequential generation elapsed=%v (expected >= %v)", elapsed, expectedMin)
	}

	// Optional concurrency exercise
	if *fConcurrent > 0 {
		t.Logf("concurrency test: %d goroutines, each Generate() once", *fConcurrent)
		vals := make([]int64, *fConcurrent)
		var wg sync.WaitGroup
		wg.Add(*fConcurrent)
		for i := 0; i < *fConcurrent; i++ {
			go func(i int) {
				defer wg.Done()
				vals[i] = g.Generate()
			}(i)
		}
		wg.Wait()
		// Check uniqueness
		sorted := append([]int64(nil), vals...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
		dup := false
		for i := 1; i < len(sorted); i++ {
			if sorted[i] == sorted[i-1] {
				dup = true
				break
			}
		}
		if dup {
			t.Fatalf("duplicate values detected in concurrent generate: %v", sorted)
		}
		t.Logf("concurrent generate OK: %d unique values", *fConcurrent)
	}

	// Summary line for easy visual scan
	t.Logf("summary: n=%d concurrent=%d pace=%s width=%d bits=%d rounds=%d elapsed=%v", *fCount, *fConcurrent, g.pace, g.width, g.bits, *fRounds, elapsed)

	// Provide example command in output for convenience
	t.Logf("run with: go test -tags=manual -run TestManualOptions -v -- -epoch=%s -pace=%s -width=%d -n=%d -concurrent=%d", epoch.UTC().Format(time.RFC3339), g.pace, g.width, *fCount, *fConcurrent)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
