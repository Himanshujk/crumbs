package crumbs

import "context"

// crumbsKey is the unexported context key under which the []Crumb snapshot
// is stored. A struct{} type guarantees no collisions with other packages.
type crumbsKey struct{}

// AddCrumb returns a new context carrying the supplied key-value pairs in
// addition to any crumbs already present on ctx. Duplicate keys follow
// last-write-wins. The input ctx is never mutated.
func AddCrumb(ctx context.Context, kv ...any) context.Context {
	if len(kv) == 0 {
		return ctx
	}

	var crumbs []Crumb

	if existing, ok := ctx.Value(crumbsKey{}).([]Crumb); ok {
		crumbs = make([]Crumb, len(existing), len(existing)+(len(kv)+1)/2)
		copy(crumbs, existing)
	} else {
		crumbs = make([]Crumb, 0, (len(kv)+1)/2)
	}

	crumbs = appendKV(crumbs, kv)

	return context.WithValue(ctx, crumbsKey{}, crumbs)
}

// GetCrumbs returns a defensive copy of the crumbs stored on ctx, or nil if
// none are present. A nil ctx is tolerated and yields nil.
func GetCrumbs(ctx context.Context) []Crumb {
	if ctx == nil {
		return nil
	}

	if crumbs, ok := ctx.Value(crumbsKey{}).([]Crumb); ok {
		result := make([]Crumb, len(crumbs))
		copy(result, crumbs)
		return result
	}

	return nil
}
