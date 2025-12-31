// Package decoders provides QR code decoder implementations.
package decoders

import (
	"fmt"
	"image"

	"github.com/liyue201/goqr"
)

// GoqrDecoder wraps github.com/liyue201/goqr for QR code decoding.
//
// IMPORTANT: This library is archived (last commit July 2021) and may have unfixed bugs.
// It is included for historical compatibility testing only.
// Decode failures from this decoder are expected and acceptable.
type GoqrDecoder struct{}

// Name returns the decoder identifier.
func (d *GoqrDecoder) Name() string {
	return "goqr"
}

// Decode extracts data from a QR code image.
// This archived library may fail on valid QR codes.
func (d *GoqrDecoder) Decode(img image.Image) ([]byte, error) {
	if img == nil {
		return nil, fmt.Errorf("goqr: image is nil")
	}

	// Recognize QR codes in the image
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return nil, fmt.Errorf("goqr: recognition failed: %w", err)
	}

	if len(qrCodes) == 0 {
		return nil, fmt.Errorf("goqr: no QR code found")
	}

	// Extract data from the first QR code
	// Note: goqr returns multiple codes if present, we take the first
	payload := qrCodes[0].Payload
	if payload == nil {
		return nil, fmt.Errorf("goqr: QR code payload is nil")
	}

	return payload, nil
}
