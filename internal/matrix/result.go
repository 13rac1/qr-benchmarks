// Package matrix provides test execution and result aggregation for QR code compatibility testing.
package matrix

import "time"

// TestResult captures the outcome of a single encode→decode test cycle.
// Each test uses one encoder, one decoder, one data payload, and one pixel size.
type TestResult struct {
	// EncoderName identifies which encoder generated the QR code.
	EncoderName string

	// DecoderName identifies which decoder read the QR code.
	DecoderName string

	// DataSize is the input data length in bytes.
	DataSize int

	// PixelSize is the QR code image dimension (width and height).
	// Critical for identifying fractional module size issues.
	PixelSize int

	// QRVersion is the QR code version number (1-40).
	// Determined by data size and error correction level.
	// Version determines module count: moduleCount = 17 + 4*version.
	QRVersion int

	// ModuleCount is the number of modules (black/white squares) per side.
	// Includes data modules and function patterns, excludes quiet zone.
	ModuleCount int

	// ModulePixelSize is the calculated pixel dimension per module.
	// Computed as: PixelSize / (ModuleCount + quietZone).
	// Fractional values indicate potential decoder compatibility issues.
	ModulePixelSize float64

	// IsFractionalModule indicates whether ModulePixelSize is non-integer.
	// True when ModulePixelSize != floor(ModulePixelSize).
	// Fractional modules are a known source of decode failures.
	IsFractionalModule bool

	// EncodeTime measures encoding duration.
	EncodeTime time.Duration

	// DecodeTime measures decoding duration.
	DecodeTime time.Duration

	// Success indicates whether the decode completed without error.
	// False if decoding failed, timed out, or panicked.
	Success bool

	// Error captures the decode error if Success is false.
	// Nil if Success is true.
	Error error

	// DataMatches indicates whether decoded data equals input data.
	// Only meaningful when Success is true.
	// False indicates data corruption during encode/decode cycle.
	DataMatches bool
}

// ModuleInfo captures QR code structural metadata.
// Used to calculate module pixel sizes and identify fractional sizing issues.
type ModuleInfo struct {
	// Version is the QR code version number (1-40).
	Version int

	// ModuleCount is the number of modules per side.
	// Formula: 17 + 4*Version (e.g., version 1 = 21 modules).
	ModuleCount int

	// ModulePixelSize is the pixel dimension per module.
	// Calculated as imageSize / (ModuleCount + quietZone).
	ModulePixelSize float64

	// IsFractional indicates non-integer module pixel size.
	// True when ModulePixelSize has a fractional component.
	IsFractional bool
}

// CompatibilityMatrix aggregates test results across encoder/decoder combinations.
// Represents a multi-dimensional test matrix: encoders × decoders × data sizes × pixel sizes.
type CompatibilityMatrix struct {
	// Results contains all individual test outcomes.
	Results []TestResult

	// Encoders lists encoder names tested.
	Encoders []string

	// Decoders lists decoder names tested.
	Decoders []string

	// DataSizes lists input data sizes tested (in bytes).
	DataSizes []int

	// PixelSizes lists image dimensions tested (in pixels).
	PixelSizes []int
}

// IncompatibilityPattern identifies systematic failure patterns between encoder/decoder pairs.
// Used for analysis and reporting of known compatibility issues.
type IncompatibilityPattern struct {
	// EncoderName identifies the encoder in this failure pattern.
	EncoderName string

	// DecoderName identifies the decoder in this failure pattern.
	DecoderName string

	// FailureCount is the number of failed tests for this pair.
	FailureCount int

	// FailureRate is the percentage of failed tests (0.0-1.0).
	FailureRate float64

	// PixelSizesAffected lists pixel sizes where failures occurred.
	PixelSizesAffected []int

	// IsFractionalRelated indicates whether failures correlate with fractional module sizes.
	// True when failures occur predominantly at fractional module pixel sizes.
	IsFractionalRelated bool
}
