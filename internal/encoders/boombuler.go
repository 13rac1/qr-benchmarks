// Package encoders provides QR code encoder implementations.
package encoders

import (
	"fmt"
	"image"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// BoombulerEncoder wraps github.com/boombuler/barcode for QR code generation.
// This encoder returns image.Image directly from the barcode interface.
type BoombulerEncoder struct{}

// Name returns the encoder identifier.
func (e *BoombulerEncoder) Name() string {
	return "boombuler/barcode"
}

// Encode generates a QR code image from the input data.
// The boombuler/barcode library generates a Barcode interface which implements image.Image.
func (e *BoombulerEncoder) Encode(data []byte, opts EncodeOptions) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("boombuler: cannot encode empty data")
	}

	// Map error correction level to qr package constants
	var level qr.ErrorCorrectionLevel
	switch opts.ErrorCorrectionLevel {
	case ErrorCorrectionL:
		level = qr.L
	case ErrorCorrectionM:
		level = qr.M
	case ErrorCorrectionQ:
		level = qr.Q
	case ErrorCorrectionH:
		level = qr.H
	default:
		return nil, fmt.Errorf("boombuler: invalid error correction level %q", opts.ErrorCorrectionLevel)
	}

	// Encode using Unicode mode for binary data
	qrCode, err := qr.Encode(string(data), level, qr.Unicode)
	if err != nil {
		return nil, fmt.Errorf("boombuler: encode failed: %w", err)
	}

	// Scale barcode to desired pixel size
	scaled, err := barcode.Scale(qrCode, opts.PixelSize, opts.PixelSize)
	if err != nil {
		return nil, fmt.Errorf("boombuler: scale failed: %w", err)
	}

	return scaled, nil
}
