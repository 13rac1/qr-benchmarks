package encoders

import (
	"testing"
)

func TestBoombulerEncoder_Encode_Success(t *testing.T) {
	enc := &BoombulerEncoder{}
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

	// Verify image bounds match requested pixel size
	bounds := result.Image.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != opts.PixelSize || height != opts.PixelSize {
		t.Errorf("Image size = %dx%d, want %dx%d", width, height, opts.PixelSize, opts.PixelSize)
	}
}

func TestBoombulerEncoder_Encode_EmptyData(t *testing.T) {
	enc := &BoombulerEncoder{}
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

func TestBoombulerEncoder_Encode_ErrorCorrectionLevels(t *testing.T) {
	enc := &BoombulerEncoder{}
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
			} else {
				if err == nil {
					t.Errorf("Encode() with invalid level %q should fail", tt.level)
				}
			}
		})
	}
}

func TestBoombulerEncoder_Encode_VariousDataSizes(t *testing.T) {
	enc := &BoombulerEncoder{}

	tests := []struct {
		name     string
		dataSize int
	}{
		{"Small_100", 100},
		{"Medium_500", 500},
		{"Large_1000", 1000},
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

			// Verify image is valid
			bounds := result.Image.Bounds()
			if bounds.Empty() {
				t.Error("Encode() returned image with empty bounds")
			}
		})
	}
}

func TestBoombulerEncoder_Encode_DifferentPixelSizes(t *testing.T) {
	enc := &BoombulerEncoder{}
	data := []byte("Testing pixel size variations")

	pixelSizes := []int{320, 480, 512}

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

			bounds := result.Image.Bounds()
			width := bounds.Dx()
			height := bounds.Dy()

			if width != pixelSize || height != pixelSize {
				t.Errorf("Image size = %dx%d, want %dx%d", width, height, pixelSize, pixelSize)
			}
		})
	}
}
