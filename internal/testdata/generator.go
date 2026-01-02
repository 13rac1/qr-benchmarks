// Package testdata provides test data generation for QR code compatibility testing.
package testdata

import (
	"math/rand"
	"strings"
	"unicode/utf8"
)

// ContentType identifies the character set used in test data.
// Different content types affect QR encoding efficiency and version selection.
type ContentType int

const (
	// ContentNumeric uses digits 0-9 only.
	// QR codes can encode numeric data most efficiently (3.3 bits per character).
	ContentNumeric ContentType = iota

	// ContentAlphanumeric uses A-Z, 0-9, and symbols (space $ % * + - . / :).
	// QR codes encode alphanumeric data at 5.5 bits per character.
	ContentAlphanumeric

	// ContentBinary uses arbitrary byte values including null bytes.
	// QR codes encode binary data at 8 bits per byte (least efficient).
	ContentBinary

	// ContentUTF8 uses Unicode text encoded as UTF-8.
	// QR codes treat UTF-8 as binary data (8 bits per byte).
	// Useful for testing internationalization.
	ContentUTF8
)

// TestCase represents a single test data payload with metadata.
// Each test case combines specific data content with target pixel size.
type TestCase struct {
	// Name is a human-readable identifier for this test case.
	Name string

	// Data is the payload to encode in the QR code.
	Data []byte

	// DataSize is the length of Data in bytes.
	// Stored separately for convenience in reporting.
	DataSize int

	// PixelSize is the target QR code image dimension.
	// This value is critical for testing fractional module size issues.
	PixelSize int

	// ContentType identifies the character set used in Data.
	ContentType ContentType

	// ErrorCorrectionLevel specifies the QR error correction level.
	// Valid values: "L" (Low ~7%), "M" (Medium ~15%), "Q" (Quartile ~25%), "H" (High ~30%).
	// This affects QR version selection and capacity.
	ErrorCorrectionLevel string
}

// GeneratePixelSizeMatrix generates the primary test matrix for pixel size testing.
// This is the core test set for testing fractional module sizing issues across
// focused content types and error correction levels.
//
// Matrix dimensions:
//   - Data sizes: [100, 300, 500, 750] bytes (4 sizes → QR versions 3, 7, 10, 14)
//   - Pixel sizes: [264, 270, 360, 392, 445, 462] pixels (6 sizes)
//   - Content types: 2 types (alphanumeric, UTF-8) for focused testing
//   - Error correction: 2 levels (L, H) to test low and high EC behavior
//   - Total: 4 × 6 × 2 × 2 = 96 test cases
//
// Data sizes are carefully chosen to trigger specific QR versions:
//   - 100 bytes → version 3 (29 modules)
//   - 300 bytes → version 7 (45 modules)
//   - 500 bytes → version 10 (57 modules)
//   - 750 bytes → version 14 (73 modules)
//
// Pixel sizes are chosen for a balanced mix of fractional and integer modules:
//   - 264: Integer for v3 (264/33 = 8.0)
//   - 270: Integer for v6 (270/45 = 6.0)
//   - 360: Integer for v6 (360/45 = 8.0)
//   - 392: Integer for v7 (392/49 = 8.0)
//   - 445: Integer for v17 (445/89 = 5.0)
//   - 462: Integer for v3 (462/33 = 14.0) and v14 (462/77 = 6.0)
//
// Content types:
//   - Alphanumeric: Medium efficiency (5.5 bits/char), tests encoding mode handling
//   - UTF-8: Forces byte mode, represents real-world international data
//
// Error correction levels:
//   - L (Low ~7%): Minimal redundancy, maximum capacity
//   - H (High ~30%): Maximum redundancy, tests error correction impact
//
// The data is deterministic (uses repeating patterns) for reproducible testing.
func GeneratePixelSizeMatrix() []TestCase {
	// Data sizes chosen to trigger specific QR versions
	dataSizes := []int{100, 300, 500, 750}

	// Pixel sizes chosen for balanced mix of fractional and integer modules
	pixelSizes := []int{264, 270, 360, 392, 445, 462}

	// Focused content types: alphanumeric and UTF-8
	contentTypes := []ContentType{
		ContentAlphanumeric,
		ContentUTF8,
	}

	// Test low and high error correction levels
	errorLevels := []string{"L", "H"}

	cases := make([]TestCase, 0, len(dataSizes)*len(pixelSizes)*len(contentTypes)*len(errorLevels))

	for _, dataSize := range dataSizes {
		for _, pixelSize := range pixelSizes {
			for _, contentType := range contentTypes {
				for _, ecLevel := range errorLevels {
					var data []byte
					var contentTypeStr string

					switch contentType {
					case ContentAlphanumeric:
						data = generateAlphanumeric(dataSize)
						contentTypeStr = "alphanumeric"
					case ContentUTF8:
						data = generateUTF8(dataSize)
						contentTypeStr = "utf8"
					}

					// Include EC level in test name
					name := formatTestNameWithEC(contentTypeStr, dataSize, pixelSize, ecLevel)

					cases = append(cases, TestCase{
						Name:                 name,
						Data:                 data,
						DataSize:             dataSize,
						PixelSize:            pixelSize,
						ContentType:          contentType,
						ErrorCorrectionLevel: ecLevel,
					})
				}
			}
		}
	}

	return cases
}

