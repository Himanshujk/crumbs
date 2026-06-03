package crumbs

import (
	"fmt"
	"sync"
)

// Crumb is a single structured key-value pair carried by a context or an
// *Error. Value is intentionally typed as any so callers can attach numbers,
// strings, custom structs, etc. without conversion.
type Crumb struct {
	Key   string
	Value any
}

// Error is a custom error type that wraps a standard error and supports
// key-value crumbs. Its state is fully encapsulated; use the Message, Cause,
// and GetCrumbs accessors to inspect it. The zero value is usable but
// degenerate; prefer the constructors in this package.
type Error struct {
	cause  error
	msg    string
	crumbs []Crumb

	mu sync.RWMutex // guards crumbs on post-creation mutation (With)
}

// Error implements the error interface. If msg is empty the wrapped cause's
// message is returned; if neither is set, the literal "unknown error" is
// returned so log lines never collapse to a blank string.
func (e *Error) Error() string {
	if e.msg == "" {
		if e.cause != nil {
			return e.cause.Error()
		}
		return "unknown error"
	}

	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.cause.Error())
	}

	return e.msg
}

// Unwrap returns the underlying error for errors.Is and errors.As compatibility.
func (e *Error) Unwrap() error {
	return e.cause
}

// Message returns the message attached to this error (without the wrapped
// cause's message). Use Error() to get the full "msg: cause" string.
func (e *Error) Message() string {
	return e.msg
}

// Cause is an alias for Unwrap, returning the wrapped error if any.
func (e *Error) Cause() error {
	return e.cause
}

// GetCrumbs returns a defensive copy of the key-value pairs associated with
// the error. Mutating the returned slice does not affect the error.
func (e *Error) GetCrumbs() []Crumb {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.crumbs) == 0 {
		return nil
	}
	out := make([]Crumb, len(e.crumbs))
	copy(out, e.crumbs)
	return out
}

// With appends key-value crumbs to this error in place and returns it for
// chaining. Duplicate keys follow last-write-wins. Useful for adding crumbs
// after a formatted constructor:
//
//	err := crumbs.Errorf(ctx, "code %d", 500).With("op", "x")
//
// Safe to call concurrently.
func (e *Error) With(kv ...any) *Error {
	if len(kv) == 0 {
		return e
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.crumbs = appendKV(e.crumbs, kv)
	return e
}
