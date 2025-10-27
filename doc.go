// Package idgen provides compact, human-friendly, reversible identifiers.
//
// Overview
//   - Generate: emits a raw tick (int64) based on time since a configurable epoch,
//     paced by a minimum interval (e.g., 1ms). The pacing enforces single-process
//     monotonicity and acts as a natural throttle.
//   - Format/Parse: converts between raw ticks and fixed-width, lowercase base36
//     strings using a reversible bounded-domain obfuscation (Feistel network).
//     The obfuscation makes adjacent times appear non-sequential to humans while
//     remaining fully reversible for operations/debugging.
//   - Timestamp helpers: convert to/from UTC timestamps.
//
// Non-goals
//   - Cryptography: the obfuscation is not encryption and is not designed to
//     resist adversarial analysis. It is intended only to reduce visual patterns
//     that encourage human error when copying/reading IDs.
//   - Global uniqueness across processes: the library ensures monotonic spacing
//     within a single process. If you need multi-process safety, incorporate a
//     coordinated node ID or use a database/UUID/ULID strategy.
//
// Defaults
//   - Epoch: 2025-01-01T00:00:00Z
//   - Pace:  1ms per ID (at most ~1000 IDs/sec), no sequence field
//   - Width: 8 base36 characters (~41 bits domain)
//   - Obfuscation: Feistel(k=bits, rounds=4)
//
// Example quickstart:
//
//	g, _ := idgen.New()                  // default epoch/pace/width/obfuscation
//	raw := g.Generate()                  // paced monotonic raw tick
//	id  := g.Format(raw)                 // fixed 8-char base36 string
//	back, _ := g.Parse(id)               // == raw
//	when := g.TimestampFromID(id)        // UTC timestamp
//
// Width, bits, and horizon
// With width=8 the domain fits ~41 bits, which at 1ms pace comfortably covers
// decades (until ~2094 relative to the default epoch). Increase width or reduce
// pace for longer horizons.
package idgen