// GenerateComprehensiveMatrix generates an extensive test matrix for comprehensive testing.
// This test suite covers a wide range of configurations to find edge cases and determine
// the best encoder/decoder combinations across all scenarios.
//
// Matrix dimensions:
//   - Data sizes: 9 sizes from 10 to 2500 bytes (covers QR versions 1-32)
//   - Pixel sizes: 11 sizes from 128 to 1024 pixels (covers tiny to high-res)
//   - Content types: All 4 types (numeric, alphanumeric, binary, UTF-8)
//   - Error correction: 3 levels (L, M, H) for complete EC coverage
//   - Total: 9 × 11 × 4 × 3 = 1,188 test cases per encoder/decoder pair
//
// Data size progression:
//   - Tiny: 10, 25, 50 bytes (QR versions 1-2)
//   - Small: 100, 300 bytes (QR versions 3-7)
//   - Medium: 500, 1000 bytes (QR versions 10-15)
//   - Large: 2000, 2500 bytes (QR versions 25-32)
//
// Pixel size progression:
//   - Minimal: 128, 200 (edge cases, likely failures for larger versions)
//   - Small: 264, 270 (integer for common versions)
//   - Medium: 360, 392, 445, 480, 512 (mix of integer/fractional)
//   - Standard: 720 (common size)
//   - Large: 1024 (high resolution, safe for all versions)
//
// Content types tested:
//   - Numeric: Most efficient QR encoding (3.3 bits/char)
//   - Alphanumeric: Medium efficiency (5.5 bits/char)
//   - Binary: Random bytes (8 bits/byte)
//   - UTF-8: Real-world text forcing byte mode
//
// Error correction levels:
//   - L (Low ~7%): Minimal redundancy, maximum capacity
//   - M (Medium ~15%): Balanced redundancy and capacity
//   - H (High ~30%): Maximum redundancy, tests high EC impact
//
// This comprehensive test helps identify:
//   - Minimum viable pixel sizes for each data size
//   - Optimal encoder/decoder combinations
//   - Data type encoding mode issues
//   - Fractional module size boundaries
//   - Error correction level impact on version selection
//   - Maximum capacity limits across EC levels
func GenerateComprehensiveMatrix() []TestCase {
	// Comprehensive data size progression (9 sizes covering QR versions 1-32)
	// Removed 200, 700, 1500 per requirements
	dataSizes := []int{
		10,    // Tiny - QR version 1
		25,    // Tiny - QR version 1
		50,    // Small - QR version 2
		100,   // Small - QR version 3
		300,   // Medium-small - QR version 6-7
		500,   // Medium - QR version 10
		1000,  // Medium-large - QR version 15
		2000,  // Large - QR version 25
		2500,  // Very large - QR version 32 (near max at medium EC)
	}

	// Comprehensive pixel size progression (11 sizes from minimal to high-res)
	// Removed 600 per requirements
	// Includes mix of integer-producing and fractional sizes for comprehensive testing
	pixelSizes := []int{
		128,  // Minimal - will fail for larger QR versions
		200,  // Minimal - edge case testing
		264,  // Small - integer for v3 (skip2/yeqown numeric)
		270,  // Small - integer for v6 (boombuler all types)
		360,  // Medium - integer for v6 (boombuler all types)
		392,  // Medium - integer for v7 (skip2 alphanumeric)
		445,  // Medium - integer for v17 (boombuler large)
		480,  // Medium - fractional for most versions
		512,  // Medium - power of 2, fractional for most
		720,  // Standard - 720p derivative, fractional
		1024, // Large - power of 2, safe for all versions
	}

	// All four content types for comprehensive coverage
	contentTypes := []ContentType{
		ContentNumeric,
		ContentAlphanumeric,
		ContentBinary,
		ContentUTF8,
	}

	// Test low, medium, and high error correction levels
	errorLevels := []string{"L", "M", "H"}

	// Pre-allocate for all combinations: 9 sizes × 11 pixels × 4 content types × 3 EC levels
	cases := make([]TestCase, 0, len(dataSizes)*len(pixelSizes)*len(contentTypes)*len(errorLevels))

	for _, dataSize := range dataSizes {
		for _, pixelSize := range pixelSizes {
			for _, contentType := range contentTypes {
				for _, ecLevel := range errorLevels {
					var data []byte
					var contentTypeStr string

					switch contentType {
					case ContentNumeric:
						data = generateNumeric(dataSize)
						contentTypeStr = "numeric"
					case ContentAlphanumeric:
						data = generateAlphanumeric(dataSize)
						contentTypeStr = "alphanumeric"
					case ContentBinary:
						data = generateBinary(dataSize)
						contentTypeStr = "binary"
					case ContentUTF8:
						data = generateUTF8(dataSize)
						contentTypeStr = "utf8"
					}

					// Include EC level in test name
					name := formatTestNameWithEC(contentTypeStr, dataSize, pixelSize, ecLevel)

					cases = append(cases, TestCase{
						Name:                 name,
						Data:                 data,
						DataSize:             dataSize,
						PixelSize:            pixelSize,
						ContentType:          contentType,
						ErrorCorrectionLevel: ecLevel,
					})
				}
			}
		}
	}

	return cases
}

