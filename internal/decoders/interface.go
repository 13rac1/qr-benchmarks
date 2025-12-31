// Package decoders defines the interface for QR code decoders.
package decoders

import "image"

// Decoder extracts data from QR code images.
// Implementations wrap different QR decoding libraries to provide a uniform interface.
type Decoder interface {
	// Name returns the decoder's identifier (e.g., "gozxing", "tuotoo").
	// Used for reporting and result tracking.
	Name() string

	// Decode extracts data from a QR code image.
	// Returns the decoded bytes and any error encountered.
	// Common errors: unreadable QR code, corrupted data, timeout.
	// Implementations should handle panics internally and return them as errors.
	Decode(img image.Image) ([]byte, error)
}
