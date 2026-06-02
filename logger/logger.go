// Package logger defines a minimal, dependency-free logger interface that
// crumbs adapters implement. Concrete adapters live under
// github.com/sri-shubham/crumbs/integrations/* and bridge this interface to
// a real logging backend (e.g. log/slog) while transparently extracting
// crumbs from context.Context and *crumbs.Error values in the args list.
package logger

import "context"

// Logger is a generic interface for logging with context. All level methods
// share the same signature, matching the slog convention: an error can be
// passed as a regular key-value pair (e.g. "error", err) and adapters built
// on top of crumbs will extract crumbs from any error in the args list.
//
// With returns a derived Logger that includes the supplied key-value args
// on every subsequent call. Implementations should return a logger that
// shares the same underlying sink.
type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}
