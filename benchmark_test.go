package crumbs

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

var (
	errBench      = errors.New("benchmark error")
	benchResult   error
	benchCtxCrumb context.Context
)

// Benchmarks for error creation
func BenchmarkErrorsNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchResult = errors.New("benchmark error")
	}
}

func BenchmarkCrumbsNewError(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		benchResult = NewError(ctx, "benchmark error")
	}
}

func BenchmarkCrumbsNewErrorWithCrumbs(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		benchResult = NewError(ctx, "benchmark error",
			"key1", "value1",
			"key2", 2,
			"key3", true)
	}
}

// Benchmarks for error wrapping
func BenchmarkErrorsWrap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchResult = fmt.Errorf("wrapped: %w", errBench)
	}
}

func BenchmarkCrumbsWrapError(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		benchResult = WrapError(ctx, errBench, "wrapped")
	}
}

func BenchmarkCrumbsWrapErrorWithCrumbs(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		benchResult = WrapError(ctx, errBench, "wrapped",
			"key1", "value1",
			"key2", 2,
			"key3", true)
	}
}

// Benchmarks for context operations
func BenchmarkAddCrumb(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		benchCtxCrumb = AddCrumb(ctx, "key", "value")
	}
}

func BenchmarkAddMultipleCrumbs(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		benchCtxCrumb = AddCrumb(ctx,
			"key1", "value1",
			"key2", 2,
			"key3", true)
	}
}

func BenchmarkGetCrumbs(b *testing.B) {
	ctx := context.Background()
	ctx = AddCrumb(ctx,
		"key1", "value1",
		"key2", 2,
		"key3", true)

	b.ResetTimer()
	var result []Crumb
	for i := 0; i < b.N; i++ {
		result = GetCrumbs(ctx)
	}
	_ = result
}

// Benchmark error formatting
func BenchmarkFormatError(b *testing.B) {
	ctx := context.Background()
	err := NewError(ctx, "benchmark error", "key1", "value1", "key2", 2)

	b.ResetTimer()
	var result string
	for i := 0; i < b.N; i++ {
		result = FormatError(err, true)
	}
	_ = result
}
