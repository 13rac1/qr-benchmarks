// Package matrix provides test execution and result aggregation.
package matrix

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/13rac1/qr-library-test/internal/config"
	"github.com/13rac1/qr-library-test/internal/decoders"
	"github.com/13rac1/qr-library-test/internal/encoders"
	"github.com/13rac1/qr-library-test/internal/testdata"
)

// Runner orchestrates QR code compatibility matrix testing.
// It executes encode→decode→validate cycles for all combinations of:
// encoders × decoders × test cases.
type Runner struct {
	Encoders  []encoders.Encoder
	Decoders  []decoders.Decoder
	TestCases []testdata.TestCase
	Config    *config.Config
}

// NewRunner creates a test runner with the provided components.
func NewRunner(cfg *config.Config, encs []encoders.Encoder, decs []decoders.Decoder, cases []testdata.TestCase) *Runner {
	return &Runner{
		Encoders:  encs,
		Decoders:  decs,
		TestCases: cases,
		Config:    cfg,
	}
}

// RunAll executes the complete test matrix and returns aggregated results.
// For each test case, it runs encoding with each encoder, then decoding with each decoder.
// This is currently single-threaded; parallel execution will be added in commit 9.
func (r *Runner) RunAll() (*CompatibilityMatrix, error) {
	if len(r.Encoders) == 0 {
		return nil, fmt.Errorf("no encoders provided")
	}
	if len(r.Decoders) == 0 {
		return nil, fmt.Errorf("no decoders provided")
	}
	if len(r.TestCases) == 0 {
		return nil, fmt.Errorf("no test cases provided")
	}

	// Calculate total number of tests
	totalTests := len(r.Encoders) * len(r.Decoders) * len(r.TestCases)
	results := make([]TestResult, 0, totalTests)

	// Collect unique data sizes and pixel sizes for matrix metadata
	dataSizeMap := make(map[int]bool)
	pixelSizeMap := make(map[int]bool)
	encoderNames := make([]string, len(r.Encoders))
	decoderNames := make([]string, len(r.Decoders))

	for i, enc := range r.Encoders {
		encoderNames[i] = enc.Name()
	}
	for i, dec := range r.Decoders {
		decoderNames[i] = dec.Name()
	}

	// Run all test combinations
	testNum := 0
	for _, testCase := range r.TestCases {
		dataSizeMap[testCase.DataSize] = true
		pixelSizeMap[testCase.PixelSize] = true

		for _, encoder := range r.Encoders {
			for _, decoder := range r.Decoders {
				testNum++
				result := r.runTest(testCase, encoder, decoder)
				results = append(results, result)

				// Print progress
				r.printProgress(testNum, totalTests, testCase, encoder, decoder, result)
			}
		}
	}

	// Convert maps to sorted slices
	dataSizes := make([]int, 0, len(dataSizeMap))
	for size := range dataSizeMap {
		dataSizes = append(dataSizes, size)
	}

	pixelSizes := make([]int, 0, len(pixelSizeMap))
	for size := range pixelSizeMap {
		pixelSizes = append(pixelSizes, size)
	}

	return &CompatibilityMatrix{
		Results:    results,
		Encoders:   encoderNames,
		Decoders:   decoderNames,
		DataSizes:  dataSizes,
		PixelSizes: pixelSizes,
	}, nil
}

