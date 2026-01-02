package encoders

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"

	"github.com/KangSpace/gqrcode"
	"github.com/KangSpace/gqrcode/core/cons"
	"github.com/KangSpace/gqrcode/core/mode"
	"github.com/KangSpace/gqrcode/core/output"
)

// GqrcodeEncoder wraps github.com/KangSpace/gqrcode for QR code generation.
// This is a pure Go implementation following ISO/IEC 18004-2015.
type GqrcodeEncoder struct{}

// Name returns the encoder identifier.
func (e *GqrcodeEncoder) Name() string {
	return "KangSpace/gqrcode"
}

// Encode generates a QR code image from the input data.
func (e *GqrcodeEncoder) Encode(data []byte, opts EncodeOptions) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("gqrcode: cannot encode empty data")
	}

	// Map error correction level
	var ecLevel cons.ErrorCorrectionLevel
	switch opts.ErrorCorrectionLevel {
	case ErrorCorrectionL:
		ecLevel = cons.L
	case ErrorCorrectionM:
		ecLevel = cons.M
	case ErrorCorrectionQ:
		ecLevel = cons.Q
	case ErrorCorrectionH:
		ecLevel = cons.H
	default:
		return nil, fmt.Errorf("gqrcode: invalid error correction level %q", opts.ErrorCorrectionLevel)
	}

	// Create QR code with error correction level
	ec := mode.NewErrorCorrection(ecLevel)
	qr, err := gqrcode.NewQRCode0(string(data), cons.QrcodeModel2, ec, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("gqrcode: create failed: %w", err)
	}

	// Create PNG output with specified size
	out := output.NewPNGOutput(opts.PixelSize)

	// Encode to writer
	var buf bytes.Buffer
	if err := qr.EncodeToWriter(out, &buf); err != nil {
		return nil, fmt.Errorf("gqrcode: encode failed: %w", err)
	}

	// Decode PNG bytes to image.Image
	img, _, err := image.Decode(&buf)
	if err != nil {
		return nil, fmt.Errorf("gqrcode: PNG decode failed: %w", err)
	}

	return img, nil
}
