// Package decoders provides QR code decoder implementations.
package decoders

import (
	"fmt"
	"image"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// GozxingDecoder wraps github.com/makiuchi-d/gozxing for QR code decoding.
// This decoder has known issues with fractional module pixel sizes,
// particularly when paired with the skip2/go-qrcode encoder.
type GozxingDecoder struct{}

// Name returns the decoder identifier.
func (d *GozxingDecoder) Name() string {
	return "makiuchi-d/gozxing"
}

// Decode extracts data from a QR code image.
// The gozxing library requires conversion to BinaryBitmap for decoding.
func (d *GozxingDecoder) Decode(img image.Image) ([]byte, error) {
	if img == nil {
		return nil, fmt.Errorf("gozxing: image is nil")
	}

	// Convert image to gozxing BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("gozxing: failed to create binary bitmap: %w", err)
	}

	// Create QR code reader
	reader := qrcode.NewQRCodeReader()

	// Decode the QR code
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return nil, fmt.Errorf("gozxing: decode failed: %w", err)
	}

	// Extract raw bytes from result
	return []byte(result.GetText()), nil
}