// GenerateEdgeCases generates secondary test cases for edge conditions.
// These tests verify encoder/decoder behavior with unusual inputs:
//
//   - Empty data (0 bytes)
//   - Minimal data (1 byte)
//   - Numeric content (efficient encoding)
//   - Alphanumeric content (medium efficiency)
//   - UTF-8 multilingual text (internationalization)
//   - UTF-8 with emoji (complex Unicode)
//
// These tests use a single pixel size (480px) and Medium error correction (M)
// as they focus on content variation rather than pixel size or EC variation.
func GenerateEdgeCases() []TestCase {
	// Standard pixel size for edge case testing
	pixelSize := 480

	// Use Medium error correction for all edge cases
	ecLevel := "M"

	return []TestCase{
		{
			Name:                 "empty-ecM",
			Data:                 []byte{},
			DataSize:             0,
			PixelSize:            pixelSize,
			ContentType:          ContentBinary,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "single-byte-ecM",
			Data:                 []byte{0x42},
			DataSize:             1,
			PixelSize:            pixelSize,
			ContentType:          ContentBinary,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "numeric-small-ecM",
			Data:                 generateNumeric(50),
			DataSize:             50,
			PixelSize:            pixelSize,
			ContentType:          ContentNumeric,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "numeric-large-ecM",
			Data:                 generateNumeric(500),
			DataSize:             500,
			PixelSize:            pixelSize,
			ContentType:          ContentNumeric,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "alphanumeric-url-ecM",
			Data:                 generateAlphanumeric(50),
			DataSize:             50,
			PixelSize:            pixelSize,
			ContentType:          ContentAlphanumeric,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "alphanumeric-large-ecM",
			Data:                 generateAlphanumeric(1000),
			DataSize:             1000,
			PixelSize:            pixelSize,
			ContentType:          ContentAlphanumeric,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "utf8-multilingual-ecM",
			Data:                 utf8Bytes("Hello World 你好世界 Привет мир こんにちは世界"),
			DataSize:             len(utf8Bytes("Hello World 你好世界 Привет мир こんにちは世界")),
			PixelSize:            pixelSize,
			ContentType:          ContentUTF8,
			ErrorCorrectionLevel: ecLevel,
		},
		{
			Name:                 "utf8-emoji-ecM",
			Data:                 utf8Bytes("QR Code Testing 🔍📱✅❌🎉"),
			DataSize:             len(utf8Bytes("QR Code Testing 🔍📱✅❌🎉")),
			PixelSize:            pixelSize,
			ContentType:          ContentUTF8,
			ErrorCorrectionLevel: ecLevel,
		},
	}
}

// generateNumeric creates test data containing only digits 0-9.
// The data is deterministic: repeating pattern "0123456789" up to the requested size.
//
// QR codes encode numeric data efficiently using 3.3 bits per digit,
// allowing more data in lower QR versions.
func generateNumeric(size int) []byte {
	if size <= 0 {
		return []byte{}
	}

	digits := "0123456789"
	result := make([]byte, size)

	for i := 0; i < size; i++ {
		result[i] = digits[i%len(digits)]
	}

	return result
}

