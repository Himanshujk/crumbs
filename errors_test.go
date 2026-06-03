package crumbs

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func crumbsToMap(crumbs []Crumb) map[string]interface{} {
	m := make(map[string]interface{})
	for _, c := range crumbs {
		m[c.Key] = c.Value
	}
	return m
}

func TestNewError(t *testing.T) {
	ctx := context.Background()

	t.Run("basic error creation", func(t *testing.T) {
		err := NewError(ctx, "test error")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if err.Error() != "test error" {
			t.Errorf("Expected 'test error', got '%s'", err.Error())
		}
	})

	t.Run("with crumbs", func(t *testing.T) {
		cerr := NewError(ctx, "test error", "key1", "value1", "key2", 42)

		crumbs := crumbsToMap(cerr.GetCrumbs())
		if crumbs["key1"] != "value1" {
			t.Errorf("Expected crumbs['key1'] = 'value1', got '%v'", crumbs["key1"])
		}

		if crumbs["key2"] != 42 {
			t.Errorf("Expected crumbs['key2'] = 42, got '%v'", crumbs["key2"])
		}
	})
}

func TestWrapError(t *testing.T) {
	ctx := context.Background()
	baseErr := errors.New("base error")

	t.Run("basic wrapping", func(t *testing.T) {
		err := WrapError(ctx, baseErr, "wrapped error")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !strings.Contains(err.Error(), "wrapped error") {
			t.Errorf("Expected message to contain 'wrapped error', got '%s'", err.Error())
		}

		if !strings.Contains(err.Error(), "base error") {
			t.Errorf("Expected message to contain 'base error', got '%s'", err.Error())
		}

		if !errors.Is(err, baseErr) {
			t.Error("errors.Is failed to match the base error")
		}
	})

	t.Run("with crumbs", func(t *testing.T) {
		err := WrapError(ctx, baseErr, "wrapped error", "key1", "value1")

		cerr, ok := err.(*Error)
		if !ok {
			t.Fatal("Expected *Error type")
		}

		crumbs := crumbsToMap(cerr.GetCrumbs())
		if crumbs["key1"] != "value1" {
			t.Errorf("Expected crumbs['key1'] = 'value1', got '%v'", crumbs["key1"])
		}
	})

	t.Run("wrap nil", func(t *testing.T) {
		err := WrapError(ctx, nil, "wrapped nil")
		if err != nil {
			t.Errorf("Expected nil when wrapping nil, got '%v'", err)
		}
	})
}

func TestErrorf(t *testing.T) {
	ctx := context.Background()

	err := Errorf(ctx, "formatted %s %d", "error", 42)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected := "formatted error 42"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestWrapf(t *testing.T) {
	ctx := context.Background()
	baseErr := errors.New("base error")

	err := Wrapf(ctx, baseErr, "formatted %s %d", "wrapper", 42)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "formatted wrapper 42") {
		t.Errorf("Expected message to contain 'formatted wrapper 42', got '%s'", err.Error())
	}

	if !errors.Is(err, baseErr) {
		t.Error("errors.Is failed to match the base error")
	}
}

func TestErrorsIs(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("sentinel error")

	t.Run("direct wrap", func(t *testing.T) {
		err := WrapError(ctx, sentinel, "wrapped")
		if !errors.Is(err, sentinel) {
			t.Error("errors.Is should find sentinel in direct wrap")
		}
	})

	t.Run("deep wrap", func(t *testing.T) {
		err1 := WrapError(ctx, sentinel, "inner")
		err2 := WrapError(ctx, err1, "middle")
		err3 := WrapError(ctx, err2, "outer")

		if !errors.Is(err3, sentinel) {
			t.Error("errors.Is should find sentinel in deep wrap")
		}
	})

	t.Run("different error", func(t *testing.T) {
		other := errors.New("other error")
		err := WrapError(ctx, sentinel, "wrapped")

		if errors.Is(err, other) {
			t.Error("errors.Is should not match different errors")
		}
	})
}

type customError struct {
	value int
}

func (e *customError) Error() string {
	return "custom error"
}

