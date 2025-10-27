# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [v0.1.0] - 2025-10-27
### Added
- Initial public release of `idgen`.
- `Generator` with configurable epoch, pace, bits/width, and pluggable `Obfuscator`.
- Default reversible obfuscation via Feistel(k=bits, rounds=4).
- Fixed-width lowercase base36 formatting and parsing.
- Timestamp helpers: `TimestampFromRaw`, `TimestampFromID`.
- Comprehensive tests: property tests, Feistel bijection (small domain), pacing/monotonicity, concurrency uniqueness.
- Manual-only exploratory test (`-tags=manual`) with flags for interactive evaluation.
- Package documentation and README Quickstart.
- CI workflow (build, vet, test -race, lint, tidy check, govulncheck).
- MIT License.

[v0.1.0]: https://github.com/dan-sherwin/idgen/releases/tag/v0.1.0
