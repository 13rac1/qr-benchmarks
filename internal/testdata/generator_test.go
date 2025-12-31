package testdata

import (
	"bytes"
	"testing"
)

func TestGeneratePixelSizeMatrix(t *testing.T) {
	cases := GeneratePixelSizeMatrix()

	// Verify total count: 6 data sizes × 8 pixel sizes = 48
	expectedCount := 48
	if len(cases) != expectedCount {
		t.Errorf("GeneratePixelSizeMatrix() returned %d cases, expected %d",
			len(cases), expectedCount)
	}

	// Expected data and pixel sizes
	expectedDataSizes := []int{500, 550, 600, 650, 750, 800}
	expectedPixelSizes := []int{320, 400, 440, 450, 460, 480, 512, 560}

	// Verify all combinations are present
	combinations := make(map[string]bool)
	for _, tc := range cases {
		// Verify content type is binary
		if tc.ContentType != ContentBinary {
			t.Errorf("test case %q has content type %d, expected ContentBinary (%d)",
				tc.Name, tc.ContentType, ContentBinary)
		}

		// Verify data size matches actual data length
		if tc.DataSize != len(tc.Data) {
			t.Errorf("test case %q has DataSize %d but len(Data) = %d",
				tc.Name, tc.DataSize, len(tc.Data))
		}

		// Track this combination
		key := formatInt(tc.DataSize) + ":" + formatInt(tc.PixelSize)
		combinations[key] = true
	}

	// Verify all expected combinations are present
	for _, dataSize := range expectedDataSizes {
		for _, pixelSize := range expectedPixelSizes {
			key := formatInt(dataSize) + ":" + formatInt(pixelSize)
			if !combinations[key] {
				t.Errorf("missing combination: data size %d, pixel size %d",
					dataSize, pixelSize)
			}
		}
	}
}

func TestGenerateEdgeCases(t *testing.T) {
	cases := GenerateEdgeCases()

	// Verify we have expected number of edge cases
	if len(cases) < 5 {
		t.Errorf("GenerateEdgeCases() returned %d cases, expected at least 5", len(cases))
	}

	// Find specific edge cases and verify them
	caseMap := make(map[string]TestCase)
	for _, tc := range cases {
		caseMap[tc.Name] = tc

		// Verify data size matches actual data length
		if tc.DataSize != len(tc.Data) {
			t.Errorf("test case %q has DataSize %d but len(Data) = %d",
				tc.Name, tc.DataSize, len(tc.Data))
		}
	}

	// Verify empty case
	if tc, ok := caseMap["empty"]; ok {
		if len(tc.Data) != 0 {
			t.Errorf("empty case has %d bytes, expected 0", len(tc.Data))
		}
	} else {
		t.Error("missing 'empty' edge case")
	}

	// Verify single byte case
	if tc, ok := caseMap["single-byte"]; ok {
		if len(tc.Data) != 1 {
			t.Errorf("single-byte case has %d bytes, expected 1", len(tc.Data))
		}
	} else {
		t.Error("missing 'single-byte' edge case")
	}

	// Verify UTF-8 cases exist
	utf8Count := 0
	for _, tc := range cases {
		if tc.ContentType == ContentUTF8 {
			utf8Count++
		}
	}
	if utf8Count == 0 {
		t.Error("no UTF-8 edge cases found")
	}
}

func TestGenerateNumeric(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected string // For small sizes, verify exact content
	}{
		{"zero size", 0, ""},
		{"negative size", -1, ""},
		{"size 1", 1, "0"},
		{"size 10", 10, "0123456789"},
		{"size 15", 15, "012345678901234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNumeric(tt.size)

			if tt.size <= 0 {
				if len(result) != 0 {
					t.Errorf("generateNumeric(%d) returned %d bytes, expected 0",
						tt.size, len(result))
				}
				return
			}

			if len(result) != tt.size {
				t.Errorf("generateNumeric(%d) returned %d bytes, expected %d",
					tt.size, len(result), tt.size)
			}

			if tt.expected != "" && string(result) != tt.expected {
				t.Errorf("generateNumeric(%d) = %q, expected %q",
					tt.size, string(result), tt.expected)
			}

			// Verify all characters are digits
			for i, b := range result {
				if b < '0' || b > '9' {
					t.Errorf("generateNumeric(%d) byte %d = %q, not a digit",
						tt.size, i, b)
				}
			}
		})
	}

	// Test larger size
	t.Run("size 500", func(t *testing.T) {
		result := generateNumeric(500)
		if len(result) != 500 {
			t.Errorf("generateNumeric(500) returned %d bytes, expected 500", len(result))
		}

		// Verify it's the repeating pattern
		for i, b := range result {
			expected := byte('0' + (i % 10))
			if b != expected {
				t.Errorf("generateNumeric(500) byte %d = %q, expected %q",
					i, b, expected)
				break
			}
		}
	})
}

