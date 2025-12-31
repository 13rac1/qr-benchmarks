# QR Code Library Compatibility Matrix Tester

A systematic testing tool for identifying QR encoder/decoder incompatibilities across Go libraries.

## The Problem

Different QR code libraries handle module sizing differently. When you encode a QR code with one library and decode with another, the decode can fail at certain pixel sizes due to fractional module boundaries.

**Real-world example**: The `skip2/go-qrcode` encoder and `makiuchi-d/gozxing` decoder are incompatible at specific pixel sizes:

```
skip2/go-qrcode encode + gozxing decode
Bytes\Px  320  400  440  450  460  480  512  560
--------+----------------------------------------
 500    |  ✓    ✓    ✗    ✓    ✓    ✓    ✓    ✓
 550    |  ✓    ✓    ✗    ✓    ✗    ✓    ✓    ✓
 600    |  ✓    ✓    ✓    ✗    ✓    ✓    ✓    ✓
 650    |  ✓    ✗    ✗    ✓    ✓    ✓    ✓    ✓
```

Notice the failures are non-monotonic: 440px fails while 460px succeeds. This is the fractional module sizing problem.

**Root cause**: skip2 uses fractional module pixel sizes (e.g., 12.857 pixels per module) while gozxing's decoder assumes integer boundaries. When the decoder calculates module positions, rounding errors cause misreads.

**Solution**: This tool systematically tests all encoder/decoder combinations to identify which pairs work reliably and at what pixel sizes.

## Quick Start

```bash
# Clone
git clone https://github.com/yourusername/qr-library-test.git
cd qr-library-test

# Install dependencies
go mod download

# Build
make build

# Run tests (generates reports in ./results/)
./bin/qr-tester

# View results
ls results/
```

## Features

- **4 Encoders**: skip2/go-qrcode, boombuler/barcode, yeqown/go-qrcode, gozxing
- **4 Decoders**: gozxing, tuotoo/qrcode, goqr (archived), goquirc (CGO)
- **Matrix Testing**: 6 data sizes × 8 pixel sizes × 4 error levels = 192 test cases per encoder/decoder pair
- **Module Analysis**: Calculates actual module pixel sizes and identifies fractional values
- **Pattern Detection**: Finds non-monotonic failures that indicate module boundary issues
- **Parallel Execution**: Concurrent testing with configurable worker pool
- **Multiple Reports**: Markdown reports per encoder/decoder combination

## Installation

### Prerequisites

- Go 1.20 or later
- (Optional) C compiler for CGO support (goquirc decoder)

### Build Options

**Without CGO (3 decoders, portable)**:
```bash
make build
# Creates: bin/qr-tester
```

**With CGO (4 decoders, includes goquirc)**:
```bash
make build-cgo
# Creates: bin/qr-tester-cgo
```

### Install Dependencies

```bash
make deps
# Or: go mod download
```

## Usage

### Basic Usage

Run all tests with default settings:

```bash
./bin/qr-tester
```

This tests:
- 4 encoders × 3 decoders (or 4 with CGO) = 12-16 combinations
- 6 data sizes: 500, 550, 600, 650, 750, 800 bytes
- 8 pixel sizes: 320, 400, 440, 450, 460, 480, 512, 560 pixels
- 4 error correction levels: L, M, Q, H
- Generates markdown reports in `./results/`

### Custom Configuration

Test specific combinations:

```bash
./bin/qr-tester \
  -data-sizes "500,600,700" \
  -pixel-sizes "400,480,512" \
  -error-levels "M,Q" \
  -output ./my-results
```

Skip archived and CGO libraries:

```bash
./bin/qr-tester -skip-archived=true -skip-cgo=true
```

Run serially (easier for debugging):

```bash
./bin/qr-tester -parallel=false
```