func TestErrorsAs(t *testing.T) {
	ctx := context.Background()
	custom := &customError{value: 42}

	t.Run("direct wrap", func(t *testing.T) {
		err := WrapError(ctx, custom, "wrapped")

		var ce *customError
		if !errors.As(err, &ce) {
			t.Error("errors.As should extract custom error")
		} else if ce.value != 42 {
			t.Errorf("Expected value 42, got %d", ce.value)
		}
	})

	t.Run("deep wrap", func(t *testing.T) {
		err1 := WrapError(ctx, custom, "inner")
		err2 := WrapError(ctx, err1, "middle")
		err3 := WrapError(ctx, err2, "outer")

		var ce *customError
		if !errors.As(err3, &ce) {
			t.Error("errors.As should extract custom error from deep wrap")
		} else if ce.value != 42 {
			t.Errorf("Expected value 42, got %d", ce.value)
		}
	})
}

func TestContextCrumbs(t *testing.T) {
	ctx := context.Background()

	// Add crumbs to context
	ctx = AddCrumb(ctx, "ctx1", "value1", "ctx2", 42)

	// Check GetCrumbs
	crumbs := crumbsToMap(GetCrumbs(ctx))
	if crumbs["ctx1"] != "value1" || crumbs["ctx2"] != 42 {
		t.Errorf("GetCrumbs failed, got %v", crumbs)
	}

	// Check that crumbs are included in errors
	cerr := NewError(ctx, "test error")

	errCrumbs := crumbsToMap(cerr.GetCrumbs())
	if errCrumbs["ctx1"] != "value1" || errCrumbs["ctx2"] != 42 {
		t.Errorf("Context crumbs not included in error, got %v", errCrumbs)
	}
}

func TestAddCrumb(t *testing.T) {
	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		ctx = AddCrumb(ctx, "key", "value")

		crumbs := crumbsToMap(GetCrumbs(ctx))
		if crumbs["key"] != "value" {
			t.Errorf("Expected crumbs['key'] = 'value', got '%v'", crumbs["key"])
		}
	})

	t.Run("existing crumbs", func(t *testing.T) {
		ctx := context.Background()
		ctx = AddCrumb(ctx, "key1", "value1")
		ctx = AddCrumb(ctx, "key2", "value2")

		crumbs := crumbsToMap(GetCrumbs(ctx))
		if crumbs["key1"] != "value1" || crumbs["key2"] != "value2" {
			t.Errorf("Expected both crumbs, got %v", crumbs)
		}
	})

	t.Run("override crumb", func(t *testing.T) {
		ctx := context.Background()
		ctx = AddCrumb(ctx, "key", "value1")
		ctx = AddCrumb(ctx, "key", "value2")

		crumbs := crumbsToMap(GetCrumbs(ctx))
		if crumbs["key"] != "value2" {
			t.Errorf("Expected crumbs['key'] = 'value2', got '%v'", crumbs["key"])
		}
	})

	t.Run("multiple crumbs at once", func(t *testing.T) {
		ctx := context.Background()
		ctx = AddCrumb(ctx, "key1", "value1", "key2", "value2")

		crumbs := crumbsToMap(GetCrumbs(ctx))
		if crumbs["key1"] != "value1" || crumbs["key2"] != "value2" {
			t.Errorf("Expected both crumbs, got %v", crumbs)
		}
	})

	t.Run("non-string key", func(t *testing.T) {
		ctx := context.Background()
		ctx = AddCrumb(ctx, 123, "value") // Should be ignored

		crumbs := crumbsToMap(GetCrumbs(ctx))
		if len(crumbs) > 0 {
			t.Errorf("Expected no crumbs, got %v", crumbs)
		}
	})
}

func TestFormatError(t *testing.T) {
	ctx := context.Background()
	err := NewError(ctx, "test error", "key1", "value1")

	t.Run("basic format", func(t *testing.T) {
		formatted := FormatError(err, false)
		if !strings.Contains(formatted, "test error") {
			t.Errorf("Formatted error should contain message, got: %s", formatted)
		}
	})

	t.Run("with crumbs", func(t *testing.T) {
		formatted := FormatError(err, true)
		if !strings.Contains(formatted, "key1") || !strings.Contains(formatted, "value1") {
			t.Errorf("Formatted error should contain crumbs, got: %s", formatted)
		}
	})

	t.Run("nil", func(t *testing.T) {
		if FormatError(nil, true) != "" {
			t.Error("FormatError(nil) should return empty string")
		}
	})
}