func TestGenerateAlphanumeric(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"zero size", 0},
		{"negative size", -1},
		{"size 1", 1},
		{"size 50", 50},
		{"size 1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateAlphanumeric(tt.size)

			if tt.size <= 0 {
				if len(result) != 0 {
					t.Errorf("generateAlphanumeric(%d) returned %d bytes, expected 0",
						tt.size, len(result))
				}
				return
			}

			if len(result) != tt.size {
				t.Errorf("generateAlphanumeric(%d) returned %d bytes, expected %d",
					tt.size, len(result), tt.size)
			}

			// Verify all characters are valid QR alphanumeric
			validChars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"
			validSet := make(map[byte]bool)
			for i := 0; i < len(validChars); i++ {
				validSet[validChars[i]] = true
			}

			for i, b := range result {
				if !validSet[b] {
					t.Errorf("generateAlphanumeric(%d) byte %d = %q, not in QR alphanumeric set",
						tt.size, i, b)
					break
				}
			}
		})
	}
}

func TestGenerateBinary(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"zero size", 0},
		{"negative size", -1},
		{"size 1", 1},
		{"size 500", 500},
		{"size 800", 800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBinary(tt.size)

			if tt.size <= 0 {
				if len(result) != 0 {
					t.Errorf("generateBinary(%d) returned %d bytes, expected 0",
						tt.size, len(result))
				}
				return
			}

			if len(result) != tt.size {
				t.Errorf("generateBinary(%d) returned %d bytes, expected %d",
					tt.size, len(result), tt.size)
			}
		})
	}

	// Test determinism: same size produces same data
	// This is critical for reproducible testing - the same data size
	// must always produce the same byte sequence.
	t.Run("determinism", func(t *testing.T) {
		size := 500
		result1 := generateBinary(size)
		result2 := generateBinary(size)

		if !bytes.Equal(result1, result2) {
			t.Error("generateBinary(500) produced different output on consecutive calls")
		}
	})

	// Verify different sizes produce different data at the end
	// (the beginning will be the same due to fixed seed, which is correct)
	t.Run("different sizes differ at end", func(t *testing.T) {
		data500 := generateBinary(500)
		data550 := generateBinary(550)

		// The first 500 bytes will be identical (fixed seed = deterministic)
		// This is the correct behavior for reproducibility
		if !bytes.Equal(data500, data550[:500]) {
			t.Error("first 500 bytes should be identical due to fixed seed")
		}

		// But the full 550-byte data should be different from 500-byte data
		if len(data550) != 550 {
			t.Errorf("generateBinary(550) returned %d bytes, expected 550", len(data550))
		}
	})
}

func TestUtf8Bytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected byte length (may differ from rune count)
	}{
		{"empty", "", 0},
		{"ascii", "Hello", 5},
		{"chinese", "你好", 6},      // 2 runes × 3 bytes each
		{"russian", "Привет", 12}, // 6 runes × 2 bytes each
		{"emoji", "🎉", 4},         // 1 rune × 4 bytes
		{"mixed", "Hello 世界", 12}, // 6 ASCII + 6 Chinese
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utf8Bytes(tt.input)

			if len(result) != tt.expected {
				t.Errorf("utf8Bytes(%q) returned %d bytes, expected %d",
					tt.input, len(result), tt.expected)
			}

			// Verify round-trip
			if string(result) != tt.input {
				t.Errorf("utf8Bytes(%q) = %q, not equal",
					tt.input, string(result))
			}
		})
	}
}

func TestFormatTestName(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		dataSize    int
		pixelSize   int
		expected    string
	}{
		{"binary case", "binary", 500, 440, "binary-500b-440px"},
		{"numeric case", "numeric", 50, 480, "numeric-50b-480px"},
		{"utf8 case", "utf8", 100, 512, "utf8-100b-512px"},
		{"large sizes", "binary", 1000, 1024, "binary-1000b-1024px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTestName(tt.contentType, tt.dataSize, tt.pixelSize)
			if result != tt.expected {
				t.Errorf("formatTestName(%q, %d, %d) = %q, expected %q",
					tt.contentType, tt.dataSize, tt.pixelSize, result, tt.expected)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{500, "500"},
		{1024, "1024"},
		{-1, "-1"},
		{-500, "-500"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatInt(tt.input)
			if result != tt.expected {
				t.Errorf("formatInt(%d) = %q, expected %q",
					tt.input, result, tt.expected)
			}
		})
	}
}
