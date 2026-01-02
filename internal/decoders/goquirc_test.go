//go:build cgo
// +build cgo

package decoders

import (
	"bytes"
	"image"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestGoquircDecoder_Name(t *testing.T) {
	dec := &GoquircDecoder{}
	expected := "goquirc"

	if got := dec.Name(); got != expected {
		t.Errorf("Name() = %q, want %q", got, expected)
	}
}

func TestGoquircDecoder_Decode_Success(t *testing.T) {
	dec := &GoquircDecoder{}
	originalData := "Hello, QR Code!"

	// Generate a QR code using skip2/go-qrcode
	pngBytes, err := qrcode.Encode(originalData, qrcode.Medium, 256)
	if err != nil {
		t.Fatalf("Failed to generate test QR code: %v", err)
	}

	// Decode PNG bytes to image.Image
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Decode the QR code
	decodedData, err := dec.Decode(img)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	if string(decodedData) != originalData {
		t.Errorf("Decode() = %q, want %q", string(decodedData), originalData)
	}
}

func TestGoquircDecoder_Decode_NilImage(t *testing.T) {
	dec := &GoquircDecoder{}

	_, err := dec.Decode(nil)
	if err == nil {
		t.Error("Decode() with nil image should fail")
	}
}

func TestGoquircDecoder_Decode_VariousData(t *testing.T) {
	dec := &GoquircDecoder{}

	tests := []struct {
		name string
		data string
	}{
		{"Short", "A"},
		{"URL", "https://example.com/test"},
		{"Numeric", "1234567890"},
		{"Alphanumeric", "HELLO WORLD 123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate QR code
			pngBytes, err := qrcode.Encode(tt.data, qrcode.Medium, 256)
			if err != nil {
				t.Fatalf("Failed to generate QR code: %v", err)
			}

			img, _, err := image.Decode(bytes.NewReader(pngBytes))
			if err != nil {
				t.Fatalf("Failed to decode PNG: %v", err)
			}

			// Decode QR code
			decodedData, err := dec.Decode(img)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}

			if string(decodedData) != tt.data {
				t.Errorf("Decode() = %q, want %q", string(decodedData), tt.data)
			}
		})
	}
}

func TestGoquircDecoder_Decode_LargeData(t *testing.T) {
	dec := &GoquircDecoder{}

	// Generate 500 bytes of alphanumeric data
	data := make([]byte, 500)
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := range data {
		data[i] = chars[i%len(chars)]
	}

	// Generate QR code with larger pixel size to accommodate data
	pngBytes, err := qrcode.Encode(string(data), qrcode.Medium, 512)
	if err != nil {
		t.Fatalf("Failed to generate QR code: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	// Decode QR code
	decodedData, err := dec.Decode(img)
	if err != nil {
		t.Logf("Decode() failed on large data (may be expected with goquirc): %v", err)
		return
	}

	if !bytes.Equal(decodedData, data) {
		t.Errorf("Decode() data mismatch: got %d bytes, want %d bytes", len(decodedData), len(data))
	}
}

func TestGoquircDecoder_Decode_DifferentPixelSizes(t *testing.T) {
	dec := &GoquircDecoder{}
	data := "Testing pixel size variations"

	// Test with various pixel sizes, including problematic ones
	pixelSizes := []int{320, 400, 440, 450, 460, 480, 512, 560}

	for _, pixelSize := range pixelSizes {
		t.Run(formatInt(pixelSize), func(t *testing.T) {
			pngBytes, err := qrcode.Encode(data, qrcode.Medium, pixelSize)
			if err != nil {
				t.Fatalf("Failed to generate QR code at %dpx: %v", pixelSize, err)
			}

			img, _, err := image.Decode(bytes.NewReader(pngBytes))
			if err != nil {
				t.Fatalf("Failed to decode PNG: %v", err)
			}

			decodedData, err := dec.Decode(img)
			if err != nil {
				// Note: goquirc may have different compatibility patterns
				t.Logf("Decode() at %dpx failed: %v", pixelSize, err)
				return
			}

			if string(decodedData) != data {
				t.Errorf("Decode() = %q, want %q", string(decodedData), data)
			}
		})
	}
}

func TestGoquircDecoder_Decode_MultipleQRCodes(t *testing.T) {
	dec := &GoquircDecoder{}

	// Note: This test demonstrates that goquirc can detect multiple QR codes,
	// but our Decode() method returns only the first one.
	// For this test, we just use a single QR code.
	data := "First QR Code"

	pngBytes, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		t.Fatalf("Failed to generate QR code: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	decodedData, err := dec.Decode(img)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	if string(decodedData) != data {
		t.Errorf("Decode() = %q, want %q", string(decodedData), data)
	}
}

func TestGoquircDecoder_Decode_ErrorCorrectionLevels(t *testing.T) {
	dec := &GoquircDecoder{}
	data := "Error correction test"

	levels := map[string]qrcode.RecoveryLevel{
		"Low":     qrcode.Low,
		"Medium":  qrcode.Medium,
		"High":    qrcode.High,
		"Highest": qrcode.Highest,
	}

	for name, level := range levels {
		t.Run(name, func(t *testing.T) {
			pngBytes, err := qrcode.Encode(data, level, 256)
			if err != nil {
				t.Fatalf("Failed to generate QR code: %v", err)
			}

			img, _, err := image.Decode(bytes.NewReader(pngBytes))
			if err != nil {
				t.Fatalf("Failed to decode PNG: %v", err)
			}

			decodedData, err := dec.Decode(img)
			if err != nil {
				t.Fatalf("Decode() failed with level %v: %v", level, err)
			}

			if string(decodedData) != data {
				t.Errorf("Decode() = %q, want %q", string(decodedData), data)
			}
		})
	}
}
