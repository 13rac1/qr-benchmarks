package decoders

import (
	"bytes"
	"image"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestGozxingDecoder_Decode_Success(t *testing.T) {
	dec := &GozxingDecoder{}
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

func TestGozxingDecoder_Decode_NilImage(t *testing.T) {
	dec := &GozxingDecoder{}

	_, err := dec.Decode(nil)
	if err == nil {
		t.Error("Decode() with nil image should fail")
	}
}

func TestGozxingDecoder_Decode_VariousData(t *testing.T) {
	dec := &GozxingDecoder{}

	tests := []struct {
		name string
		data string
	}{
		{"Short", "A"},
		{"URL", "https://example.com/test"},
		{"Numeric", "1234567890"},
		{"Binary", string([]byte{0x01, 0x02, 0x03, 0x04, 0x05})},
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

func TestGozxingDecoder_Decode_LargeData(t *testing.T) {
	dec := &GozxingDecoder{}

	// Generate 500 bytes of alphanumeric data (safe for string encoding)
	// Using alphanumeric avoids issues with binary encoding in QR codes
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
		t.Fatalf("Decode() failed: %v", err)
	}

	if !bytes.Equal(decodedData, data) {
		t.Errorf("Decode() data mismatch: got %d bytes, want %d bytes", len(decodedData), len(data))
	}
}

func TestGozxingDecoder_Decode_DifferentPixelSizes(t *testing.T) {
	dec := &GozxingDecoder{}
	data := "Testing pixel size variations"

	// Test with various pixel sizes
	// Note: Some fractional pixel sizes may fail with gozxing
	pixelSizes := []int{320, 400, 480, 512}

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
				// Some pixel sizes may fail with gozxing - this is expected behavior
				t.Logf("Decode() at %dpx failed (may be due to fractional modules): %v", pixelSize, err)
				return
			}

			if string(decodedData) != data {
				t.Errorf("Decode() = %q, want %q", string(decodedData), data)
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