### Configuration Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-data-sizes` | string | "500,550,600,650,750,800" | Comma-separated byte sizes to test |
| `-pixel-sizes` | string | "320,400,440,450,460,480,512,560" | Comma-separated image dimensions |
| `-error-levels` | string | "L,M,Q,H" | Comma-separated error correction levels |
| `-parallel` | bool | true | Run tests concurrently |
| `-timeout` | duration | 10s | Timeout per decoder operation |
| `-max-workers` | int | NumCPU | Maximum concurrent workers |
| `-skip-cgo` | bool | false | Skip CGO-based decoders (goquirc) |
| `-skip-archived` | bool | false | Skip archived libraries (goqr) |
| `-output` | string | "./results" | Output directory for reports |
| `-format` | string | "markdown,html" | Report formats (currently only markdown) |
| `-timestamp` | bool | true | Add timestamp to report filenames |
| `-version` | bool | false | Print version and exit |

### Example Workflows

**Quick check of skip2+gozxing incompatibility**:
```bash
./bin/qr-tester \
  -data-sizes "500,550,600,650" \
  -pixel-sizes "320,400,440,450,460,480,512,560" \
  -error-levels "M"
```

**Test only gozxing encoder+decoder (should always work)**:
```bash
# Edit main.go to use only GozxingEncoder, rebuild, then:
./bin/qr-tester
```

**Production validation - test your actual use case**:
```bash
./bin/qr-tester \
  -data-sizes "750" \
  -pixel-sizes "512" \
  -error-levels "H"
```

## Output

### Report Structure

```
results/
├── skip2-go-qrcode_gozxing_20251231-120000.md
├── skip2-go-qrcode_tuotoo_20251231-120000.md
├── skip2-go-qrcode_goqr_20251231-120000.md
├── boombuler-barcode_gozxing_20251231-120000.md
└── ... (one file per encoder/decoder combination)
```

### Report Contents

Each report includes:

1. **Summary Statistics**
   - Total tests run
   - Success rate
   - Total execution time
   - Average encode/decode times

2. **Compatibility Matrix**
   - 2D grid: pixel size (columns) × data size (rows)
   - ✓ = Success, ✗ = Failure
   - Easy visual identification of problem areas

3. **Failure Analysis**
   - List of all failures with details
   - Error messages
   - Module size calculations

4. **Module Size Analysis**
   - Calculated module pixel sizes for each test
   - Fractional values highlighted
   - Helps identify root cause

5. **Timing Breakdown**
   - Encode time per test case
   - Decode time per test case
   - Identifies performance outliers

### Interpreting Results

**✓ (Success)**: Encode/decode cycle completed successfully, data matches exactly.

**✗ (Failure)**: Either:
- Decoder returned an error
- Decoder succeeded but data doesn't match original
- Decoder panicked (caught and converted to error)
- Decoder timed out

**Fractional Module Size**: When a module size is non-integer (e.g., 12.857), incompatibility is likely. Integer module sizes (e.g., 13.0) are safer.

**Non-monotonic Failures**: If 440px fails but 450px succeeds, this indicates module boundary issues, not data size problems.

## Architecture

```
qr-library-test/
├── cmd/
│   └── qr-tester/          # CLI entry point, flag parsing
│       └── main.go
├── internal/
│   ├── config/             # Configuration management
│   │   ├── config.go       # Config struct, defaults, validation
│   │   └── config_test.go
│   ├── encoders/           # Encoder wrappers (4 implementations)
│   │   ├── interface.go    # Encoder interface
│   │   ├── skip2.go        # skip2/go-qrcode wrapper
│   │   ├── boombuler.go    # boombuler/barcode wrapper
│   │   ├── yeqown.go       # yeqown/go-qrcode wrapper
│   │   ├── gozxing_encoder.go  # gozxing encoder wrapper
│   │   └── *_test.go
│   ├── decoders/           # Decoder wrappers (4 implementations)
│   │   ├── interface.go    # Decoder interface
│   │   ├── gozxing_decoder.go  # gozxing decoder wrapper
│   │   ├── tuotoo.go       # tuotoo/qrcode wrapper (panic recovery)
│   │   ├── goqr.go         # goqr wrapper (archived library)
│   │   ├── goquirc.go      # goquirc CGO wrapper
│   │   ├── goquirc_stub.go # No-op when CGO disabled
│   │   ├── registry*.go    # Decoder registration (build tags)
│   │   └── *_test.go
│   ├── testdata/           # Test case generation
│   │   ├── generator.go    # Creates test data at various sizes
│   │   ├── module_calc.go  # QR module size calculations
│   │   └── *_test.go
│   └── matrix/             # Test orchestration
│       ├── runner.go       # Executes test matrix
│       ├── result.go       # Result aggregation
│       └── *_test.go
├── pkg/
│   └── report/             # Report generation (public API)
│       ├── markdown.go     # Markdown report generator
│       └── *_test.go
├── docs/
│   └── prd-mvp.md          # Original problem analysis
├── Makefile                # Build automation
├── go.mod
├── go.sum
└── README.md
```

