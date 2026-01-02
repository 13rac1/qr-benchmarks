package encoders

import (
	"testing"
)

func TestSkip2Encoder_Encode_Success(t *testing.T) {
	enc := &Skip2Encoder{}
	data := []byte("Hello, QR Code!")

	opts := EncodeOptions{
		ErrorCorrectionLevel: ErrorCorrectionM,
		PixelSize:            256,
	}

	result, err := enc.Encode(data, opts)
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	if result.Image == nil {
		t.Fatal("Encode() returned nil image")
	}

	if result.Version < 1 || result.Version > 40 {
		t.Errorf("Version = %d, want 1-40", result.Version)
	}

	// Verify image bounds match requested pixel size
	bounds := result.Image.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != opts.PixelSize || height != opts.PixelSize {
		t.Errorf("Image size = %dx%d, want %dx%d", width, height, opts.PixelSize, opts.PixelSize)
	}
}

func TestSkip2Encoder_Encode_EmptyData(t *testing.T) {
	enc := &Skip2Encoder{}
	data := []byte{}

	opts := EncodeOptions{
		ErrorCorrectionLevel: ErrorCorrectionM,
		PixelSize:            256,
	}

	_, err := enc.Encode(data, opts)
	if err == nil {
		t.Error("Encode() with empty data should fail")
	}
}

func TestSkip2Encoder_Encode_ErrorCorrectionLevels(t *testing.T) {
	enc := &Skip2Encoder{}
	data := []byte("Test data for error correction levels")

	tests := []struct {
		name  string
		level string
		valid bool
	}{
		{"Low", ErrorCorrectionL, true},
		{"Medium", ErrorCorrectionM, true},
		{"Quartile", ErrorCorrectionQ, true},
		{"High", ErrorCorrectionH, true},
		{"Invalid", "X", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := EncodeOptions{
				ErrorCorrectionLevel: tt.level,
				PixelSize:            256,
			}

			result, err := enc.Encode(data, opts)

			if tt.valid {
				if err != nil {
					t.Errorf("Encode() with level %q failed: %v", tt.level, err)
				}
				if result.Image == nil {
					t.Error("Encode() returned nil image")
				}
				if result.Version < 1 || result.Version > 40 {
					t.Errorf("Version = %d, want 1-40", result.Version)
				}
			} else {
				if err == nil {
					t.Errorf("Encode() with invalid level %q should fail", tt.level)
				}
			}
		})
	}
}

func TestSkip2Encoder_Encode_VariousDataSizes(t *testing.T) {
	enc := &Skip2Encoder{}

	tests := []struct {
		name     string
		dataSize int
	}{
		{"Small", 50},
		{"Medium", 500},
		{"Large", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate test data
			data := make([]byte, tt.dataSize)
			for i := range data {
				data[i] = byte(i % 256)
			}

			opts := EncodeOptions{
				ErrorCorrectionLevel: ErrorCorrectionM,
				PixelSize:            512,
			}

			result, err := enc.Encode(data, opts)
			if err != nil {
				t.Fatalf("Encode() with %d bytes failed: %v", tt.dataSize, err)
			}

			if result.Image == nil {
				t.Fatal("Encode() returned nil image")
			}

			if result.Version < 1 || result.Version > 40 {
				t.Errorf("Version = %d, want 1-40", result.Version)
			}

			// Verify image is valid
			bounds := result.Image.Bounds()
			if bounds.Empty() {
				t.Error("Encode() returned image with empty bounds")
			}
		})
	}
}

func TestSkip2Encoder_Encode_DifferentPixelSizes(t *testing.T) {
	enc := &Skip2Encoder{}
	data := []byte("Testing pixel size variations")

	pixelSizes := []int{320, 400, 440, 450, 480, 512, 560}

	for _, pixelSize := range pixelSizes {
		t.Run(formatInt(pixelSize), func(t *testing.T) {
			opts := EncodeOptions{
				ErrorCorrectionLevel: ErrorCorrectionM,
				PixelSize:            pixelSize,
			}

			result, err := enc.Encode(data, opts)
			if err != nil {
				t.Fatalf("Encode() at %dpx failed: %v", pixelSize, err)
			}

			if result.Version < 1 || result.Version > 40 {
				t.Errorf("Version = %d, want 1-40", result.Version)
			}

			bounds := result.Image.Bounds()
			width := bounds.Dx()
			height := bounds.Dy()

			if width != pixelSize || height != pixelSize {
				t.Errorf("Image size = %dx%d, want %dx%d", width, height, pixelSize, pixelSize)
			}
		})
	}
}

// formatInt is a simple helper to convert int to string for test names
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := 0
	temp := n
	for temp > 0 {
		digits++
		temp /= 10
	}

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
