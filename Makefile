.PHONY: all build build-nocgo test test-nocgo test-coverage lint fmt clean run run-full deps tidy help generate-site serve-site build-site

# Variables for CGO dependency management
GOQUIRC_VERSION := $(shell go list -m -f '{{.Version}}' github.com/kdar/goquirc)
GOMODCACHE := $(shell go env GOMODCACHE)
GOQUIRC_SRC := $(GOMODCACHE)/github.com/kdar/goquirc@$(GOQUIRC_VERSION)

# Default target
all: fmt lint test build

# Vendor Go dependencies
vendor: go.mod go.sum
	@echo "Vendoring Go dependencies..."
	go mod vendor
	@touch vendor

# Copy C sources from module cache to vendor directory
vendor/github.com/kdar/goquirc/internal: vendor
	@echo "Copying goquirc C sources (version $(GOQUIRC_VERSION))..."
	@test -d "$(GOQUIRC_SRC)/internal" || \
		(echo "Error: goquirc sources not found at $(GOQUIRC_SRC)/internal" && exit 1)
	mkdir -p vendor/github.com/kdar/goquirc
	cp -r "$(GOQUIRC_SRC)/internal" vendor/github.com/kdar/goquirc/
	chmod -R u+w vendor/github.com/kdar/goquirc/internal
	@touch vendor/github.com/kdar/goquirc/internal

# Build C library in vendor directory
vendor/github.com/kdar/goquirc/libquirc.a: vendor/github.com/kdar/goquirc/internal
	@echo "Building libquirc.a in vendor directory..."
	cd vendor/github.com/kdar/goquirc && rm -f libquirc.a && $(MAKE)
	@test -f vendor/github.com/kdar/goquirc/libquirc.a || \
		(echo "Error: libquirc.a build failed" && exit 1)

# Copy library to project root (required by goquirc CGO directive)
# The goquirc package has: #cgo darwin LDFLAGS: ./libquirc.a
# This hardcoded path requires the library in the working directory
libquirc.a: vendor/github.com/kdar/goquirc/libquirc.a
	@echo "Copying libquirc.a to project root..."
	cp vendor/github.com/kdar/goquirc/libquirc.a .

# Build with CGO (default - includes goquirc decoder - 4 decoders)
build: libquirc.a
	@echo "Building with CGO..."
	CGO_ENABLED=1 go build -mod=vendor -o bin/qr-tester ./cmd/qr-tester
	@echo "Binary: bin/qr-tester"

# Build without CGO (3 decoders)
build-nocgo:
	@echo "Building without CGO..."
	CGO_ENABLED=0 go build -o bin/qr-tester-nocgo ./cmd/qr-tester
	@echo "Binary: bin/qr-tester-nocgo"

# Run tests (with CGO)
test: libquirc.a
	@echo "Running tests (CGO enabled)..."
	CGO_ENABLED=1 go test -mod=vendor -v ./...

# Run tests (without CGO)
test-nocgo:
	@echo "Running tests (CGO disabled)..."
	CGO_ENABLED=0 go test -v ./...

# Run tests with coverage
test-coverage: libquirc.a
	@echo "Running tests with coverage..."
	CGO_ENABLED=1 go test -mod=vendor -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint code
lint: libquirc.a
	@echo "Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed, skipping..."; exit 0; }
	@CGO_ENABLED=1 golangci-lint run ./... || { echo "Linting completed with warnings"; exit 0; }

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Running go vet..."
	CGO_ENABLED=0 go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ results/ coverage.out coverage.html vendor/ libquirc.a
	go clean

# Run with default settings
run: build
	@echo "Running QR compatibility tests..."
	./bin/qr-tester

# Run comprehensive test matrix (576 tests per encoder/decoder pair)
run-full: build
	@echo "Running comprehensive test matrix..."
	./bin/qr-tester -test-mode=comprehensive

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Generate Hugo site data from benchmark results
generate-site:
	@echo "Generating Hugo site data..."
	@test -d results || (echo "No results found. Run 'make run' first." && exit 1)
	go run ./cmd/generate-site

# Serve Hugo site locally for preview
serve-site: generate-site
	@echo "Starting Hugo dev server..."
	cd website && hugo server --buildDrafts

# Build Hugo site for production
build-site: generate-site
	@echo "Building Hugo site..."
	cd website && hugo --minify
	@echo "Site built in website/public/"

# Help
help:
	@echo "QR Library Test Matrix - Makefile targets:"
	@echo ""
	@echo "  make build         - Build with CGO (4 decoders, includes goquirc)"
	@echo "  make build-nocgo   - Build without CGO (3 decoders)"
	@echo "  make test          - Run tests with CGO"
	@echo "  make test-nocgo    - Run tests without CGO"
	@echo "  make test-coverage - Generate coverage report"
	@echo "  make lint          - Run linter (requires golangci-lint)"
	@echo "  make fmt           - Format code and run go vet"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make run           - Build and run standard tests (96/pair)"
	@echo "  make run-full      - Build and run comprehensive tests (576/pair)"
	@echo "  make deps          - Download dependencies"
	@echo "  make tidy          - Tidy go.mod"
	@echo "  make generate-site - Generate Hugo data from results"
	@echo "  make serve-site    - Preview site locally"
	@echo "  make build-site    - Build production site"
	@echo "  make help          - Show this help"
	@echo ""
	@echo "CGO Build Process (automatic):"
	@echo "  vendor -> vendor C sources -> build libquirc.a -> copy to root -> build"
	@echo ""
	@echo "Website Build Process:"
	@echo "  run -> generate-site -> serve-site/build-site"
	@echo ""
	@echo "Examples:"
	@echo "  make                      # Format, lint, test, build (with CGO)"
	@echo "  make build-nocgo          # Build without CGO support"
	@echo "  make test-coverage        # Generate coverage report"
	@echo "  make run && make serve-site  # Run benchmarks and preview site"
