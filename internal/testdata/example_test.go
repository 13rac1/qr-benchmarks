package testdata

import (
	"fmt"
	"testing"
)

// ExampleGeneratePixelSizeMatrix demonstrates the pixel size matrix generation.
// This example is not run automatically but serves as documentation.
func ExampleGeneratePixelSizeMatrix() {
	cases := GeneratePixelSizeMatrix()

	fmt.Printf("Generated %d test cases\n", len(cases))
	fmt.Printf("\nFirst few test cases:\n")

	for i := 0; i < 4 && i < len(cases); i++ {
		tc := cases[i]
		fmt.Printf("  %s: %d bytes at %dpx EC%s\n", tc.Name, tc.DataSize, tc.PixelSize, tc.ErrorCorrectionLevel)
	}

	// Output:
	// Generated 96 test cases
	//
	// First few test cases:
	//   alphanumeric-100b-264px-ecL: 100 bytes at 264px ECL
	//   alphanumeric-100b-264px-ecH: 100 bytes at 264px ECH
	//   utf8-100b-264px-ecL: 100 bytes at 264px ECL
	//   utf8-100b-264px-ecH: 100 bytes at 264px ECH
}

// TestModuleCalculationExample demonstrates module size calculations
// that identify fractional module sizing issues.
func TestModuleCalculationExample(t *testing.T) {
	// Version 15 is known to be problematic at certain pixel sizes
	version := 15
	moduleCount := CalculateModuleCount(version)

	testCases := []struct {
		pixelSize int
		describe  string
	}{
		{440, "problematic"},
		{450, "problematic"},
		{480, "safe"},
		{512, "safe"},
	}

	t.Logf("QR Version %d has %d modules", version, moduleCount)
	t.Logf("\nPixel size analysis:")

	for _, tc := range testCases {
		modulePixelSize := CalculateModulePixelSize(tc.pixelSize, moduleCount, QuietZoneModules)
		isFractional := IsFractionalModuleSize(modulePixelSize)

		fractionalStr := "integer"
		if isFractional {
			fractionalStr = "FRACTIONAL"
		}

		t.Logf("  %dpx: %.2f pixels/module [%s] - %s",
			tc.pixelSize, modulePixelSize, fractionalStr, tc.describe)
	}

	// Find optimal pixel size
	optimal := CalculateOptimalPixelSize(moduleCount, QuietZoneModules)
	t.Logf("\nOptimal pixel size: %dpx (integer modules)", optimal)
}

// TestDataGenerationProperties verifies key properties of generated test data.
func TestDataGenerationProperties(t *testing.T) {
	t.Run("pixel size matrix coverage", func(t *testing.T) {
		cases := GeneratePixelSizeMatrix()

		// Count unique data sizes and pixel sizes
		dataSizes := make(map[int]bool)
		pixelSizes := make(map[int]bool)

		for _, tc := range cases {
			dataSizes[tc.DataSize] = true
			pixelSizes[tc.PixelSize] = true
		}

		if len(dataSizes) != 4 {
			t.Errorf("expected 4 unique data sizes, got %d", len(dataSizes))
		}

		if len(pixelSizes) != 6 {
			t.Errorf("expected 6 unique pixel sizes, got %d", len(pixelSizes))
		}

		if len(cases) != 96 {
			t.Errorf("expected 96 test cases (4×6×4), got %d", len(cases))
		}

		t.Logf("Matrix: %d data sizes × %d pixel sizes = %d test cases",
			len(dataSizes), len(pixelSizes), len(cases))
	})

	t.Run("binary data determinism", func(t *testing.T) {
		// Same parameters should produce identical data
		data1 := generateBinary(500)
		data2 := generateBinary(500)

		if len(data1) != len(data2) {
			t.Fatalf("data lengths differ: %d vs %d", len(data1), len(data2))
		}

		for i := range data1 {
			if data1[i] != data2[i] {
				t.Errorf("byte %d differs: %d vs %d", i, data1[i], data2[i])
				break
			}
		}

		t.Log("Binary data generation is deterministic")
	})
}
