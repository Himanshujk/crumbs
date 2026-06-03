# Slog Integration for Crumbs

This package provides an adapter to use Go's structured logging library `log/slog` with `crumbs`. It implements the `logger.Logger` interface, allowing you to seamlessly integrate `crumbs` rich context and error handling with `slog`.

## Features

- **Automatic Context Extraction**: Automatically extracts crumbs from `context.Context` and adds them as structured attributes to your logs.
- **Error Enrichment**: Automatically extracts crumbs attached to `crumbs.Error` and adds them to the log entry when logging errors.
- **Standard Interface**: Implements the generic `logger.Logger` interface, decoupling your application code from the specific logging implementation.

## Usage

### Installation

```bash
go get github.com/sri-shubham/crumbs/integrations/slog
```

### Basic Usage

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/sri-shubham/crumbs"
	crumbslog "github.com/sri-shubham/crumbs/integrations/slog"
)

func main() {
	// 1. Create your slog logger
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	slogger := slog.New(jsonHandler)

	// 2. Create the crumbs adapter
	log := crumbslog.New(slogger)

	// 3. Use the adapter in your application
	ctx := context.Background()
	
	// Add context crumbs
	ctx = crumbs.AddCrumb(ctx, "request_id", "req-123")

	// Log info - context crumbs are automatically included
	log.Info(ctx, "starting operation", "component", "api")
	// Output: {"time":"...", "level":"INFO", "msg":"starting operation", "component":"api", "request_id":"req-123"}

	// Create an error with crumbs
	err := crumbs.New(ctx, "database connection failed", "db_host", "localhost")

	// Log error using the slog-style key/value pairing. The adapter
	// stringifies error values and splats crumbs from the first
	// *crumbs.Error it finds in the args.
	log.Error(ctx, "operation failed", "error", err)
	// Output: {"time":"...", "level":"ERROR", "msg":"operation failed", "error":"database connection failed: ...", "db_host":"localhost", "request_id":"req-123"}
}
```

### Derived loggers

`Adapter.With(args ...any)` returns a child `logger.Logger` that automatically
attaches the supplied key/value pairs to every subsequent call, mirroring
`slog.Logger.With`:

```go
reqLog := log.With("request_id", "req-123")
reqLog.Info(ctx, "handling request")
```
