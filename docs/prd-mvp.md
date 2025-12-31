## Golang QR Code Libraries

### Encoders

- skip2/go-qrcode (github.com/skip2/go-qrcode) – Pure Go QR code encoder with support for four error correction levels. Most popular and widely used encoder library.[1][2]

- boombuler/barcode (github.com/boombuler/barcode) – Multi-format barcode library including QR code generation. Supports multiple barcode types beyond QR codes.[3][4]

- yeqown/go-qrcode (github.com/yeqown/go-qrcode) – Feature-rich QR encoder with customizable styles (colors, shapes, icons, borders). Supports all QR versions 1-40 with automatic version detection.[5][6]

- gozxing (github.com/makiuchi-d/gozxing) – Port of ZXing supporting both encoding and decoding. Multi-format support for various barcode types.[7]

### Decoders

- gozxing (github.com/makiuchi-d/gozxing) – Pure Go port of ZXing for multi-format barcode decoding. Most comprehensive decoder option.[7]

- tuotoo/qrcode (github.com/tuotoo/qrcode) – Pure Go QR-specific decoder with dynamic binarization. Handles various angles and error correction modes.[8]

- liyue201/goqr (github.com/liyue201/goqr) – Pure Go QR decoder. Note: Archived in July 2021, read-only.[9]

- goquirc (github.com/kdar/goquirc) – CGO wrapper around Quirc C library. Requires CGO and C compiler.[10]

- goBarcodeQrSDK (github.com/yushulx/goBarcodeQrSDK) – Commercial CGO wrapper for Dynamsoft Barcode Reader. Requires license for production.[11]

## Application Design: QR Compatibility Matrix Tester

Here's a modular architecture design for testing all encoder/decoder combinations across data sizes:

### Architecture Overview

```go
// Core interfaces for extensibility
type Encoder interface {
    Name() string
    Encode(data []byte, options EncodeOptions) (image.Image, error)
}

type Decoder interface {
    Name() string
    Decode(img image.Image) ([]byte, error)
}

type EncodeOptions struct {
    ErrorCorrectionLevel string
    Size                 int
}

// Test result tracking
type TestResult struct {
    EncoderName    string
    DecoderName    string
    DataSize       int
    Success        bool
    Error          error
    EncodeTime     time.Duration
    DecodeTime     time.Duration
    DataMatches    bool
}

type CompatibilityMatrix struct {
    Results   []TestResult
    Encoders  []string
    Decoders  []string
    DataSizes []int
}
```

### Directory Structure

```
qr-test-matrix/
├── cmd/
│   └── qr-tester/
│       └── main.go
├── internal/
│   ├── encoders/
│   │   ├── interface.go
│   │   ├── skip2.go
│   │   ├── boombuler.go
│   │   ├── yeqown.go
│   │   └── gozxing_encoder.go
│   ├── decoders/
│   │   ├── interface.go
│   │   ├── gozxing_decoder.go
│   │   ├── tuotoo.go
│   │   ├── goqr.go
│   │   ├── goquirc.go
│   │   └── dynamsoft.go (optional if licensed)
│   ├── testdata/
│   │   └── generator.go
│   ├── matrix/
│   │   ├── runner.go
│   │   ├── result.go
│   │   └── reporter.go
│   └── config/
│       └── config.go
├── pkg/
│   └── report/
│       ├── html.go
│       ├── csv.go
│       └── markdown.go
├── go.mod
└── README.md
```

### Key Components

#### 1. Test Data Generator
```go
type DataGenerator struct {
    Sizes []int // byte sizes to test
}

func (g *DataGenerator) Generate() [][]byte {
    // Generate test data at various sizes:
    // - Small: 10, 50, 100 bytes
    // - Medium: 500, 1KB, 2KB
    // - Large: 4KB (max for QR)
    // - Edge cases: empty, single byte, max capacity
    // - Content types: numeric, alphanumeric, binary, UTF-8
}
```

#### 2. Test Runner
```go
type MatrixRunner struct {
    Encoders  []Encoder
    Decoders  []Decoder
    TestData  [][]byte
    Parallel  bool
    Timeout   time.Duration
}

func (r *MatrixRunner) RunAll() *CompatibilityMatrix {
    // For each test data:
    //   For each encoder:
    //     Encode data
    //     For each decoder:
    //       Decode image
    //       Compare original vs decoded
    //       Record result with timing
}
```

#### 3. Result Reporter
```go
type Reporter interface {
    Generate(matrix *CompatibilityMatrix) error
}

// Implementations:
// - HTMLReporter: Interactive table with color coding
// - CSVReporter: Raw data for analysis
// - MarkdownReporter: GitHub-friendly reports
```

### Implementation Strategy

#### Phase 1: Core Infrastructure
1. Define interfaces and base types
2. Implement test data generator with various sizes and content types
3. Create test runner with proper error handling and timeouts
4. Build basic CSV reporter

#### Phase 2: Encoder Integration
1. Implement wrappers for each encoder library
2. Normalize error correction levels across libraries
3. Handle library-specific quirks and options
4. Add unit tests for each wrapper

