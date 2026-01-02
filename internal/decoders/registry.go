// Package decoders provides QR code decoder implementations.
package decoders

import "github.com/13rac1/qr-library-test/internal/config"

// GetAvailableDecoders returns the list of decoders available based on configuration.
// Always includes pure Go decoders (gozxing, tuotoo).
// Conditionally includes:
//   - goqr if !cfg.SkipArchived
//   - goquirc if !cfg.SkipCGO and CGO is enabled at build time
func GetAvailableDecoders(cfg *config.Config) []Decoder {
	decoders := []Decoder{
		&GozxingDecoder{},
		&TuotooDecoder{},
	}

	if !cfg.SkipArchived {
		decoders = append(decoders, &GoqrDecoder{})
	}

	// CGO decoders - only include if CGO enabled at build time and not skipped
	if !cfg.SkipCGO && cgoEnabled() {
		decoders = append(decoders, &GoquircDecoder{})
	}

	return decoders
}

// GetAllDecoders returns all decoders regardless of configuration.
// Used for testing and full matrix runs.
func GetAllDecoders() []Decoder {
	decoders := []Decoder{
		&GozxingDecoder{},
		&TuotooDecoder{},
		&GoqrDecoder{},
	}

	// Include CGO decoders if available at build time
	if cgoEnabled() {
		decoders = append(decoders, &GoquircDecoder{})
	}

	return decoders
}
