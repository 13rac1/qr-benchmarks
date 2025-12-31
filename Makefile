.PHONY: all build build-cgo test test-cgo test-coverage lint fmt clean run run-full deps tidy help

# Default target
all: fmt lint test build

# Build without CGO (default - 3 decoders)
build:
	@echo "Building without CGO..."
	CGO_ENABLED=0 go build -o bin/qr-tester ./cmd/qr-tester
	@echo "Binary: bin/qr-tester"

# Build with CGO (includes goquirc decoder - 4 decoders)
build-cgo:
	@echo "Building with CGO..."
	CGO_ENABLED=1 go build -tags cgo -o bin/qr-tester-cgo ./cmd/qr-tester
	@echo "Binary: bin/qr-tester-cgo"

# Run tests (without CGO)
test:
	@echo "Running tests (CGO disabled)..."
	CGO_ENABLED=0 go test -v ./...

# Run tests (with CGO)
test-cgo:
	@echo "Running tests (CGO enabled)..."
	CGO_ENABLED=1 go test -tags cgo -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint code
lint:
	@echo "Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed, skipping..."; exit 0; }
	@CGO_ENABLED=0 golangci-lint run ./... || { echo "Linting completed with warnings"; exit 0; }

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Running go vet..."
	CGO_ENABLED=0 go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ results/ coverage.out coverage.html
	go clean

# Run with default settings
run: build
	@echo "Running QR compatibility tests..."
	./bin/qr-tester

# Run with custom config
run-full: build
	@echo "Running full test matrix..."
	./bin/qr-tester -parallel=true -output=./results

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Help
help:
	@echo "QR Library Test Matrix - Makefile targets:"
	@echo ""
	@echo "  make build         - Build without CGO (3 decoders)"
	@echo "  make build-cgo     - Build with CGO (4 decoders, requires C compiler)"
	@echo "  make test          - Run tests without CGO"
	@echo "  make test-cgo      - Run tests with CGO"
	@echo "  make test-coverage - Generate coverage report"
	@echo "  make lint          - Run linter (requires golangci-lint)"
	@echo "  make fmt           - Format code and run go vet"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make run           - Build and run with default settings"
	@echo "  make run-full      - Build and run full test matrix"
	@echo "  make deps          - Download dependencies"
	@echo "  make tidy          - Tidy go.mod"
	@echo "  make help          - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make                      # Format, lint, test, build"
	@echo "  make build-cgo            # Build with CGO support"
	@echo "  make test-coverage        # Generate coverage report"
