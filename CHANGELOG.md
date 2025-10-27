# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [v0.1.3] - 2025-10-27
### Changed
- Human-readable format now groups base36 IDs in chunks of 4 separated by dashes (e.g., `xxxx-xxxx`). Padding to the configured width still occurs before grouping.
- `Parse` accepts dashed strings by stripping `-` before parsing; undelimited inputs remain supported.
- Updated tests and examples to validate grouping and ignore dashes for width checks.

### Notes
- This is a backward-compatible change for parsing; only display formatting changed.

## [v0.1.2] - 2025-10-27
### Fixed
- GitHub CI workflow stability and run fixes.

### Changed
- Documentation updates and minor wording tweaks. No library code changes.

## [v0.1.1] - 2025-10-27
### Fixed
- Initial GitHub CI workflow setup and stability tweaks.

### Notes
- No functional/library code changes in this release.

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

[v0.1.3]: https://github.com/dan-sherwin/idgen/releases/tag/v0.1.3
[v0.1.2]: https://github.com/dan-sherwin/idgen/releases/tag/v0.1.2
[v0.1.1]: https://github.com/dan-sherwin/idgen/releases/tag/v0.1.1
[v0.1.0]: https://github.com/dan-sherwin/idgen/releases/tag/v0.1.0
