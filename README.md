# QR Code Library Compatibility Benchmark

Systematic benchmarking tool for identifying QR encoder/decoder incompatibilities, performance characteristics, and capacity limits across Go libraries.

## The Problem

Different QR code libraries handle encoding/decoding differently:

1. **Fractional module sizing**: Some encoders create non-integer pixels-per-module, causing decoder failures
2. **Capacity handling**: Libraries differ in how they report data exceeding QR capacity
3. **UTF-8 handling**: String encoding can add metadata bytes or replace invalid sequences
4. **Performance**: Encode/decode speeds vary significantly between libraries

This tool systematically tests all encoder/decoder combinations to identify incompatibilities and measure performance.

## Quick Start

```bash
# Build (with CGO for goquirc decoder)
make build

# Run standard test matrix (96 tests per encoder/decoder pair)
make run

# Run comprehensive tests (576 tests per pair)
make run-full

# Generate Hugo website from results
make generate-site

# Preview website locally
make serve-site
```

## Features

- **4 Encoders**: skip2/go-qrcode, boombuler/barcode, yeqown/go-qrcode, makiuchi-d/gozxing
- **4 Decoders**: makiuchi-d/gozxing, tuotoo/qrcode, liyue201/goqr, kdar/goquirc (CGO)
- **Test Modes**:
  - Standard: 4 data sizes × 4 content types × 6 pixel sizes = 96 tests per pair
  - Comprehensive: 12 data sizes × 4 content types × 12 pixel sizes = 576 tests per pair
- **Content Types**: Numeric, alphanumeric, binary, UTF-8
- **Capacity Tracking**: Separates QR capacity exceeded from encoder bugs
- **JSON Output**: Raw test results split by encoder and decoder
- **Hugo Website**: Interactive benchmark results with sortable tables and charts

## Installation

### Prerequisites

- Go 1.20 or later
- C compiler (Xcode CLI tools on macOS, gcc on Linux)

### Build Options

**Default build (4 decoders, includes goquirc via CGO)**:
```bash
make build
# Creates: bin/qr-tester
```

**Without CGO (3 decoders, portable)**:
```bash
make build-nocgo
# Creates: bin/qr-tester-nocgo
```

### Install Dependencies

```bash
make deps
# Or: go mod download
```

## Usage

### Running Tests

**Standard test mode** (4 data sizes × 4 content types × 6 pixel sizes):
```bash
./bin/qr-tester
# or
make run
```

**Comprehensive test mode** (12 data sizes × 4 content types × 12 pixel sizes):
```bash
./bin/qr-tester -test-mode=comprehensive
# or
make run-full
```

### Test Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-test-mode` | `standard` | Test mode: `standard` or `comprehensive` |
| `-output-dir` | `./results` | Output directory for JSON results |

**Standard mode**:
- Data sizes: 10, 25, 50, 100 bytes
- Pixel sizes: 128, 200, 256, 320, 400, 512
- Content types: numeric, alphanumeric, binary, utf8

**Comprehensive mode**:
- Data sizes: 10, 25, 50, 100, 200, 300, 500, 700, 1000, 1500, 2000, 2500 bytes
- Pixel sizes: 128, 200, 256, 320, 400, 440, 450, 460, 480, 512, 560, 600, 720, 800, 1024
- Content types: numeric, alphanumeric, binary, utf8

## Output

### JSON Results Structure

```
results/
├── encoders/
│   ├── skip2_go-qrcode.json
│   ├── boombuler_barcode.json
│   ├── yeqown_go-qrcode.json
│   └── makiuchi-d_gozxing.json
└── decoders/
    ├── makiuchi-d_gozxing.json
    ├── tuotoo_qrcode.json
    ├── liyue201_goqr.json
    └── kdar_goquirc.json
```

Each JSON file contains:
```json
{
  "timestamp": "2026-01-01T12:00:00Z",
  "results": [
    {
      "encoder": "skip2/go-qrcode",
      "decoder": "makiuchi-d/gozxing",
      "dataSize": 100,
      "pixelSize": 256,
      "contentType": "binary",
      "success": true,
      "isCapacityExceeded": false,
      "encodeTimeMs": 1.234,
      "decodeTimeMs": 0.567,
      "qrVersion": 2,
      "moduleCount": 25,
      "modulePixelSize": 10.24,
      "isFractionalModule": true
    }
  ]
}
```

