package encoders

import (
	"testing"
)

func TestGqrcodeEncoder_Name(t *testing.T) {
	enc := &GqrcodeEncoder{}
	expected := "KangSpace/gqrcode"

	if got := enc.Name(); got != expected {
		t.Errorf("Name() = %q, want %q", got, expected)
	}
}

func TestGqrcodeEncoder_Encode_Success(t *testing.T) {
	enc := &GqrcodeEncoder{}
	data := []byte("Hello, QR Code!")

	opts := EncodeOptions{
		ErrorCorrectionLevel: ErrorCorrectionM,
		PixelSize:            256,
	}

	img, err := enc.Encode(data, opts)
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	if img == nil {
		t.Fatal("Encode() returned nil image")
	}

	// Verify image has valid dimensions
	bounds := img.Bounds()
	if bounds.Empty() {
		t.Error("Encode() returned image with empty bounds")
	}
}

func TestGqrcodeEncoder_Encode_EmptyData(t *testing.T) {
	enc := &GqrcodeEncoder{}
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

func TestGqrcodeEncoder_Encode_ErrorCorrectionLevels(t *testing.T) {
	enc := &GqrcodeEncoder{}
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

			img, err := enc.Encode(data, opts)

			if tt.valid {
				if err != nil {
					t.Errorf("Encode() with level %q failed: %v", tt.level, err)
				}
				if img == nil {
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

func TestGqrcodeEncoder_Encode_VariousDataSizes(t *testing.T) {
	enc := &GqrcodeEncoder{}

	tests := []struct {
		name     string
		dataSize int
	}{
		{"Small", 50},
		{"Medium", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate alphanumeric test data (gqrcode handles this well)
			data := make([]byte, tt.dataSize)
			for i := range data {
				data[i] = byte('A' + (i % 26))
			}

			opts := EncodeOptions{
				ErrorCorrectionLevel: ErrorCorrectionL,
				PixelSize:            512,
			}

			img, err := enc.Encode(data, opts)
			if err != nil {
				t.Fatalf("Encode() with %d bytes failed: %v", tt.dataSize, err)
			}

			if img == nil {
				t.Fatal("Encode() returned nil image")
			}

			bounds := img.Bounds()
			if bounds.Empty() {
				t.Error("Encode() returned image with empty bounds")
			}
		})
	}
}
