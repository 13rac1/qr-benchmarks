//go:build !cgo
// +build !cgo

package decoders

import (
	"fmt"
	"image"
)

// GoquircDecoder is a stub when CGO is not available.
// This allows the code to compile without CGO while making the type unavailable.
// The registry will not include this decoder when cgoEnabled() returns false.
type GoquircDecoder struct{}

// Name returns the decoder identifier.
func (d *GoquircDecoder) Name() string {
	return "goquirc"
}

// Decode always returns an error when CGO is not available.
// This method should never be called because the registry excludes
// GoquircDecoder when CGO is disabled.
func (d *GoquircDecoder) Decode(img image.Image) ([]byte, error) {
	return nil, fmt.Errorf("goquirc: decoder not available (CGO not enabled)")
}