### Key Design Decisions

**Interfaces**: Encoder and Decoder interfaces allow uniform testing despite library differences.

**Panic Recovery**: Some decoders (tuotoo) panic on invalid input. All decoder calls are wrapped with recover() to convert panics to errors.

**Build Tags**: CGO decoders use build tags (`// +build cgo`) so the project builds without a C compiler by default.

**Module Calculation**: We calculate expected module size as `(pixelSize - 2*quietZone) / (symbolSize + 2*border)` to identify fractional values.

**Parallel Execution**: Tests run concurrently with a worker pool, but each test is independent (no shared state).

**Timeouts**: Each decoder call has a timeout to prevent hanging on problematic inputs.

## Development

### Running Tests

```bash
# Run all tests (no CGO)
make test

# Run tests with CGO
make test-cgo

# Generate coverage report
make test-coverage
# Opens coverage.html in browser
```

### Code Quality

```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run everything (format, lint, test, build)
make
```

### Adding a New Encoder

1. Create `internal/encoders/yourencoder.go`:

```go
package encoders

import "image"

type YourEncoder struct{}

func (e *YourEncoder) Name() string {
    return "yourlib"
}

func (e *YourEncoder) Encode(data []byte, opts EncodeOptions) (image.Image, error) {
    // Wrap your library's encoder
}
```

2. Add tests in `internal/encoders/yourencoder_test.go`

3. Update `getAllEncoders()` in `cmd/qr-tester/main.go`:

```go
func getAllEncoders() []encoders.Encoder {
    return []encoders.Encoder{
        &encoders.Skip2Encoder{},
        &encoders.YourEncoder{},  // Add here
    }
}
```

### Adding a New Decoder

1. Create `internal/decoders/yourdecoder.go`:

```go
package decoders

import "image"

type YourDecoder struct{}

func (d *YourDecoder) Name() string {
    return "yourlib"
}

func (d *YourDecoder) Decode(img image.Image) ([]byte, error) {
    // Wrap your library's decoder
    // Add panic recovery if needed
}
```

2. Add tests in `internal/decoders/yourdecoder_test.go`

3. Register in `internal/decoders/registry.go`:

```go
func GetAvailableDecoders(cfg *config.Config) []Decoder {
    decoders := []Decoder{
        &GozxingDecoder{},
        &YourDecoder{},  // Add here
    }
    // ...
}
```

### Build Tags for CGO

If your decoder requires CGO:

1. Add build tag to file: `// +build cgo`
2. Create stub file: `yourdecoder_stub.go` with `// +build !cgo`
3. Update `registry_cgo.go` to include your decoder

## Known Issues

### skip2 + gozxing Incompatibility

**Issue**: Fails at certain pixel sizes (440, 450, etc.) in non-monotonic pattern.

**Cause**: skip2 uses fractional module pixel sizes while gozxing assumes integer boundaries.

**Example**: At 440px with 500 bytes, skip2 creates modules of 12.857 pixels. Gozxing's decoder rounds differently than the encoder, causing misreads.

**Solution**:
- Use the same library for both encoding and decoding
- Choose pixel sizes that result in integer module sizes
- Use this tool to find safe pixel sizes for your data size

### goqr Decoder (Archived Library)

**Status**: Repository archived July 2021, read-only.

**Issues**:
- Fails on some valid QR codes
- No longer maintained
- May have unfixed bugs

**Workaround**: Use `-skip-archived=true` to exclude from tests.

**Why included**: Historical compatibility testing, shows what doesn't work.

### tuotoo Decoder Panics

