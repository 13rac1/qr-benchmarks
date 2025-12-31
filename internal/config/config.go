// Package config manages test parameters and execution options for the QR code compatibility matrix tester.
package config

import (
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config holds all test parameters and execution options.
// Use DefaultConfig() for sensible defaults or RegisterFlags() for CLI configuration.
type Config struct {
	// DataSizes specifies byte sizes to test.
	// Default: [500, 550, 600, 650, 750, 800] - focused on pixel size matrix testing.
	DataSizes []int

	// PixelSizes specifies image pixel dimensions to test.
	// Default: [320, 400, 440, 450, 460, 480, 512, 560] - tests fractional module boundaries.
	PixelSizes []int

	// ErrorLevels specifies QR error correction levels to test.
	// Valid values: L, M, Q, H
	// Default: [L, M, Q, H] - all levels.
	ErrorLevels []string

	// Parallel enables concurrent test execution.
	// Default: true
	Parallel bool

	// Timeout sets the maximum duration for each decoder operation.
	// Default: 10s
	Timeout time.Duration

	// MaxWorkers limits concurrent worker goroutines.
	// Default: runtime.NumCPU()
	MaxWorkers int

	// SkipCGO excludes CGO-based decoders from testing.
	// Default: false
	SkipCGO bool

	// SkipArchived excludes archived libraries (e.g., goqr) from testing.
	// Default: false
	SkipArchived bool

	// OutputDir specifies the directory for test results.
	// Default: ./results
	OutputDir string

	// OutputFormats specifies report formats to generate.
	// Valid values: markdown, html, csv
	// Default: [markdown, html]
	OutputFormats []string

	// Timestamp adds timestamp to output filenames.
	// Default: true
	Timestamp bool
}

// DefaultConfig returns a Config with sensible defaults.
// Focuses on pixel size matrix testing (500-800 bytes, 320-560px).
func DefaultConfig() *Config {
	return &Config{
		DataSizes:     []int{500, 550, 600, 650, 750, 800},
		PixelSizes:    []int{320, 400, 440, 450, 460, 480, 512, 560},
		ErrorLevels:   []string{"L", "M", "Q", "H"},
		Parallel:      true,
		Timeout:       10 * time.Second,
		MaxWorkers:    runtime.NumCPU(),
		SkipCGO:       false,
		SkipArchived:  false,
		OutputDir:     "./results",
		OutputFormats: []string{"markdown", "html"},
		Timestamp:     true,
	}
}

// RegisterFlags registers CLI flags with the provided FlagSet.
// After calling fs.Parse(), call ParseFlags() to populate the Config from flag values.
//
// Example usage:
//
//	fs := flag.NewFlagSet("qr-tester", flag.ExitOnError)
//	cfg, parse := config.RegisterFlags(fs)
//	fs.Parse(os.Args[1:])
//	if err := parse(); err != nil {
//	    log.Fatal(err)
//	}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func RegisterFlags(fs *flag.FlagSet) (*Config, func() error) {
	cfg := DefaultConfig()

	var dataSizesStr string
	var pixelSizesStr string
	var errorLevelsStr string
	var outputFormatsStr string

	fs.StringVar(&dataSizesStr, "data-sizes", "", "Comma-separated data sizes in bytes (default: 500,550,600,650,750,800)")
	fs.StringVar(&pixelSizesStr, "pixel-sizes", "", "Comma-separated pixel dimensions (default: 320,400,440,450,460,480,512,560)")
	fs.StringVar(&errorLevelsStr, "error-levels", "", "Comma-separated error correction levels: L,M,Q,H (default: L,M,Q,H)")
	fs.BoolVar(&cfg.Parallel, "parallel", true, "Run tests in parallel")
	fs.DurationVar(&cfg.Timeout, "timeout", 10*time.Second, "Timeout per decoder operation")
	fs.IntVar(&cfg.MaxWorkers, "max-workers", runtime.NumCPU(), "Maximum concurrent workers")
	fs.BoolVar(&cfg.SkipCGO, "skip-cgo", false, "Skip CGO-based decoders")
	fs.BoolVar(&cfg.SkipArchived, "skip-archived", false, "Skip archived libraries")
	fs.StringVar(&cfg.OutputDir, "output", "./results", "Output directory for results")
	fs.StringVar(&outputFormatsStr, "format", "", "Comma-separated output formats: markdown,html,csv (default: markdown,html)")
	fs.BoolVar(&cfg.Timestamp, "timestamp", true, "Add timestamp to output filenames")

	// Return parse function to be called after fs.Parse()
	parse := func() error {
		if dataSizesStr != "" {
			sizes, err := parseIntSlice(dataSizesStr)
			if err != nil {
				return fmt.Errorf("invalid data-sizes: %w", err)
			}
			cfg.DataSizes = sizes
		}

		if pixelSizesStr != "" {
			sizes, err := parseIntSlice(pixelSizesStr)
			if err != nil {
				return fmt.Errorf("invalid pixel-sizes: %w", err)
			}
			cfg.PixelSizes = sizes
		}

		if errorLevelsStr != "" {
			cfg.ErrorLevels = parseStringSlice(errorLevelsStr)
		}

		if outputFormatsStr != "" {
			cfg.OutputFormats = parseStringSlice(outputFormatsStr)
		}

		return nil
	}

	return cfg, parse
}

// Validate checks that the configuration is valid.
// Returns an error if any values are invalid.
func (c *Config) Validate() error {
	if len(c.DataSizes) == 0 {
		return fmt.Errorf("data-sizes cannot be empty")
	}

	if len(c.PixelSizes) == 0 {
		return fmt.Errorf("pixel-sizes cannot be empty")
	}

	if len(c.ErrorLevels) == 0 {
		return fmt.Errorf("error-levels cannot be empty")
	}

	// Validate error correction levels
	for _, level := range c.ErrorLevels {
		if !isValidErrorLevel(level) {
			return fmt.Errorf("invalid error level %q: must be L, M, Q, or H", level)
		}
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0, got %v", c.Timeout)
	}

	if c.MaxWorkers <= 0 {
		return fmt.Errorf("max-workers must be greater than 0, got %d", c.MaxWorkers)
	}

	if len(c.OutputFormats) == 0 {
		return fmt.Errorf("output-formats cannot be empty")
	}

	// Validate output formats
	for _, format := range c.OutputFormats {
		if !isValidOutputFormat(format) {
			return fmt.Errorf("invalid output format %q: must be markdown, html, or csv", format)
		}
	}

	return nil
}

// parseIntSlice parses a comma-separated string into a slice of integers.
func parseIntSlice(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", part, err)
		}

		result = append(result, val)
	}

	return result, nil
}

// parseStringSlice parses a comma-separated string into a slice of strings.
func parseStringSlice(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// isValidErrorLevel checks if the error correction level is valid.
func isValidErrorLevel(level string) bool {
	switch level {
	case "L", "M", "Q", "H":
		return true
	default:
		return false
	}
}

// isValidOutputFormat checks if the output format is valid.
func isValidOutputFormat(format string) bool {
	switch format {
	case "markdown", "html", "csv":
		return true
	default:
		return false
	}
}
