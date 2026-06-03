package crumbs

import (
	"context"
	"errors"
	"fmt"
)

// NewError creates a new *Error with the given message and key-value pairs.
// The concrete return type lets callers chain With(...) without a type
// assertion.
func NewError(ctx context.Context, msg string, kv ...any) *Error {
	return newError(ctx, nil, msg, kv...)
}

// WrapError creates an *Error wrapping err. Returns nil if err is nil.
// Returns error (not *Error) to avoid typed-nil hazards at call sites.
func WrapError(ctx context.Context, err error, msg string, kv ...any) error {
	if err == nil {
		return nil
	}
	return newError(ctx, err, msg, kv...)
}

// New is an alias for NewError.
func New(ctx context.Context, msg string, kv ...any) *Error {
	return newError(ctx, nil, msg, kv...)
}

// Wrap is an alias for WrapError.
func Wrap(ctx context.Context, err error, msg string, kv ...any) error {
	if err == nil {
		return nil
	}
	return newError(ctx, err, msg, kv...)
}

// Errorf creates a new *Error with a formatted message. To attach crumbs,
// chain With(...): crumbs.Errorf(ctx, "code %d", 500).With("op", "x").
func Errorf(ctx context.Context, format string, args ...any) *Error {
	return newError(ctx, nil, fmt.Sprintf(format, args...))
}

// Wrapf wraps err with a formatted message. Returns nil if err is nil.
// To attach crumbs, prefer WrapError(ctx, err, fmt.Sprintf(...), kv...).
func Wrapf(ctx context.Context, err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return newError(ctx, err, fmt.Sprintf(format, args...))
}

// newError builds an *Error, merging crumbs from the wrapped error (if any)
// or from ctx, then applying kv pairs last-write-wins.
//
// Crumb merge rules:
//   - If err is already an *Error, its crumbs are snapshotted under its read
//     lock (so concurrent With() on inner is safe). The ctx is not re-read,
//     avoiding duplicate ctx crumbs as the error bubbles up the wrap chain.
//   - Otherwise, ctx crumbs are used as the starting set.
//   - kv pairs are then applied via upsert; existing keys are overwritten.
func newError(ctx context.Context, err error, msg string, kv ...any) *Error {
	var inner *Error
	if err != nil {
		errors.As(err, &inner)
	}

	var base []Crumb
	if inner != nil {
		inner.mu.RLock()
		if len(inner.crumbs) > 0 {
			base = make([]Crumb, len(inner.crumbs))
			copy(base, inner.crumbs)
		}
		inner.mu.RUnlock()
	} else if ctx != nil {
		if c, ok := ctx.Value(crumbsKey{}).([]Crumb); ok {
			base = c
		}
	}

	kvLen := (len(kv) + 1) / 2
	totalCap := len(base) + kvLen

	var crumbs []Crumb
	if totalCap > 0 {
		crumbs = make([]Crumb, 0, totalCap)
		crumbs = append(crumbs, base...)
		crumbs = appendKV(crumbs, kv)
	}

	return &Error{
		cause:  err,
		msg:    msg,
		crumbs: crumbs,
	}
}

// upsertCrumb appends a crumb or replaces an existing entry with the same key.
func upsertCrumb(crumbs []Crumb, key string, value any) []Crumb {
	for i := range crumbs {
		if crumbs[i].Key == key {
			crumbs[i].Value = value
			return crumbs
		}
	}
	return append(crumbs, Crumb{Key: key, Value: value})
}

// appendKV merges kv pairs into crumbs using upsert semantics
// (last-write-wins on duplicate string keys).
//
// Malformed input is preserved rather than silently dropped, so callers see
// the mistake at log time:
//   - A non-string key at an even index skips the pair.
//   - A dangling key at the end (odd-length slice) is recorded as a crumb
//     with key "!BADKEY" and the stray token as its value.
func appendKV(crumbs []Crumb, kv []any) []Crumb {
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			key, ok := kv[i].(string)
			if !ok {
				continue
			}
			crumbs = upsertCrumb(crumbs, key, kv[i+1])
			continue
		}
		crumbs = append(crumbs, Crumb{Key: "!BADKEY", Value: kv[i]})
	}
	return crumbs
}
