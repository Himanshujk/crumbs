package crumbs

import (
	"context"
	"testing"
)

// FuzzAddCrumb drives AddCrumb / GetCrumbs round-trips with random byte
// sequences. The invariants exercised are:
//   - AddCrumb never panics on arbitrary inputs.
//   - GetCrumbs returns a slice whose last entry for a given even-position
//     string key reflects last-write-wins.
//   - The returned slice is a defensive copy: mutating it does not affect
//     subsequent GetCrumbs calls.
func FuzzAddCrumb(f *testing.F) {
	f.Add("k1", "v1", "k2", "v2")
	f.Add("", "", "k", "")
	f.Add("dup", "a", "dup", "b")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		ctx := context.Background()
		ctx = AddCrumb(ctx, k1, v1, k2, v2)

		got := GetCrumbs(ctx)
		// Defensive copy check: mutating result must not influence stored ctx.
		if len(got) > 0 {
			orig := got[0].Value
			got[0].Value = "MUTATED-BY-FUZZ"
			again := GetCrumbs(ctx)
			if len(again) > 0 && again[0].Value != orig {
				t.Fatalf("GetCrumbs returned a shared slice; mutation leaked back: %v", again[0])
			}
		}

		// Last-write-wins for the same string key.
		if k1 == k2 && k1 != "" {
			again := GetCrumbs(ctx)
			seen := 0
			var last any
			for _, c := range again {
				if c.Key == k1 {
					seen++
					last = c.Value
				}
			}
			if seen != 1 {
				t.Fatalf("dup key %q should appear once, saw %d in %+v", k1, seen, again)
			}
			if last != v2 {
				t.Fatalf("last-write-wins violated for %q: got %v, want %v", k1, last, v2)
			}
		}
	})
}

// FuzzNewError checks that NewError never panics on random kv shapes and
// that well-formed string-keyed input never produces a "!BADKEY" crumb.
func FuzzNewError(f *testing.F) {
	f.Add("msg", "k", "v")
	f.Add("", "", "")
	f.Add("oops", "a", "b")

	f.Fuzz(func(t *testing.T, msg, k, v string) {
		ctx := context.Background()
		err := NewError(ctx, msg, k, v)
		if err == nil {
			t.Fatal("NewError returned nil")
		}
		_ = err.Error() // must not panic regardless of msg
		for _, c := range err.GetCrumbs() {
			if c.Key == "!BADKEY" {
				t.Fatalf("well-formed input produced !BADKEY: input=(%q,%q) crumbs=%+v",
					k, v, err.GetCrumbs())
			}
		}
	})
}
