// Package decoders provides QR code decoder implementations.
package decoders

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/tuotoo/qrcode"
)

// TuotooDecoder wraps github.com/tuotoo/qrcode for QR code decoding.
// This is a pure Go decoder with dynamic binarization that may have
// different success patterns compared to gozxing.
type TuotooDecoder struct{}

// Name returns the decoder identifier.
func (d *TuotooDecoder) Name() string {
	return "tuotoo/qrcode"
}

// Decode extracts data from a QR code image.
// The tuotoo library requires an io.Reader, so we convert the image to PNG bytes.
// This decoder handles panics from the underlying library and returns them as errors.
func (d *TuotooDecoder) Decode(img image.Image) (data []byte, err error) {
	// Recover from panics in the tuotoo library
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tuotoo: panic during decode: %v", r)
		}
	}()

	if img == nil {
		return nil, fmt.Errorf("tuotoo: image is nil")
	}

	// Convert image to PNG bytes in buffer
	buf := new(bytes.Buffer)
	if encodeErr := png.Encode(buf, img); encodeErr != nil {
		return nil, fmt.Errorf("tuotoo: failed to encode image to PNG: %w", encodeErr)
	}

	// Decode QR code from buffer
	qrData, decodeErr := qrcode.Decode(buf)
	if decodeErr != nil {
		return nil, fmt.Errorf("tuotoo: decode failed: %w", decodeErr)
	}

	// Extract raw data from QR code
	return []byte(qrData.Content), nil
}
