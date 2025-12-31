//go:build cgo
// +build cgo

// Package decoders provides QR code decoder implementations.
package decoders

import (
	"fmt"
	"image"

	"github.com/kdar/goquirc"
)

// GoquircDecoder wraps github.com/kdar/goquirc for QR code decoding.
// This decoder requires CGO and a C compiler with the libquirc library.
//
// IMPORTANT: This decoder only compiles when CGO is enabled.
// Build with: CGO_ENABLED=1 go build -tags cgo
//
// This is a wrapper around the Quirc C library, which may have different
// performance and compatibility characteristics compared to pure Go decoders.
type GoquircDecoder struct{}

// Name returns the decoder identifier.
func (d *GoquircDecoder) Name() string {
	return "goquirc"
}

// Decode extracts data from a QR code image using the goquirc library.
// This decoder requires CGO and will only be available when built with CGO enabled.
//
// The goquirc library uses the Quirc C library for decoding, which may handle
// fractional module sizes differently than pure Go implementations.
func (d *GoquircDecoder) Decode(img image.Image) (data []byte, err error) {
	// Recover from panics in the goquirc library
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("goquirc: panic during decode: %v", r)
		}
	}()

	if img == nil {
		return nil, fmt.Errorf("goquirc: image is nil")
	}

	// Use goquirc.Recognize to decode the QR code
	qrCodes, decodeErr := goquirc.Recognize(img)
	if decodeErr != nil {
		return nil, fmt.Errorf("goquirc: decode failed: %w", decodeErr)
	}

	if len(qrCodes) == 0 {
		return nil, fmt.Errorf("goquirc: no QR code found in image")
	}

	// Return the payload from the first QR code found
	// goquirc can detect multiple QR codes, but we only return the first one
	return qrCodes[0].Payload, nil
}