// generateAlphanumeric creates test data using QR alphanumeric character set.
// Valid characters: 0-9, A-Z, and symbols (space $ % * + - . / :).
//
// The data is deterministic: repeating pattern of uppercase alphanumeric.
// QR codes encode alphanumeric data at 5.5 bits per character.
func generateAlphanumeric(size int) []byte {
	if size <= 0 {
		return []byte{}
	}

	// QR alphanumeric character set (45 characters)
	// Using a subset for simple repeating pattern
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"
	result := make([]byte, size)

	for i := 0; i < size; i++ {
		result[i] = chars[i%len(chars)]
	}

	return result
}

// generateUTF8 creates UTF-8 string data that forces byte mode encoding.
// Uses characters outside the QR alphanumeric set (multi-byte UTF-8 characters)
// which cannot be encoded in alphanumeric mode, forcing QR byte mode.
//
// The data is deterministic: repeating pattern of mixed-script text.
// This represents real-world international data and demonstrates encoding
// mode correlation with decoder behavior.
//
// Important: Truncates at valid UTF-8 character boundaries to avoid splitting
// multi-byte sequences. Actual size may be slightly less than requested.
func generateUTF8(size int) []byte {
	if size <= 0 {
		return []byte{}
	}

	// Mix of ASCII, accented characters, and CJK characters
	// These cannot be encoded in alphanumeric mode, forcing byte mode
	pattern := "Hello世界Café你好Москва"
	result := make([]byte, 0, size)

	for len(result) < size {
		result = append(result, []byte(pattern)...)
	}

	// Truncate at valid UTF-8 character boundary
	// Walk backwards until we find a valid UTF-8 sequence
	truncated := result[:size]
	for !utf8.Valid(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}

	return truncated
}

// generateBinary creates deterministic pseudo-random binary data.
// Uses a fixed seed (42) to ensure the same data is generated every time.
//
// This is critical for reproducible testing: the same data size always
// produces the same byte sequence, allowing test results to be compared
// across runs.
//
// QR codes encode binary data at 8 bits per byte (no compression).
// Binary data typically requires higher QR versions than numeric/alphanumeric
// of the same byte length.
func generateBinary(size int) []byte {
	if size <= 0 {
		return []byte{}
	}

	// Use fixed seed for deterministic output
	src := rand.NewSource(42)
	rng := rand.New(src)

	data := make([]byte, size)
	rng.Read(data)

	return data
}

// utf8Bytes encodes a UTF-8 string as bytes.
// The content parameter should contain the exact Unicode text to encode.
//
// QR codes treat UTF-8 as binary data (no special encoding optimization).
// This is useful for testing international text and emoji support.
func utf8Bytes(content string) []byte {
	return []byte(content)
}

// formatTestName creates a consistent test case identifier.
// Format: "content-type-NNNb-NNNpx"
//
// Examples:
//   - "binary-500b-440px"
//   - "numeric-50b-480px"
//   - "utf8-100b-512px"
func formatTestName(contentType string, dataSize, pixelSize int) string {
	var sb strings.Builder
	sb.WriteString(contentType)
	sb.WriteString("-")
	sb.WriteString(formatInt(dataSize))
	sb.WriteString("b-")
	sb.WriteString(formatInt(pixelSize))
	sb.WriteString("px")
	return sb.String()
}

// formatTestNameWithEC creates a test case identifier including error correction level.
// Format: "content-type-NNNb-NNNpx-ecX"
//
// Examples:
//   - "alphanumeric-500b-440px-ecL"
//   - "utf8-100b-512px-ecH"
//   - "binary-750b-480px-ecM"
func formatTestNameWithEC(contentType string, dataSize, pixelSize int, ecLevel string) string {
	var sb strings.Builder
	sb.WriteString(contentType)
	sb.WriteString("-")
	sb.WriteString(formatInt(dataSize))
	sb.WriteString("b-")
	sb.WriteString(formatInt(pixelSize))
	sb.WriteString("px-ec")
	sb.WriteString(ecLevel)
	return sb.String()
}

// formatInt converts an integer to a string without allocations (via strings.Builder).
// This is a simple helper to avoid fmt.Sprintf overhead in test name generation.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	// Handle negative numbers (unlikely in our use case, but for completeness)
	negative := n < 0
	if negative {
		n = -n
	}

	// Count digits
	digits := 0
	temp := n
	for temp > 0 {
		digits++
		temp /= 10
	}

	// Build string
	result := make([]byte, digits)
	for i := digits - 1; i >= 0; i-- {
		result[i] = byte('0' + n%10)
		n /= 10
	}

	if negative {
		return "-" + string(result)
	}
	return string(result)
}
