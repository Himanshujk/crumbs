package slog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/sri-shubham/crumbs"
	crumbslog "github.com/sri-shubham/crumbs/integrations/slog"
	"github.com/sri-shubham/crumbs/logger"
)

func TestAdapter_ImplementsLogger(t *testing.T) {
	var _ logger.Logger = (*crumbslog.Adapter)(nil)
}

func TestAdapter_Logging(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h)
	adapter := crumbslog.New(l)

	ctx := context.Background()

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		adapter.Info(ctx, "info message", "key", "value")

		var logEntry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to unmarshal log entry: %v", err)
		}

		if logEntry["msg"] != "info message" {
			t.Errorf("expected msg 'info message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "INFO" {
			t.Errorf("expected level 'INFO', got %v", logEntry["level"])
		}
		if logEntry["key"] != "value" {
			t.Errorf("expected key 'value', got %v", logEntry["key"])
		}
	})

	t.Run("Context Crumbs", func(t *testing.T) {
		buf.Reset()
		ctxWithCrumbs := crumbs.AddCrumb(ctx, "request_id", "12345")
		adapter.Info(ctxWithCrumbs, "context crumbs")

		var logEntry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to unmarshal log entry: %v", err)
		}

		if logEntry["request_id"] != "12345" {
			t.Errorf("expected request_id '12345', got %v", logEntry["request_id"])
		}
	})

	t.Run("Error Crumbs", func(t *testing.T) {
		buf.Reset()
		err := crumbs.NewError(ctx, "something went wrong", "user_id", "u-999")
		adapter.Error(ctx, "error occurred", "error", err)

		var logEntry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to unmarshal log entry: %v", err)
		}

		if logEntry["msg"] != "error occurred" {
			t.Errorf("expected msg 'error occurred', got %v", logEntry["msg"])
		}
		if logEntry["error"] != "something went wrong" {
			t.Errorf("expected error 'something went wrong', got %v", logEntry["error"])
			t.Logf("full log entry: %v", logEntry)
		}
		if logEntry["user_id"] != "u-999" {
			t.Errorf("expected user_id 'u-999', got %v", logEntry["user_id"])
		}
	})

	// Regression for review #11: when the error carries crumbs sourced from
	// ctx, the adapter must not emit the ctx crumbs a second time.
	t.Run("NoDuplicateCrumbs", func(t *testing.T) {
		buf.Reset()
		ctxWithCrumbs := crumbs.AddCrumb(ctx, "request_id", "rid-1")
		err := crumbs.NewError(ctxWithCrumbs, "boom")
		adapter.Error(ctxWithCrumbs, "failed", "error", err)

		// Count occurrences of "request_id" as a key in raw output.
		raw := buf.String()
		count := strings.Count(raw, `"request_id"`)
		if count != 1 {
			t.Errorf("expected request_id to appear exactly once, got %d in: %s", count, raw)
		}
	})

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		adapter.Debug(ctx, "debug message", "k", "v")
		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if entry["level"] != "DEBUG" || entry["msg"] != "debug message" || entry["k"] != "v" {
			t.Errorf("unexpected debug entry: %v", entry)
		}
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		adapter.Warn(ctx, "warn message")
		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if entry["level"] != "WARN" || entry["msg"] != "warn message" {
			t.Errorf("unexpected warn entry: %v", entry)
		}
	})

	t.Run("NonCrumbsErrorValue", func(t *testing.T) {
		buf.Reset()
		adapter.Error(ctx, "plain", "error", errors.New("boom"))
		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if entry["error"] != "boom" {
			t.Errorf("plain error not stringified, got %v", entry["error"])
		}
	})

	t.Run("MultipleErrorsOnlyFirstCrumbsSplatted", func(t *testing.T) {
		buf.Reset()
		first := crumbs.NewError(ctx, "first", "src", "a")
		second := crumbs.NewError(ctx, "second", "src", "b")
		adapter.Error(ctx, "both", "first", first, "second", second)
		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		// First crumbs.Error's "src" wins; second.src must not overwrite it.
		if entry["src"] != "a" {
			t.Errorf("expected src=a from first crumbs.Error, got %v", entry["src"])
		}
	})
}

// DisabledLevel verifies the early-exit path when the underlying handler
// does not enable the requested level: nothing is written, including no
// crumbs lookup side effects.
func TestAdapter_DisabledLevel(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	adapter := crumbslog.New(slog.New(h))

	ctx := crumbs.AddCrumb(context.Background(), "should_not_appear", "x")
	adapter.Debug(ctx, "noop")
	adapter.Info(ctx, "noop")
	adapter.Warn(ctx, "noop")
	if buf.Len() != 0 {
		t.Errorf("expected no output when level is disabled, got: %s", buf.String())
	}
}

// With produces a child logger that carries fixed attributes on every call
// while still preserving crumb extraction semantics.
func TestAdapter_With(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	adapter := crumbslog.New(slog.New(h))

	child := adapter.With("service", "checkout")
	ctx := crumbs.AddCrumb(context.Background(), "request_id", "rid-2")
	child.Info(ctx, "ok", "stage", "auth")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry["service"] != "checkout" {
		t.Errorf("With did not propagate base attribute, got %v", entry["service"])
	}
	if entry["stage"] != "auth" {
		t.Errorf("per-call attribute missing, got %v", entry["stage"])
	}
	if entry["request_id"] != "rid-2" {
		t.Errorf("ctx crumb missing, got %v", entry["request_id"])
	}
}
