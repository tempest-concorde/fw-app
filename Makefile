.PHONY: help build test test-integration lint lint-fix run clean

# Variables
BINARY_NAME=fw-app
MAIN_PATH=./cmd/server
GO=go
GOLANGCI_LINT=golangci-lint

help:
	@echo "Flight Wall Application - Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build            Build the binary"
	@echo "  test             Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  lint             Run golangci-lint"
	@echo "  lint-fix         Run golangci-lint with auto-fix"
	@echo "  run              Run the application locally"
	@echo "  clean            Remove build artifacts"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=1 $(GO) build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Built: $(BINARY_NAME)"

test:
	@echo "Running unit tests..."
	$(GO) test -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "✅ Tests passed"

test-integration:
	@echo "Running integration tests..."
	$(GO) test -race -tags=integration ./test/integration/...
	@echo "✅ Integration tests passed"

lint:
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run
	@echo "✅ Linting passed"

lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	$(GOLANGCI_LINT) run --fix
	@echo "✅ Linting complete"

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.txt
	@echo "✅ Cleaned"
