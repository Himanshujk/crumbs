.PHONY: test test-fuzz benchmark coverage lint clean

# Default target
all: test lint benchmark

# Run all tests with the race detector
test:
	go test -race -count=1 -v ./...

# Run all fuzz tests for a short period (smoke test, not a full corpus build)
test-fuzz:
	go test -run=NONE -fuzz=FuzzAddCrumb -fuzztime=10s .
	go test -run=NONE -fuzz=FuzzNewError -fuzztime=10s .

# Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

# Generate test coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, installing v2.6.0..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.0; \
		golangci-lint run ./...; \
	fi

# Clean up generated files
clean:
	rm -f coverage.out coverage.html
	
# Show help
help:
	@echo "Available targets:"
	@echo "  all        : Run tests, linter and benchmarks"
	@echo "  test       : Run all tests with the race detector"
	@echo "  test-fuzz  : Run fuzz suites for 10s each (smoke test)"
	@echo "  benchmark  : Run all benchmarks"
	@echo "  coverage   : Generate test coverage report"
	@echo "  lint       : Run golangci-lint"
	@echo "  clean      : Remove generated files"
