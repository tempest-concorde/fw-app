.PHONY: help build test test-integration lint lint-fix fmt-check run run-dev clean swagger swagger-check verify

# Variables
BINARY_NAME=fw-app
MAIN_PATH=./cmd/server
GO=go
GOLANGCI_LINT=golangci-lint
SWAG=swag

help:
	@echo "Flight Wall Application - Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build            Build the binary"
	@echo "  test             Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  lint             Run golangci-lint"
	@echo "  lint-fix         Run golangci-lint with auto-fix"
	@echo "  fmt-check        Check code formatting with gofmt"
	@echo "  verify           Run all verification checks (fmt-check + lint + swagger-check)"
	@echo "  swagger          Generate Swagger documentation"
	@echo "  swagger-check    Verify Swagger docs are up to date"
	@echo "  run              Run the application locally"
	@echo "  run-dev          Run the application with debug logging"
	@echo "  clean            Remove build artifacts"
	@echo ""

swagger:
	@echo "Generating Swagger documentation..."
	$(SWAG) init -g cmd/server/main.go -o docs/ --parseDependency --parseInternal
	@echo "✅ Swagger docs generated"

swagger-check:
	@echo "Checking Swagger documentation..."
	@mkdir -p .tmp-swagger
	$(SWAG) init -g cmd/server/main.go -o .tmp-swagger/ --parseDependency --parseInternal
	@diff -r docs/ .tmp-swagger/ || (echo "❌ Swagger docs are out of date. Run 'make swagger' to update." && rm -rf .tmp-swagger && exit 1)
	@rm -rf .tmp-swagger
	@echo "✅ Swagger docs are up to date"

fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -l .)" || (echo "❌ The following files need formatting:" && gofmt -l . && exit 1)
	@echo "✅ All Go files are properly formatted"

verify: fmt-check lint swagger-check
	@echo "✅ All verification checks passed"

build: swagger
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 $(GO) build -o $(BINARY_NAME) $(MAIN_PATH)
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

run-dev: build
	@echo "Running $(BINARY_NAME) in development mode..."
	./$(BINARY_NAME) --log-format=text --log-level=debug

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.txt
	rm -rf .tmp-swagger
	@echo "✅ Cleaned"
