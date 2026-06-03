// Package crumbs is a rich-observability error library: it attaches structured
// key-value pairs ("crumbs") to errors and contexts and propagates them
// transparently through wrap chains so logs and reports can include the full
// context of where things went wrong without polluting function signatures.
//
// # Mental model
//
//   - Crumbs live in two places: a context.Context (via AddCrumb / GetCrumbs)
//     and any *Error value (carried as an immutable snapshot from creation).
//   - When you create an *Error (NewError / Errorf / Wrap / WrapError / Wrapf),
//     all ctx crumbs are snapshotted onto it. Wrapping an existing *Error
//     re-uses its snapshot rather than re-reading ctx, so duplicates do not
//     accumulate as the error bubbles up.
//   - kv pairs supplied to constructors and to (*Error).With follow
//     last-write-wins semantics keyed on the string key.
//   - Adapters (e.g. integrations/slog) extract crumbs at log time so error
//     messages, context fields, and crumb fields all land in one log line.
//
// # Concurrency
//
// *Error is safe for concurrent reads (Error, Message, Cause, Unwrap,
// GetCrumbs) and concurrent calls to With. AddCrumb returns a new
// context.Context and never mutates the input.
package crumbs
