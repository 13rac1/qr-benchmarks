# QR Code Decoders

This package provides multiple QR code decoder implementations for compatibility testing.

## Decoder Overview

| Decoder | Type | Build Requirements | Status |
|---------|------|-------------------|---------|
| **gozxing** | Pure Go | None | Active |
| **tuotoo** | Pure Go | None | Active |
| **goqr** | Pure Go | None | Archived (July 2021) |
| **goquirc** | CGO | C compiler + libquirc | Active |

## Pure Go Decoders

### gozxing
- **Package**: `github.com/makiuchi-d/gozxing`
- **Build**: Always available
- **Notes**: Port of ZXing (Zebra Crossing) barcode library

### tuotoo
- **Package**: `github.com/tuotoo/qrcode`
- **Build**: Always available
- **Notes**: Pure Go implementation with dynamic binarization

### goqr
- **Package**: `github.com/liyue201/goqr`
- **Build**: Always available by default
- **Notes**: Archived library (last update July 2021)
- **Configuration**: Use `--skip-archived` flag to exclude

## CGO Decoder (goquirc)

The **goquirc** decoder requires CGO and a C compiler.

### Requirements

- C compiler (gcc, clang, etc.)
- CGO enabled at build time
- libquirc C library (automatically included via cgo)

### Building with CGO

To build with the goquirc decoder:

```bash
# Enable CGO (required)
CGO_ENABLED=1 go build ./...

# Or explicitly with cgo tag
CGO_ENABLED=1 go build -tags cgo ./...

# Run tests with CGO
CGO_ENABLED=1 go test -tags cgo ./internal/decoders/...
```

### Building without CGO

To build without the goquirc decoder (default on most systems):

```bash
# Disable CGO (default behavior)
CGO_ENABLED=0 go build ./...

# The goquirc decoder will not be included
```

### Runtime Behavior

The decoder registry automatically adapts based on build configuration:

- **Without CGO**: `GetAvailableDecoders()` returns 3 decoders (gozxing, tuotoo, goqr)
- **With CGO**: `GetAvailableDecoders()` returns 4 decoders (adds goquirc)
- **Skip CGO flag**: Use `--skip-cgo` to exclude goquirc even if built with CGO enabled

### Build Tags

The goquirc implementation uses build tags to ensure it only compiles when CGO is available:

```go
//go:build cgo
// +build cgo
```

This prevents compilation errors when CGO is not available.

## Configuration Options

### Skip Archived Libraries

```bash
# Exclude archived libraries (goqr) from testing
./qr-tester --skip-archived
```

### Skip CGO Libraries

```bash
# Exclude CGO-based libraries (goquirc) from testing
./qr-tester --skip-cgo
```

### Skip Both

```bash
# Use only actively maintained pure Go decoders (gozxing, tuotoo)
./qr-tester --skip-archived --skip-cgo
```

## Usage Example

```go
package main

import (
    "github.com/13rac1/qr-library-test/internal/config"
    "github.com/13rac1/qr-library-test/internal/decoders"
)

func main() {
    cfg := config.DefaultConfig()

    // Get available decoders based on configuration
    decoders := decoders.GetAvailableDecoders(cfg)

    // Available decoders depend on:
    // 1. Build configuration (CGO enabled/disabled)
    // 2. Runtime configuration (skip flags)

    for _, dec := range decoders {
        println("Decoder available:", dec.Name())
    }
}
```

## Testing

### Test without CGO (default)

```bash
# Tests run with pure Go decoders only
CGO_ENABLED=0 go test ./internal/decoders/...
```

### Test with CGO

```bash
# Tests include goquirc decoder
CGO_ENABLED=1 go test -tags cgo ./internal/decoders/...
```

## Implementation Details

### Build Tag Pattern

The package uses a clean build tag pattern for CGO detection:

**`registry_nocgo.go`** (default):
```go
//go:build !cgo

func cgoEnabled() bool {
    return false
}
```

**`registry_cgo.go`** (when CGO enabled):
```go
//go:build cgo

func cgoEnabled() bool {
    return true
}
```

This ensures the registry automatically returns the correct value based on build configuration.

### Error Handling

All decoders implement consistent error handling:

- Panic recovery with defer/recover
- Nil image checking
- Wrapped errors with decoder name prefix
- Consistent error format: `"decoder_name: error message"`

### Compatibility Testing

The decoder implementations are designed to test compatibility across different libraries:

- **Pixel size variations**: Test how decoders handle different image dimensions
- **Module size handling**: Fractional vs integer module sizing
- **Error correction levels**: Support for L, M, Q, H levels
- **Data sizes**: Small to large data payloads

## Known Issues

### goquirc CGO Dependencies

The goquirc library requires CGO and will fail to build if:
- CGO is disabled (`CGO_ENABLED=0`)
- No C compiler is available
- Cross-compilation without CGO support

**Solution**: Use `--skip-cgo` flag or build without CGO to exclude goquirc.

### goqr Archived Status

The goqr library is archived (last update July 2021) and may have compatibility issues with newer Go versions or dependencies.

**Solution**: Use `--skip-archived` flag to exclude goqr from testing.

## Future Decoders

Additional decoders can be added by:

1. Implementing the `Decoder` interface in a new file
2. Adding appropriate build tags if CGO or other requirements apply
3. Registering in `registry.go` with appropriate conditional logic
4. Adding comprehensive tests
5. Documenting in this README

## References

- gozxing: https://github.com/makiuchi-d/gozxing
- tuotoo: https://github.com/tuotoo/qrcode
- goqr: https://github.com/liyue201/goqr (archived)
- goquirc: https://github.com/kdar/goquirc
- Quirc C library: https://github.com/dlbeer/quirc
