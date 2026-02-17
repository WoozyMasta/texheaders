# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

## [0.1.1][] - 2026-02-18

### Added

* `BuildOptions.Workers` parallelism modes for builder:
  serial by default (`0`/`1`), explicit worker count (`>1`), and
  `WorkersAuto` (`-1`) auto-selection mode.
* `WorkersAuto` constant and worker-resolution helpers for predictable
  opt-in parallel builds.
* `BenchmarkBuildFromAppendedFilesWorkers` with `Workers=0/4/8/16` cases.
* Builder worker-parity and worker-resolution tests.
* New sentinel errors: `ErrEmptyInputPath` and `ErrNilFile`.

### Changed

* Updated `github.com/woozymasta/paa` to `v0.2.2`.
  `paa v0.2.2` also brought faster metadata paths and convenient APIs
  (`DecodeMetadataHeaders`, `DecodeMetadataHeadersBytes`,
  `EncodeWithOptionsAndMetadataHeaders`) used by texheaders build flow.
* Build pipeline now reads PAA metadata via
  `paa.DecodeMetadataHeaders` (header-only path) instead of full tag map.
* Builder internals now avoid unnecessary re-sorting when appended inputs are
  already ordered.
* Decode/encode internals reduced append growth in fixed-size slices
  (`Textures` and `MipMaps`) by pre-sizing and index assignment.
* Read/write internals improved byte/string I/O paths
  (`io.ByteReader`/`io.StringWriter` fast paths).

[0.1.1]: https://github.com/WoozyMasta/texheaders/compare/v0.1.0...v0.1.1

## [0.1.0][] - 2026-02-08

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/texheaders/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