### Generating Website

Generate Hugo static site from JSON results:
```bash
make generate-site   # Creates website/data/*.json
make build-site      # Builds static HTML in website/public/
make serve-site      # Preview at http://localhost:1313
```

### Interpreting Results

**Success/Failure**:
- `success: true` - Encode/decode cycle completed, data matches exactly
- `success: false` - Failure with error type:
  - `encode` - Encoding failed (check `isCapacityExceeded`)
  - `decode` - Decoder returned error or panicked
  - `dataMismatch` - Decoded data doesn't match original

**Capacity Exceeded** (`isCapacityExceeded: true`):
- Encoder correctly reported data exceeds QR capacity
- Not counted as failure - it's a valid rejection
- Success rate = successes / (total - capacitySkips)

**Module Analysis**:
- `isFractionalModule: true` - Non-integer pixels per module (e.g., 10.24)
- Fractional modules often cause decoder failures
- Integer module sizes are more reliable

## Architecture

### Core Components

- **`cmd/qr-tester`** - CLI entry point with test mode flags
- **`internal/encoders`** - 4 encoder wrappers with unified interface
- **`internal/decoders`** - 4 decoder wrappers with panic recovery
- **`internal/testdata`** - Test data generation (numeric, alphanumeric, binary, UTF-8)
- **`internal/matrix`** - Test execution and result aggregation
- **`pkg/report`** - JSON output generation split by encoder/decoder
- **`cmd/generate-site`** - Converts JSON to Hugo data format
- **`website/`** - Hugo static site for interactive results

### Key Design Decisions

**Capacity vs Errors**: Encoders implement `IsCapacityError()` to distinguish valid capacity rejections from bugs

**UTF-8 Handling**: Test data generator ensures UTF-8 doesn't split multi-byte characters at boundaries

**Panic Recovery**: Decoders that panic (tuotoo) are wrapped with recover() to convert panics to errors

**CGO Support**: goquirc decoder requires C compiler; project builds without CGO using build tags

## Development

### Running Tests

```bash
# Run all tests (with CGO)
make test

# Run tests without CGO
make test-nocgo

# Generate coverage report
make test-coverage
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

### Adding New Libraries

See existing encoder/decoder wrappers in `internal/encoders/` and `internal/decoders/` for implementation patterns.

Key requirements:
- Implement `Encoder` or `Decoder` interface
- Add `IsCapacityError()` method to distinguish capacity limits from bugs
- Wrap panics in decoders with `recover()` and return as errors
- Use build tags for CGO dependencies

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

**Requirement**: C compiler (Xcode CLI tools on macOS, gcc on Linux).

**Build**: libquirc is vendored and built automatically by `make build`.

**Skip**: Use `-skip-cgo=true` at runtime or `make build-nocgo` to build without CGO.

**Why CGO**: goquirc wraps the C library "quirc", a fast and reliable decoder used in embedded systems.

### chai2010/qrcode (Removed)

**Status**: Removed from this project.

**Reason**: The library vendors OpenCV 1.1.0 (circa 2010) which does not compile on modern macOS with current clang. The vendored C++ code has numerous compatibility issues including:
- Pointer vs integer comparisons (`npts <= 0` where npts is a pointer)
- Integer narrowing errors (`0x80000000` in int arrays)
- Missing `struct` keywords for POSIX types
- Taking addresses of temporary objects

While the library works on Linux, the maintenance burden to patch the vendored OpenCV code for macOS cross-platform support was not justified given that we already have 4 working pure-Go encoders.

### KangSpace/gqrcode (Removed)

**Status**: Removed from this project.

**Reason**: The library silently truncates data that exceeds QR code capacity limits instead of returning an error. When encoding 500 bytes of data, the library logs a warning but proceeds to create a QR code containing only 20 bytes:

```
[WARN] buildCharacterCountIndicator: Data length is too long, it will be trim to max count by rule "Table3 - Number of bits in character count indicator for QR Code", max count is :20
```

This silent data corruption makes the library unsuitable for production use. Other encoders like skip2/go-qrcode properly return an error (`content too long to encode`) when data exceeds capacity.

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
