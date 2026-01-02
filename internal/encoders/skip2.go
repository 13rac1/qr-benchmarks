// Package encoders provides QR code encoder implementations.
package encoders

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"strings"

	"github.com/skip2/go-qrcode"
)

// Skip2Encoder wraps github.com/skip2/go-qrcode for QR code generation.
// This encoder is known to produce QR codes that have compatibility issues
// with gozxing decoder when using fractional module pixel sizes.
//
// Note: skip2/go-qrcode treats input as a string. Binary data containing
// null bytes and special characters may not round-trip correctly through
// the encode→decode cycle. This is a library limitation, not a bug in this wrapper.
type Skip2Encoder struct{}

// Name returns the encoder identifier.
func (e *Skip2Encoder) Name() string {
	return "skip2/go-qrcode"
}

// Encode generates a QR code image from the input data.
// The skip2/go-qrcode library generates PNG bytes which are decoded back to image.Image.
func (e *Skip2Encoder) Encode(data []byte, opts EncodeOptions) (EncodeResult, error) {
	if len(data) == 0 {
		return EncodeResult{}, fmt.Errorf("skip2: cannot encode empty data")
	}

	// Map error correction level to qrcode package constants
	var level qrcode.RecoveryLevel
	switch opts.ErrorCorrectionLevel {
	case ErrorCorrectionL:
		level = qrcode.Low
	case ErrorCorrectionM:
		level = qrcode.Medium
	case ErrorCorrectionQ:
		level = qrcode.High
	case ErrorCorrectionH:
		level = qrcode.Highest
	default:
		return EncodeResult{}, fmt.Errorf("skip2: invalid error correction level %q", opts.ErrorCorrectionLevel)
	}

	// Create QRCode struct to access version
	qr, err := qrcode.New(string(data), level)
	if err != nil {
		return EncodeResult{}, fmt.Errorf("skip2: encode failed: %w", err)
	}

	// Generate PNG at requested size
	pngBytes, err := qr.PNG(opts.PixelSize)
	if err != nil {
		return EncodeResult{}, fmt.Errorf("skip2: PNG generation failed: %w", err)
	}

	// Decode PNG bytes to image.Image
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return EncodeResult{}, fmt.Errorf("skip2: PNG decode failed: %w", err)
	}

	return EncodeResult{
		Image:   img,
		Version: qr.VersionNumber,
	}, nil
}

// IsCapacityError returns true if the error indicates data exceeds QR capacity.
func (e *Skip2Encoder) IsCapacityError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "content too long to encode")
}