#### Phase 3: Decoder Integration
1. Implement wrappers for each decoder library
2. Handle CGO dependencies (goquirc, goBarcodeQrSDK)
3. Normalize error responses
4. Add timeout protection for slow decoders

#### Phase 4: Matrix Testing
1. Run full compatibility matrix
2. Generate reports in multiple formats
3. Identify failure patterns
4. Document known incompatibilities

#### Phase 5: Multi-Language Extension
```go
// Future: Add support for other languages
type ExternalEncoder struct {
    Language string // "python", "rust", "typescript"
    Command  string
    Args     []string
}

// Run external processes and parse results
```

### Configuration Example

```go
type Config struct {
    DataSizes    []int
    ErrorLevels  []string // "L", "M", "Q", "H"
    Parallel     bool
    Timeout      time.Duration
    OutputFormat string // "html", "csv", "markdown", "all"
    SkipCGO      bool   // Skip CGO-based libraries if true
}
```

### Usage Example

```go
func main() {
    cfg := LoadConfig()
    
    // Initialize encoders
    encoders := []Encoder{
        encoders.NewSkip2Encoder(),
        encoders.NewBoombulerEncoder(),
        encoders.NewYeqownEncoder(),
        encoders.NewGozxingEncoder(),
    }
    
    // Initialize decoders
    decoders := []Decoder{
        decoders.NewGozxingDecoder(),
        decoders.NewTuotooDecoder(),
        decoders.NewGoqrDecoder(),
    }
    
    if !cfg.SkipCGO {
        decoders = append(decoders, decoders.NewGoquircDecoder())
    }
    
    // Generate test data
    generator := testdata.NewGenerator(cfg.DataSizes)
    testData := generator.GenerateAll()
    
    // Run matrix tests
    runner := matrix.NewRunner(encoders, decoders, testData)
    runner.SetParallel(cfg.Parallel)
    runner.SetTimeout(cfg.Timeout)
    
    results := runner.RunAll()
    
    // Generate reports
    reporter := report.NewMultiReporter(cfg.OutputFormat)
    reporter.Generate(results)
}
```

### Benefits of This Design

This architecture aligns with your design principles:[10]

1. **Modularity**: Clean interfaces allow independent encoder/decoder implementations
2. **Robustness**: Timeout protection, error handling, and edge case testing
3. **Extensibility**: Easy to add new libraries or languages via interfaces
4. **Maintainability**: Separation of concerns with clear package boundaries
5. **Testing**: Built-in validation framework ensures reliability
6. **Portability**: CGO dependencies are optional and configurable
7. **Documentation**: Self-documenting via interfaces and generated reports

The matrix runner can execute tests in parallel for performance, with proper timeout handling to prevent hanging on problematic decoder implementations.[1][8][7]

[1](https://github.com/skip2/go-qrcode)
[2](https://pkg.go.dev/github.com/skip2/go-qrcode)
[3](https://github.com/boombuler/barcode)
[4](https://pkg.go.dev/github.com/boombuler/barcode/qr)
[5](https://github.com/yeqown/go-qrcode)
[6](https://pkg.go.dev/github.com/yeqown/go-qrcode)
[7](https://github.com/makiuchi-d/gozxing)
[8](https://github.com/tuotoo/qrcode)
[9](https://github.com/liyue201/goqr)
[10](https://github.com/kdar/goquirc)
[11](https://github.com/yushulx/goBarcodeQrSDK)
[12](https://pkg.go.dev/github.com/yougg/go-qrcode)
[13](https://www.twilio.com/en-us/blog/developers/tutorials/building-blocks/generate-qr-code-with-go)
[14](https://pkg.go.dev/github.com/dmitrymomot/gokit/qrcode)
[15](https://github.com/skip2/go-qrcode/blob/master/encoder.go)
[16](https://packages.debian.org/sid/i386/devel/go-qrcode)
[17](https://www.linkedin.com/pulse/generating-barcodes-qr-codes-golang-kazi-ashik-qc1uc)
[18](https://git.linux.ucla.edu/di-gital/matterbridge/src/commit/cc05ba89073e16dd374f5b5a77cb1214a2e735b9/vendor/github.com/skip2/go-qrcode)
[19](https://pkg.go.dev/github.com/boombuler/barcode)
[20](https://ftp.guvi.in/hub/go-language-tutorial/generate-barcode-with-golang/)
[21](https://github.com/stokito/go-qrcode-libs-compare)
[22](https://www.reddit.com/r/golang/comments/1lsx7cn/multiple_barcodes_can_be_generated_on_a_single/)
[23](https://github.com/make-github-pseudonymous-again/awesome-qr-code)
[24](https://githubhelp.com/boombuler/barcode)
[25](https://pkg.go.dev/github.com/makiuchi-d/gozxing)
[26](https://pkg.go.dev/github.com/tuotoo/qrcode)
[27](https://pkg.go.dev/github.com/piglig/go-qr)
