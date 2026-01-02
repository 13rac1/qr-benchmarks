// Package encoders defines the interface for QR code encoders.
package encoders

import "image"

// ErrorCorrectionLevel constants define QR code error correction levels.
// Higher levels can recover from more errors but result in larger QR codes.
const (
	ErrorCorrectionL = "L" // Low: ~7% error recovery
	ErrorCorrectionM = "M" // Medium: ~15% error recovery
	ErrorCorrectionQ = "Q" // Quartile: ~25% error recovery
	ErrorCorrectionH = "H" // High: ~30% error recovery
)

// EncodeOptions configures QR code encoding parameters.
// The zero value is not useful; PixelSize must be set.
type EncodeOptions struct {
	// ErrorCorrectionLevel determines error recovery capability.
	// Valid values: L, M, Q, H (use package constants).
	// Higher levels create larger QR codes with more redundancy.
	ErrorCorrectionLevel string

	// PixelSize is the total image dimension in pixels (width and height).
	// This value is critical for testing fractional module sizing issues.
	// The resulting module pixel size is: PixelSize / (moduleCount + quietZone).
	// When this calculation results in a fractional value, some decoder
	// libraries may fail to decode the QR code.
	PixelSize int
}

// EncodeResult contains the encoded QR code image and metadata.
type EncodeResult struct {
	// Image is the generated QR code.
	Image image.Image

	// Version is the QR code version (1-40).
	// A value of -1 indicates the version could not be determined.
	// QR version determines the size of the QR code matrix:
	// - Version 1: 21x21 modules
	// - Version 2: 25x25 modules
	// - Each version adds 4 modules per side
	// - Formula: modules = 17 + (version * 4)
	Version int
}

// Encoder generates QR codes from input data.
// Implementations wrap different QR encoding libraries to provide a uniform interface.
type Encoder interface {
	// Name returns the encoder's identifier (e.g., "skip2", "gozxing").
	// Used for reporting and result tracking.
	Name() string

	// Encode generates a QR code image from the input data.
	// Returns an error if encoding fails (e.g., data too large, invalid options).
	// The returned image dimensions should match opts.PixelSize.
	// The Version field in EncodeResult should contain the QR version (1-40),
	// or -1 if the version could not be determined.
	Encode(data []byte, opts EncodeOptions) (EncodeResult, error)

	// IsCapacityError returns true if the error indicates the data exceeds
	// QR code capacity at the requested size. These errors are valid rejections,
	// not encoder bugs, and should be treated as skipped tests.
	IsCapacityError(err error) bool
}
