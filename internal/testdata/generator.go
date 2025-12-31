// Package testdata provides test data generation for QR code compatibility testing.
package testdata

import (
	"math/rand"
	"strings"
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
}

// GeneratePixelSizeMatrix generates the primary test matrix for pixel size testing.
// This is the core test set for reproducing the skip2+gozxing fractional module issue.
//
// Matrix dimensions:
//   - Data sizes: [500, 550, 600, 650, 750, 800] bytes (6 sizes)
//   - Pixel sizes: [320, 400, 440, 450, 460, 480, 512, 560] pixels (8 sizes)
//   - Total: 6 × 8 = 48 test cases
//
// All test cases use alphanumeric content (not binary) because QR encoders convert
// data to strings, and binary data with null bytes doesn't round-trip correctly.
// Alphanumeric data is string-safe and still tests the pixel size compatibility issue.
//
// The data sizes are chosen to trigger QR versions 10-15, which are known to produce
// problematic fractional module sizes at certain pixel dimensions.
//
// Pixel sizes include:
//   - 320, 400: Common mobile resolutions (low end)
//   - 440, 450: Known problematic sizes (produce fractional modules)
//   - 460, 480: Transition zone
//   - 512, 560: Higher resolutions (power-of-2 and above)
//
// The alphanumeric data is deterministic (uses repeating pattern) for reproducible testing.
func GeneratePixelSizeMatrix() []TestCase {
	dataSizes := []int{500, 550, 600, 650, 750, 800}
	pixelSizes := []int{320, 400, 440, 450, 460, 480, 512, 560}

	cases := make([]TestCase, 0, len(dataSizes)*len(pixelSizes))

	for _, dataSize := range dataSizes {
		data := generateAlphanumeric(dataSize)
		for _, pixelSize := range pixelSizes {
			cases = append(cases, TestCase{
				Name:        formatTestName("alphanumeric", dataSize, pixelSize),
				Data:        data,
				DataSize:    dataSize,
				PixelSize:   pixelSize,
				ContentType: ContentAlphanumeric,
			})
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
// These tests use a single pixel size (480px) as they focus on content variation
// rather than pixel size variation.
func GenerateEdgeCases() []TestCase {
	// Standard pixel size for edge case testing
	pixelSize := 480

	return []TestCase{
		{
			Name:        "empty",
			Data:        []byte{},
			DataSize:    0,
			PixelSize:   pixelSize,
			ContentType: ContentBinary,
		},
		{
			Name:        "single-byte",
			Data:        []byte{0x42},
			DataSize:    1,
			PixelSize:   pixelSize,
			ContentType: ContentBinary,
		},
		{
			Name:        "numeric-small",
			Data:        generateNumeric(50),
			DataSize:    50,
			PixelSize:   pixelSize,
			ContentType: ContentNumeric,
		},
		{
			Name:        "numeric-large",
			Data:        generateNumeric(500),
			DataSize:    500,
			PixelSize:   pixelSize,
			ContentType: ContentNumeric,
		},
		{
			Name:        "alphanumeric-url",
			Data:        generateAlphanumeric(50),
			DataSize:    50,
			PixelSize:   pixelSize,
			ContentType: ContentAlphanumeric,
		},
		{
			Name:        "alphanumeric-large",
			Data:        generateAlphanumeric(1000),
			DataSize:    1000,
			PixelSize:   pixelSize,
			ContentType: ContentAlphanumeric,
		},
		{
			Name:        "utf8-multilingual",
			Data:        generateUTF8("Hello World 你好世界 Привет мир こんにちは世界"),
			DataSize:    len(generateUTF8("Hello World 你好世界 Привет мир こんにちは世界")),
			PixelSize:   pixelSize,
			ContentType: ContentUTF8,
		},
		{
			Name:        "utf8-emoji",
			Data:        generateUTF8("QR Code Testing 🔍📱✅❌🎉"),
			DataSize:    len(generateUTF8("QR Code Testing 🔍📱✅❌🎉")),
			PixelSize:   pixelSize,
			ContentType: ContentUTF8,
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

// generateUTF8 encodes a UTF-8 string as bytes.
// The content parameter should contain the exact Unicode text to encode.
//
// QR codes treat UTF-8 as binary data (no special encoding optimization).
// This is useful for testing international text and emoji support.
func generateUTF8(content string) []byte {
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
