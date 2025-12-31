package testdata

import (
	"image"
	"testing"
)

func TestDetectQRVersion(t *testing.T) {
	t.Run("nil image", func(t *testing.T) {
		version, err := DetectQRVersion(nil)
		if err == nil {
			t.Fatal("expected error for nil image")
		}
		if version != -1 {
			t.Errorf("expected version -1, got %d", version)
		}
	})

	t.Run("not implemented", func(t *testing.T) {
		// Create a dummy image
		img := image.NewGray(image.Rect(0, 0, 100, 100))

		version, err := DetectQRVersion(img)
		if err == nil {
			t.Fatal("expected error for unimplemented detection")
		}
		if version != -1 {
			t.Errorf("expected version -1, got %d", version)
		}
	})
}

func TestCalculateModuleCount(t *testing.T) {
	tests := []struct {
		name     string
		version  int
		expected int
	}{
		{"version 1", 1, 21},
		{"version 2", 2, 25},
		{"version 5", 5, 37},
		{"version 10", 10, 57},
		{"version 15", 15, 77},
		{"version 20", 20, 97},
		{"version 40", 40, 177},
		{"invalid version 0", 0, 0},
		{"invalid version -1", -1, 0},
		{"invalid version 41", 41, 0},
		{"invalid version 100", 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateModuleCount(tt.version)
			if result != tt.expected {
				t.Errorf("CalculateModuleCount(%d) = %d, expected %d",
					tt.version, result, tt.expected)
			}
		})
	}
}

func TestCalculateModulePixelSize(t *testing.T) {
	tests := []struct {
		name        string
		pixelSize   int
		moduleCount int
		quietZone   int
		expected    float64
	}{
		{
			name:        "version 15 at 440px (fractional)",
			pixelSize:   440,
			moduleCount: 77,
			quietZone:   4,
			expected:    440.0 / 81.0, // ≈ 5.43
		},
		{
			name:        "version 15 at 480px (fractional)",
			pixelSize:   480,
			moduleCount: 77,
			quietZone:   4,
			expected:    480.0 / 81.0, // ≈ 5.93
		},
		{
			name:        "version 10 at 320px (fractional)",
			pixelSize:   320,
			moduleCount: 57,
			quietZone:   4,
			expected:    320.0 / 61.0, // ≈ 5.25
		},
		{
			name:        "version 1 at 100px (integer)",
			pixelSize:   100,
			moduleCount: 21,
			quietZone:   4,
			expected:    4.0,
		},
		{
			name:        "zero pixel size",
			pixelSize:   0,
			moduleCount: 21,
			quietZone:   4,
			expected:    0,
		},
		{
			name:        "zero module count",
			pixelSize:   100,
			moduleCount: 0,
			quietZone:   4,
			expected:    0,
		},
		{
			name:        "negative quiet zone",
			pixelSize:   100,
			moduleCount: 21,
			quietZone:   -1,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateModulePixelSize(tt.pixelSize, tt.moduleCount, tt.quietZone)
			if !floatEqual(result, tt.expected, 0.0001) {
				t.Errorf("CalculateModulePixelSize(%d, %d, %d) = %f, expected %f",
					tt.pixelSize, tt.moduleCount, tt.quietZone, result, tt.expected)
			}
		})
	}
}

func TestIsFractionalModuleSize(t *testing.T) {
	tests := []struct {
		name            string
		modulePixelSize float64
		expected        bool
	}{
		{"integer 5.0", 5.0, false},
		{"integer 6.0", 6.0, false},
		{"fractional 5.43", 5.43, true},
		{"fractional 5.93", 5.93, true},
		{"fractional 5.25", 5.25, true},
		{"fractional 7.72", 7.72, true},
		{"zero", 0.0, false},
		{"small fractional", 1.001, true},
		{"large integer", 100.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFractionalModuleSize(tt.modulePixelSize)
			if result != tt.expected {
				t.Errorf("IsFractionalModuleSize(%f) = %v, expected %v",
					tt.modulePixelSize, result, tt.expected)
			}
		})
	}
}

func TestCalculateOptimalPixelSize(t *testing.T) {
	tests := []struct {
		name        string
		moduleCount int
		quietZone   int
		expected    int
	}{
		{
			name:        "version 1 (21 modules)",
			moduleCount: 21,
			quietZone:   4,
			expected:    100, // First multiple of 25 >= 100 is 100
		},
		{
			name:        "version 10 (57 modules)",
			moduleCount: 57,
			quietZone:   4,
			expected:    122, // First multiple of 61 >= 100 is 122
		},
		{
			name:        "version 15 (77 modules)",
			moduleCount: 77,
			quietZone:   4,
			expected:    162, // First multiple of 81 >= 100 is 162
		},
		{
			name:        "version 40 (177 modules)",
			moduleCount: 177,
			quietZone:   4,
			expected:    181, // First multiple of 181 >= 100 is 181
		},
		{
			name:        "zero module count",
			moduleCount: 0,
			quietZone:   4,
			expected:    0,
		},
		{
			name:        "negative quiet zone",
			moduleCount: 21,
			quietZone:   -1,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOptimalPixelSize(tt.moduleCount, tt.quietZone)
			if result != tt.expected {
				t.Errorf("CalculateOptimalPixelSize(%d, %d) = %d, expected %d",
					tt.moduleCount, tt.quietZone, result, tt.expected)
			}

			// Verify result produces integer module size
			if result > 0 {
				modulePixelSize := CalculateModulePixelSize(result, tt.moduleCount, tt.quietZone)
				if IsFractionalModuleSize(modulePixelSize) {
					t.Errorf("optimal pixel size %d still produces fractional module size %f",
						result, modulePixelSize)
				}
			}
		})
	}
}

// floatEqual compares two floating point numbers within epsilon tolerance.
func floatEqual(a, b, epsilon float64) bool {
	if a == b {
		return true
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
