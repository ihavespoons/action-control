.PHONY: build test lint clean test-unit test-integration

# Build binary
build:
	go build -o bin/action-control

# Run unit tests
test-unit:
	go test -v ./internal/...

# Run integration tests
test-integration:
	go test -v ./tests/...

# Run all tests
test: test-unit test-integration
	@echo "All tests completed successfully"

# Lint the code
lint:
	go vet ./...
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed, skipping linting"; \
	fi

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Test coverage
coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out