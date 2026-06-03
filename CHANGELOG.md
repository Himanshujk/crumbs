# Changelog

All notable changes to this project are documented in this file. The format
is loosely based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and the project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Package-level godoc on `crumbs`, `crumbs/logger`, and the slog adapter
  describing the mental model and concurrency guarantees.
- `logger.Logger.With(args ...any) Logger` for deriving child loggers that
  carry persistent attributes. Implemented on the slog adapter via
  `slog.Logger.With`.
- First-class slog adapter pointer in the root README; the adapter README
  documents the `With` method and the corrected key/value error-arg style.
- Regression test (`TestWrapRaceOnInnerCrumbs`) for concurrent
  `(*Error).With` against `WrapError` under `-race`.
- Shared `appendKV` helper backing `AddCrumb`, `newError`, and
  `(*Error).With`; behavior is unchanged but parsing now lives in one place.

### Changed

- **Breaking:** `NewError`, `New`, and `Errorf` now return `*Error` instead
  of `error`. This lets callers chain `.With(...)` without a type assertion.
  Existing code that assigns the result to an `error`-typed variable
  continues to work (`*Error` satisfies `error`). Code that does
  `crumbs.NewError(...).(*Error)` becomes a compile error and should drop
  the assertion.
- `WrapError`, `Wrap`, and `Wrapf` still return `error` to preserve the
  nil-on-nil-input invariant.
- `errors.go` (≈300 LOC) split into focused files: `doc.go`, `error.go`
  (the `Error` type and its methods), `constructors.go`
  (`New*`/`Wrap*`/`Errorf`/`Wrapf` + `newError`/`appendKV`/`upsertCrumb`),
  `context.go` (`AddCrumb`/`GetCrumbs`), and `format.go` (`FormatError`).
- All public APIs now use `...any` rather than `...interface{}` in their
  signatures for consistency with current Go style.
- Minimum supported Go version bumped to **1.22**.

### Fixed

- **Data race** in `newError` reading `inner.crumbs`. The read lock on the
  inner error was released before the slice elements were copied into the
  new error, so a concurrent `(*Error).With` on the inner could overwrite
  elements via in-place upsert and race with the read. The inner slice is
  now snapshotted into a fresh backing array while the read lock is held.

### Removed

- Stack-trace support (`StackFrame`, `ConfigureStackTraces`, `ForceStack`,
  `FormatStack`, `(*Error).GetStack`, and the `includeStack` parameter on
  `FormatError`) was removed in the prior release. This entry is repeated
  here for discoverability.

## Migration notes (Unreleased)

- Drop any `crumbs.NewError(ctx, msg, ...).(*Error)` / `.(*Error)`
  assertions on `New`/`Errorf` results; the constructors return `*Error`
  directly.
- If you implement `logger.Logger` outside of the bundled slog adapter,
  add a `With(args ...any) Logger` method.
- If you depend on `go.mod` setting `go 1.21`, regenerate after pulling
  this release (`go mod tidy`).