// runTest executes a single encode→decode→validate cycle.
// Returns a TestResult capturing timing, success status, and module information.
func (r *Runner) runTest(testCase testdata.TestCase, enc encoders.Encoder, dec decoders.Decoder) TestResult {
	result := TestResult{
		EncoderName: enc.Name(),
		DecoderName: dec.Name(),
		DataSize:    testCase.DataSize,
		PixelSize:   testCase.PixelSize,
		ContentType: contentTypeToString(testCase.ContentType),
		QRVersion:   -1, // Will be updated if version detection succeeds
		ModuleCount: 0,  // Will be updated if version detection succeeds
	}

	// Encode QR code with timing
	encodeOpts := encoders.EncodeOptions{
		ErrorCorrectionLevel: encoders.ErrorCorrectionM,
		PixelSize:            testCase.PixelSize,
	}

	encodeStart := time.Now()
	img, err := enc.Encode(testCase.Data, encodeOpts)
	result.EncodeTime = time.Since(encodeStart)

	if err != nil {
		result.Error = EncodeError{Err: err}
		result.IsCapacityExceeded = enc.IsCapacityError(err)
		return result
	}

	// Attempt to detect QR version
	// This will fail with "not yet implemented" error, which is expected
	version, err := testdata.DetectQRVersion(img)
	if err == nil && version > 0 {
		result.QRVersion = version
		result.ModuleCount = testdata.CalculateModuleCount(version)

		// Calculate module pixel size
		modulePixelSize := testdata.CalculateModulePixelSize(testCase.PixelSize, result.ModuleCount, testdata.QuietZoneModules)
		result.ModulePixelSize = modulePixelSize
		result.IsFractionalModule = testdata.IsFractionalModuleSize(modulePixelSize)
	}
	// If version detection fails (expected for now), leave QRVersion as -1
	// and ModuleCount/ModulePixelSize as 0

	// Decode QR code with timing
	decodeStart := time.Now()
	decodedData, err := dec.Decode(img)
	result.DecodeTime = time.Since(decodeStart)

	if err != nil {
		result.Error = DecodeError{Err: err}
		return result
	}

	// Validate decoded data matches original
	if !bytes.Equal(testCase.Data, decodedData) {
		result.Error = DataMismatchError{
			Expected: len(testCase.Data),
			Got:      len(decodedData),
		}
	} else {
		result.Error = nil
	}

	return result
}

// printProgress outputs real-time test progress to stdout.
// Shows test number, status (✓/✗), data type, dimensions, encoder, and timing.
func (r *Runner) printProgress(testNum, totalTests int, testCase testdata.TestCase, enc encoders.Encoder, dec decoders.Decoder, result TestResult) {
	// Determine status symbol and color based on error type
	status := "✓"
	statusColor := "\033[32m" // Green

	if result.Error != nil {
		// Set status based on error type
		var encErr EncodeError
		var decErr DecodeError
		var dataErr DataMismatchError

		if errors.As(result.Error, &encErr) {
			if result.IsCapacityExceeded {
				status = "⊘ (skip)"
				statusColor = "\033[33m" // Yellow
			} else {
				status = "✗ (encode)"
				statusColor = "\033[31m" // Red
			}
		} else if errors.As(result.Error, &decErr) {
			status = "✗ (decode)"
			statusColor = "\033[31m" // Red
		} else if errors.As(result.Error, &dataErr) {
			status = "✗ (data)"
			statusColor = "\033[31m" // Red
		} else {
			status = "✗"
			statusColor = "\033[31m" // Red
		}
	}
	reset := "\033[0m"

	// Content type label
	contentLabel := contentTypeToString(testCase.ContentType)

	// Print test result
	fmt.Printf("[%d/%d] %s%s%s %s %d bytes @ %dpx (%s+%s) - %.1fms encode, %.1fms decode\n",
		testNum, totalTests,
		statusColor, status, reset,
		contentLabel,
		testCase.DataSize,
		testCase.PixelSize,
		enc.Name(),
		dec.Name(),
		float64(result.EncodeTime.Microseconds())/1000.0,
		float64(result.DecodeTime.Microseconds())/1000.0,
	)

	// Print error details if failed
	if result.Error != nil {
		fmt.Printf("  └─ %s\n", result.Error)
	}
}

// contentTypeToString converts ContentType to display string.
func contentTypeToString(ct testdata.ContentType) string {
	switch ct {
	case testdata.ContentNumeric:
		return "numeric"
	case testdata.ContentAlphanumeric:
		return "alphanumeric"
	case testdata.ContentBinary:
		return "binary"
	case testdata.ContentUTF8:
		return "utf8"
	default:
		return "unknown"
	}
}
