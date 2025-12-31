// Package decoders provides QR code decoder implementations.
package decoders

import "github.com/13rac1/qr-library-test/internal/config"

// GetAvailableDecoders returns the list of decoders available based on configuration.
// Always includes pure Go decoders (gozxing, tuotoo).
// Conditionally includes:
//   - goqr if !cfg.SkipArchived
//   - goquirc if !cfg.SkipCGO (added in commit 8)
func GetAvailableDecoders(cfg *config.Config) []Decoder {
	decoders := []Decoder{
		&GozxingDecoder{},
		&TuotooDecoder{},
	}

	if !cfg.SkipArchived {
		decoders = append(decoders, &GoqrDecoder{})
	}

	// CGO-based decoders will be added in commit 8:
	// if !cfg.SkipCGO {
	//     decoders = append(decoders, &GoquircDecoder{})
	// }

	return decoders
}

// GetAllDecoders returns all decoders regardless of configuration.
// Used for testing and full matrix runs.
// Returns all implemented decoders: gozxing, tuotoo, goqr.
// Note: goquirc will be added in commit 8.
func GetAllDecoders() []Decoder {
	return []Decoder{
		&GozxingDecoder{},
		&TuotooDecoder{},
		&GoqrDecoder{},
	}
}
