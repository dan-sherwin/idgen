package idgen

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Obfuscator defines a reversible permutation over a k-bit domain [0, 2^k).
type Obfuscator interface {
	DomainBits() uint
	Obfuscate(x uint64) uint64
	Deobfuscate(y uint64) uint64
}

// Generator produces compact, reversible IDs paced by a minimum time quantum.
type Generator struct {
	// immutable config
	epochMS int64
	pace    time.Duration
	bits    uint
	width   int
	ob      Obfuscator

	// internal state
	mu       sync.Mutex
	lastTick int64
	seqBits  uint // reserved for future, currently 0
	seq      uint64
}

// Option configures a Generator.
type Option func(*Generator) error

// WithEpoch sets the custom epoch used to compute raw ticks.
func WithEpoch(t time.Time) Option {
	return func(g *Generator) error {
		g.epochMS = t.UTC().UnixMilli()
		return nil
	}
}

// WithPace sets the minimum spacing between generated IDs.
func WithPace(d time.Duration) Option {
	return func(g *Generator) error {
		if d <= 0 {
			return errors.New("pace must be > 0")
		}
		g.pace = d
		return nil
	}
}

// WithBits sets the domain size in bits for obfuscation and encoding.
func WithBits(bits uint) Option {
	return func(g *Generator) error {
		if bits == 0 || bits > 63 {
			return errors.New("bits must be in [1,63]")
		}
		g.bits = bits
		return nil
	}
}

// WithWidth sets the fixed base36 width for formatted IDs.
func WithWidth(width int) Option {
	return func(g *Generator) error {
		if width < 1 {
			return errors.New("width must be >= 1")
		}
		g.width = width
		return nil
	}
}

// WithObfuscation sets a custom obfuscation permutation.
func WithObfuscation(ob Obfuscator) Option {
	return func(g *Generator) error {
		if ob == nil {
			return errors.New("obfuscator cannot be nil")
		}
		g.ob = ob
		return nil
	}
}

// New constructs a new Generator with provided options and sensible defaults:
// - epoch: 2025-01-01T00:00:00Z
// - pace: 1ms
// - width: 8 (implies bits≈41)
// - bits: derived from width if zero, ensuring 36^width ≥ 2^bits
// - obfuscation: Feistel(bits, 4)
func New(opts ...Option) (*Generator, error) {
	g := &Generator{
		epochMS: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
		pace:    time.Millisecond,
		bits:    0, // derive from width if 0
		width:   8, // default width
		seqBits: 0, // default no sequence
	}
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}
	// Derive bits from width if not set
	if g.bits == 0 {
		// bits = floor(log2(36^width)) = floor(width * log2(36))
		// log2(36) ≈ 5.16992500144
		bits := uint(math.Floor(float64(g.width) * 5.16992500144))
		if bits < 1 {
			bits = 1
		}
		g.bits = bits
	}
	// Ensure width is enough to hold bits (36^width >= 2^bits)
	if !widthSupportsBits(g.width, g.bits) {
		return nil, errors.New("width too small for selected bits")
	}
	// Default obfuscator if not provided
	if g.ob == nil {
		ob, err := NewFeistel(g.bits, 4)
		if err != nil {
			return nil, err
		}
		g.ob = ob
	} else if g.ob.DomainBits() != g.bits {
		return nil, errors.New("obfuscator domain bits mismatch")
	}
	return g, nil
}

func widthSupportsBits(width int, bits uint) bool {
	// Compare using logs to avoid big ints: width*log2(36) >= bits
	return float64(width)*5.16992500144+1e-12 >= float64(bits)
}

// Generate returns a raw tick count (int64) since epoch in units of pace.
// It enforces monotonicity and the configured minimum spacing.
func (g *Generator) Generate() int64 {
	q := g.pace
	if q <= 0 {
		q = time.Millisecond
	}
	d := q / time.Millisecond
	if d <= 0 {
		// support sub-ms by rounding up to 1ms granularity for now
		d = 1
	}
	for {
		nowTick := (time.Now().UnixMilli() - g.epochMS) / int64(d)
		g.mu.Lock()
		if nowTick > g.lastTick {
			g.lastTick = nowTick
			g.mu.Unlock()
			return nowTick
		}
		g.mu.Unlock()
		time.Sleep(q)
	}
}

// Format converts a raw tick into a fixed-width lowercase base36 string using the obfuscator.
func (g *Generator) Format(raw int64) string {
	mask := uint64((uint64(1) << g.bits) - 1)
	v := uint64(raw) & mask
	obf := g.ob.Obfuscate(v)
	s := strconv.FormatUint(obf, 36)
	if len(s) < g.width {
		var b strings.Builder
		for i := 0; i < g.width-len(s); i++ {
			b.WriteByte('0')
		}
		b.WriteString(s)
		s = b.String()
	}
	return s
}

// Parse reverses Format and returns the raw tick value.
func (g *Generator) Parse(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty id")
	}
	v, err := strconv.ParseUint(s, 36, 64)
	if err != nil {
		return 0, err
	}
	mask := uint64((uint64(1) << g.bits) - 1)
	v &= mask
	raw := g.ob.Deobfuscate(v) & mask
	return int64(raw), nil
}

// TimestampFromRaw converts a raw tick to time.Time in UTC.
func (g *Generator) TimestampFromRaw(raw int64) time.Time {
	q := g.pace
	if q <= 0 {
		q = time.Millisecond
	}
	d := q / time.Millisecond
	if d <= 0 {
		d = 1
	}
	ms := g.epochMS + raw*int64(d)
	return time.UnixMilli(ms).UTC()
}

// TimestampFromID parses a formatted ID and returns the UTC timestamp.
func (g *Generator) TimestampFromID(id string) (time.Time, error) {
	raw, err := g.Parse(id)
	if err != nil {
		return time.Time{}, err
	}
	return g.TimestampFromRaw(raw), nil
}
