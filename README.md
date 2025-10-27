# idgen — compact, reversible, human-friendly IDs for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/dan-sherwin/idgen.svg)](https://pkg.go.dev/github.com/dan-sherwin/idgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/dan-sherwin/idgen)](https://goreportcard.com/report/github.com/dan-sherwin/idgen)
[![CI](https://github.com/dan-sherwin/idgen/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/dan-sherwin/idgen/actions/workflows/ci.yml)

`idgen` produces short, fixed-width base36 identifiers that:
- Reversible back to a timestamp (handy for ops/debugging)
- Hide visual sequential patterns (bounded-domain Feistel obfuscation)
- Enforce a minimum spacing per process (e.g., 1ms) for monotonicity and gentle throttling
- Easy to configure: epoch, pace, width/bits, obfuscator

This package is instance-based and thread-safe. It uses no global state.

## Quickstart

```go
package main

import (
	"fmt"
	"time"

	"github.com/dan-sherwin/idgen"
)

func main() {
 g, _ := idgen.New() // defaults: epoch=2025-01-01 UTC, pace=1ms (calls may briefly wait), width=8, Feistel obfuscation
 raw := g.Generate()        // paced, monotonic raw tick (since epoch, in pace units)
 id := g.Format(raw)        // fixed 8-char base36 string (displayed grouped as xxxx-xxxx)
 back, _ := g.Parse(id)     // == raw
 when := g.TimestampFromRaw(back)
 fmt.Println(id, when.UTC().Format(time.RFC3339Nano))
}
```

## Why not UUID/ULID?
- UUID: long and not reversible; great global uniqueness, but not compact.
- ULID/KSUID: sortable and excellent properties, but 26–27 chars. `idgen` targets the shortest comfortable IDs (e.g., 8 chars base36) with reversible time semantics.
- `idgen` is for human comfort and operational reversibility, not cryptographic uniqueness.

## Guarantees and limitations
- Single-process monotonicity via pacing. By default, `Generate()` emits at most one ID per 1 ms; callers may wait briefly if called faster than the pace.
- No cross-process coordination. If you run multiple processes generating IDs simultaneously, they are independent. For multi-process global uniqueness, use a DB/UUID/ULID or add a node ID scheme in a wrapper.
- Obfuscation ≠ encryption. The Feistel permutation hides visual patterns, but it is not a security boundary.

## Configuration

Use options with `idgen.New(...)`.

- `WithEpoch(time.Time)`: base time (default: 2025-01-01T00:00:00Z)
- `WithPace(time.Duration)`: minimum spacing between IDs (default: 1ms)
- `WithWidth(int)`: fixed base36 width (default: 8)
- `WithBits(uint)`: domain size in bits; if not set, derived from width (2^bits ≤ 36^width)
- `WithObfuscation(Obfuscator)`: reversible permutation over [0, 2^bits)

Typical default (good for decades at 1ms pace): width=8 ⇒ ≈41 bits domain ⇒ fits until ~2094 with default epoch.

### Common helpers
- `Generate() int64`: returns raw tick since epoch (units of pace)
- `Format(raw) string`: base36, fixed width, obfuscated; displayed grouped in chunks of 4 with dashes (e.g., `xxxx-xxxx`)
- `Parse(id) (int64, error)`: reverse of `Format`; accepts dashed or undashed strings
- `TimestampFromRaw(raw) time.Time`: UTC timestamp for a raw tick
- `TimestampFromID(id) (time.Time, error)`: parse and convert to UTC time

### Supported Go versions
Tested with Go 1.25+. The module’s `go` directive is `1.25.1`.

## Manual exploration test
A manual-only test lets you experiment without affecting CI.

Examples (from repo root):

```bash
# defaults
go test -tags=manual -run TestManualOptions -v -- -n=10
# change pace and rounds
go test -tags=manual -run TestManualOptions -v -- -pace=5ms -rounds=6 -n=6
# choose epoch, bits, and show raw + timestamps
go test -tags=manual -run TestManualOptions -v -- -epoch=2025-06-01T00:00:00Z -width=8 -bits=41 -show_raw=true -show_ts=true -n=5
```

## Time horizon guidance
At 1ms pace with width=8 (~41 bits), you have runway well beyond 60 years from the default epoch. If you:
- Need longer: increase width to 9 (≈46–47 bits) or reduce pace.
- Need shorter output: keep width=8 and 1ms pace; it’s the sweet spot for compactness.

## License
MIT. See `LICENSE`.

## Versioning
Semantic Versioning. First public tag: `v0.1.0`.

After publishing this module to GitHub:
- Ensure CI is green (see workflow below).
- Update downstream projects (e.g., Chronix): remove any local `replace` and `require github.com/dan-sherwin/idgen v0.1.0`.

## CI expectations
- `go mod tidy` check (no diffs)
- `go build ./...`
- `go vet ./...`
- `go test ./... -race`
- `golangci-lint run`
- `govulncheck ./...` (optional)