func TestBadKeysAndCoverage(t *testing.T) {
	ctx := context.Background()

	// 1. Wrapf with nil err
	if Wrapf(ctx, nil, "fmt %s", "a") != nil {
		t.Error("Wrapf should handle nil err")
	}

	// 2. NewError basic round-trip
	_ = NewError(ctx, "msg")

	// 3. newError dangling key
	err2 := NewError(ctx, "msg", "key", "val", "dangling")
	c2 := err2.GetCrumbs()
	if len(c2) != 2 || c2[1].Key != "!BADKEY" || c2[1].Value != "dangling" {
		t.Error("Dangling key not mapped to !BADKEY")
	}

	// 4. newError odd key that is not string
	err3 := NewError(ctx, "msg", "key", "val", 123)
	c3 := err3.GetCrumbs()
	if len(c3) != 2 || c3[1].Key != "!BADKEY" || c3[1].Value != 123 {
		t.Error("Odd non-string key not mapped to !BADKEY")
	}

	// 4b. newError even non-string key ignored
	err3b := NewError(ctx, "msg", 123, "val")
	if len(err3b.GetCrumbs()) != 0 {
		t.Error("Even non-string key not ignored")
	}

	// 5. AddCrumb dangling key
	ctx2 := AddCrumb(ctx, "key", "val", "dangling")
	crumbs := GetCrumbs(ctx2)
	if len(crumbs) != 2 || crumbs[1].Key != "!BADKEY" || crumbs[1].Value != "dangling" {
		t.Error("AddCrumb dangling key failed")
	}

	// 6. AddCrumb odd non-string key
	ctx3 := AddCrumb(ctx, "key", "val", 123)
	crumbs = GetCrumbs(ctx3)
	if len(crumbs) != 2 || crumbs[1].Key != "!BADKEY" || crumbs[1].Value != 123 {
		t.Error("AddCrumb non-string dangling key failed")
	}

	// 7. AddCrumb even non-string key
	ctx4 := AddCrumb(ctx, 123, "val")
	if len(GetCrumbs(ctx4)) != 0 {
		t.Error("AddCrumb even non-string key not ignored")
	}

	// 8. Error method fallbacks
	emptyErr := &Error{}
	if emptyErr.Error() != "unknown error" {
		t.Error("Empty Error.Error() failed")
	}
	wrapErr := &Error{cause: errors.New("base")}
	if wrapErr.Error() != "base" {
		t.Error("Error() falling back to base failed")
	}
}

func TestGetCrumbsNil(t *testing.T) {
	//nolint:staticcheck // deliberately testing nil safeguard
	if GetCrumbs(nil) != nil {
		t.Error("GetCrumbs(nil) should be nil")
	}
	if GetCrumbs(context.Background()) != nil {
		t.Error("GetCrumbs(emptyCtx) should be nil")
	}
}

// --- Regression tests for review fixes ---

// #1: Errorf/Wrapf must not inject a phantom !BADKEY crumb.
func TestErrorfNoPhantomCrumb(t *testing.T) {
	ctx := context.Background()
	err := Errorf(ctx, "formatted %d", 1)
	if len(err.GetCrumbs()) != 0 {
		t.Errorf("Errorf should produce no crumbs, got %+v", err.GetCrumbs())
	}
}

func TestWrapfNoPhantomCrumb(t *testing.T) {
	ctx := context.Background()
	base := errors.New("base")
	err := Wrapf(ctx, base, "formatted %d", 1).(*Error)
	if len(err.GetCrumbs()) != 0 {
		t.Errorf("Wrapf should produce no crumbs, got %+v", err.GetCrumbs())
	}
}

// #3: ctx crumbs must not duplicate across wraps.
func TestNoDuplicateCtxCrumbsOnWrap(t *testing.T) {
	ctx := AddCrumb(context.Background(), "req", "r-1", "user", "u-1")

	err1 := NewError(ctx, "inner")
	err2 := WrapError(ctx, err1, "middle")
	err3 := WrapError(ctx, err2, "outer")

	cerr := err3.(*Error)
	counts := map[string]int{}
	for _, c := range cerr.GetCrumbs() {
		counts[c.Key]++
	}
	if counts["req"] != 1 || counts["user"] != 1 {
		t.Errorf("expected each ctx crumb exactly once, got %v", counts)
	}
}