**Issue**: Panics on some valid QR codes instead of returning errors.

**Handling**: Panic recovery implemented in wrapper, converts panic to error.

**Impact**: Some tests fail but program doesn't crash.

**Example error**: `"decoder panicked: runtime error: index out of range"`

### goquirc CGO Requirement

**Issue**: Requires C compiler and libquirc.

**Build**:
```bash
# Install libquirc (varies by OS)
# macOS: brew install quirc
# Ubuntu: apt-get install libquirc-dev

make build-cgo
```

**Skip**: Use `-skip-cgo=true` or build without CGO (`make build`).

**Why CGO**: goquirc wraps the C library "quirc", a fast and reliable decoder used in embedded systems.

## FAQ

**Q: Which encoder/decoder combination should I use?**

A: Use the same library for both. The gozxing library handles both encoding and decoding, ensuring compatibility.

**Q: Why do some pixel sizes fail while others succeed?**

A: Fractional module sizing. QR codes are grids of modules (black/white squares). If your pixel size doesn't divide evenly into the number of modules, you get fractional pixels per module, causing decoder errors.

**Q: How do I find safe pixel sizes for my data?**

A: Run this tool with your target data size and examine the module size analysis in the report. Choose pixel sizes that result in integer module values.

**Q: Can I test my own encoder/decoder?**

A: Yes. Implement the Encoder or Decoder interface, add it to the registry, and rebuild. See "Development" section above.

**Q: Why are tests so slow?**

A: Encoding/decoding images is computationally expensive. Use `-parallel=true` (default) and reduce test cases with `-data-sizes` and `-pixel-sizes` flags.

**Q: What's the maximum data size for QR codes?**

A: Approximately 2,953 bytes for binary data (version 40, error level L). This tool focuses on practical sizes (500-800 bytes) that fit in commonly-scanned QR codes.

## Performance

Typical test run (4 encoders × 3 decoders × 192 cases = 2,304 tests):

- **Parallel (8 CPUs)**: ~2-3 minutes
- **Serial**: ~15-20 minutes

Most time is spent in the encode/decode operations. Parallel execution provides near-linear speedup.

## License

MIT License - See LICENSE file for details.

## Credits

This tool wraps the following Go QR code libraries:

**Encoders**:
- [skip2/go-qrcode](https://github.com/skip2/go-qrcode) - Popular pure Go encoder
- [boombuler/barcode](https://github.com/boombuler/barcode) - Multi-format barcode library
- [yeqown/go-qrcode](https://github.com/yeqown/go-qrcode) - Feature-rich with styling support
- [makiuchi-d/gozxing](https://github.com/makiuchi-d/gozxing) - ZXing port (encoder/decoder)

**Decoders**:
- [makiuchi-d/gozxing](https://github.com/makiuchi-d/gozxing) - Pure Go ZXing port
- [tuotoo/qrcode](https://github.com/tuotoo/qrcode) - Pure Go with dynamic binarization
- [liyue201/goqr](https://github.com/liyue201/goqr) - Pure Go (archived)
- [kdar/goquirc](https://github.com/kdar/goquirc) - CGO wrapper for libquirc

## References

- [QR Code Standard (ISO/IEC 18004)](https://www.iso.org/standard/62021.html)
- [Original Problem Analysis](docs/prd-mvp.md)
- [ZXing Documentation](https://github.com/zxing/zxing)
- [QR Code Module Structure](https://www.thonky.com/qr-code-tutorial/)

## Contributing

Contributions welcome! Areas of interest:

1. **New encoders/decoders**: Add support for more Go libraries
2. **HTML/CSV reporters**: Implement additional output formats
3. **Multi-language support**: Test cross-language compatibility
4. **Performance optimization**: Speed up test execution
5. **Analysis tools**: Better failure pattern detection

Please open an issue first to discuss significant changes.

## Acknowledgments

Built to solve a real-world problem: applications that encode QR codes with one library but decode with another (e.g., mobile apps) can fail unpredictably. This tool identifies those incompatibilities before they reach production.

Special thanks to the maintainers of the libraries tested here for providing the Go community with QR code functionality.
