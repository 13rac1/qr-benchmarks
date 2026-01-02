package encoders

import "github.com/13rac1/qr-library-test/internal/config"

// GetAvailableEncoders returns the list of encoders available based on configuration.
// Always includes pure Go encoders.
// Conditionally includes CGO encoders if CGO is enabled at build time and not skipped.
func GetAvailableEncoders(cfg *config.Config) []Encoder {
	encoders := []Encoder{
		&Skip2Encoder{},
		&BoombulerEncoder{},
		&YeqownEncoder{},
		&GozxingEncoder{},
	}

	return encoders
}

// GetAllEncoders returns all encoders regardless of configuration.
func GetAllEncoders() []Encoder {
	return []Encoder{
		&Skip2Encoder{},
		&BoombulerEncoder{},
		&YeqownEncoder{},
		&GozxingEncoder{},
	}
}