// GetCrumbs must return a defensive copy.
func TestGetCrumbsReturnsCopy(t *testing.T) {
	ctx := context.Background()
	err := NewError(ctx, "msg", "k", "v")
	copy1 := err.GetCrumbs()
	if len(copy1) != 1 {
		t.Fatalf("expected 1 crumb, got %d", len(copy1))
	}
	copy1[0].Value = "mutated"
	copy2 := err.GetCrumbs()
	if copy2[0].Value != "v" {
		t.Errorf("internal slice was mutated via GetCrumbs result: %v", copy2[0].Value)
	}
}

// #10: AddCrumb upsert by key.
func TestAddCrumbDedup(t *testing.T) {
	ctx := context.Background()
	ctx = AddCrumb(ctx, "k", "v1")
	ctx = AddCrumb(ctx, "k", "v2")

	all := GetCrumbs(ctx)
	if len(all) != 1 {
		t.Fatalf("expected single crumb after upsert, got %d: %+v", len(all), all)
	}
	if all[0].Value != "v2" {
		t.Errorf("expected last-write-wins value 'v2', got %v", all[0].Value)
	}
}

// (*Error).With must append/replace crumbs in place.
func TestErrorWith(t *testing.T) {
	ctx := context.Background()
	err := Errorf(ctx, "code %d", 500)
	if len(err.GetCrumbs()) != 0 {
		t.Fatalf("Errorf must produce no crumbs, got %+v", err.GetCrumbs())
	}
	err.With("op", "x", "user", "u-1")
	got := err.GetCrumbs()
	if len(got) != 2 {
		t.Fatalf("expected 2 crumbs after With, got %d", len(got))
	}
	// last-write-wins
	err.With("op", "y")
	got = err.GetCrumbs()
	if len(got) != 2 {
		t.Fatalf("expected 2 crumbs after dedup-With, got %d: %+v", len(got), got)
	}
	for _, c := range got {
		if c.Key == "op" && c.Value != "y" {
			t.Errorf("With did not overwrite op, got %v", c.Value)
		}
	}
}

// Message / Cause accessors expose state without exposing internals.
func TestErrorMessageAndCause(t *testing.T) {
	ctx := context.Background()
	base := errors.New("base")
	err := WrapError(ctx, base, "wrapped").(*Error)
	if err.Message() != "wrapped" {
		t.Errorf("Message() = %q, want %q", err.Message(), "wrapped")
	}
	if err.Cause() != base {
		t.Errorf("Cause() = %v, want %v", err.Cause(), base)
	}
	plain := NewError(ctx, "x")
	if plain.Cause() != nil {
		t.Errorf("Cause() should be nil for non-wrapping error, got %v", plain.Cause())
	}
}

// Regression: concurrent (*Error).With on the inner error must not race with
// WrapError reading inner.crumbs. Run with `go test -race` to surface failures.
func TestWrapRaceOnInnerCrumbs(t *testing.T) {
	ctx := context.Background()
	inner := NewError(ctx, "inner", "k0", "v0")

	done := make(chan struct{})
	go func() {
		for i := 0; i < 2000; i++ {
			inner.With("k0", i, "k1", i)
		}
		close(done)
	}()
	for i := 0; i < 2000; i++ {
		_ = WrapError(ctx, inner, "outer")
	}
	<-done
}

// #2: New / Wrap aliases must exist and behave like NewError / WrapError.
func TestNewWrapAliases(t *testing.T) {
	ctx := context.Background()
	err := New(ctx, "hi")
	if err.Error() != "hi" {
		t.Errorf("New alias failed: %v", err)
	}
	base := errors.New("base")
	wrapped := Wrap(ctx, base, "outer")
	if !errors.Is(wrapped, base) {
		t.Errorf("Wrap alias did not preserve identity")
	}
	if Wrap(ctx, nil, "x") != nil {
		t.Error("Wrap(nil) should return nil")
	}
}
