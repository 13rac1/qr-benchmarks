package config

import (
	"flag"
	"runtime"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test DataSizes
	expectedDataSizes := []int{500, 550, 600, 650, 750, 800}
	if !intSliceEqual(cfg.DataSizes, expectedDataSizes) {
		t.Errorf("DataSizes = %v, want %v", cfg.DataSizes, expectedDataSizes)
	}

	// Test PixelSizes
	expectedPixelSizes := []int{320, 400, 440, 450, 460, 480, 512, 560}
	if !intSliceEqual(cfg.PixelSizes, expectedPixelSizes) {
		t.Errorf("PixelSizes = %v, want %v", cfg.PixelSizes, expectedPixelSizes)
	}

	// Test ErrorLevels
	expectedErrorLevels := []string{"L", "M", "Q", "H"}
	if !stringSliceEqual(cfg.ErrorLevels, expectedErrorLevels) {
		t.Errorf("ErrorLevels = %v, want %v", cfg.ErrorLevels, expectedErrorLevels)
	}

	// Test execution options
	if !cfg.Parallel {
		t.Error("Parallel should be true by default")
	}

	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 10*time.Second)
	}

	if cfg.MaxWorkers != runtime.NumCPU() {
		t.Errorf("MaxWorkers = %d, want %d", cfg.MaxWorkers, runtime.NumCPU())
	}

	// Test library options
	if cfg.SkipCGO {
		t.Error("SkipCGO should be false by default")
	}

	if cfg.SkipArchived {
		t.Error("SkipArchived should be false by default")
	}

	// Test output options
	if cfg.OutputDir != "./results" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "./results")
	}

	if !cfg.Timestamp {
		t.Error("Timestamp should be true by default")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := DefaultConfig()

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidate_EmptyDataSizes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DataSizes = []int{}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for empty DataSizes")
	}
}

func TestValidate_EmptyPixelSizes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PixelSizes = []int{}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for empty PixelSizes")
	}
}

func TestValidate_EmptyErrorLevels(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ErrorLevels = []string{}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for empty ErrorLevels")
	}
}

func TestValidate_InvalidErrorLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"lowercase l", "l"},
		{"lowercase m", "m"},
		{"invalid X", "X"},
		{"number", "1"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.ErrorLevels = []string{tt.level}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() error = nil, want error for invalid error level %q", tt.level)
			}
		})
	}
}

func TestValidate_ValidErrorLevels(t *testing.T) {
	validLevels := []string{"L", "M", "Q", "H"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.ErrorLevels = []string{level}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v, want nil for valid error level %q", err, level)
			}
		})
	}
}

func TestValidate_ZeroTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Timeout = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for zero Timeout")
	}
}

func TestValidate_NegativeTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Timeout = -1 * time.Second

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for negative Timeout")
	}
}

func TestValidate_ZeroMaxWorkers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxWorkers = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for zero MaxWorkers")
	}
}

func TestValidate_NegativeMaxWorkers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxWorkers = -1

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() error = nil, want error for negative MaxWorkers")
	}
}

func TestRegisterFlags_Defaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, parse := RegisterFlags(fs)

	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	err = parse()
	if err != nil {
		t.Fatalf("parse() error = %v, want nil", err)
	}

	// Should have default values since no flags were set
	if len(cfg.DataSizes) == 0 {
		t.Error("DataSizes should have default values")
	}

	if len(cfg.PixelSizes) == 0 {
		t.Error("PixelSizes should have default values")
	}
}

func TestRegisterFlags_CustomValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, parse := RegisterFlags(fs)

	err := fs.Parse([]string{
		"-data-sizes", "100,200,300",
		"-pixel-sizes", "256,512",
		"-error-levels", "L,H",
		"-parallel=false",
		"-timeout", "5s",
		"-max-workers", "2",
		"-skip-cgo=true",
		"-output", "/tmp/test",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	err = parse()
	if err != nil {
		t.Fatalf("parse() error = %v, want nil", err)
	}

	// Verify custom values
	expectedDataSizes := []int{100, 200, 300}
	if !intSliceEqual(cfg.DataSizes, expectedDataSizes) {
		t.Errorf("DataSizes = %v, want %v", cfg.DataSizes, expectedDataSizes)
	}

	expectedPixelSizes := []int{256, 512}
	if !intSliceEqual(cfg.PixelSizes, expectedPixelSizes) {
		t.Errorf("PixelSizes = %v, want %v", cfg.PixelSizes, expectedPixelSizes)
	}

	expectedErrorLevels := []string{"L", "H"}
	if !stringSliceEqual(cfg.ErrorLevels, expectedErrorLevels) {
		t.Errorf("ErrorLevels = %v, want %v", cfg.ErrorLevels, expectedErrorLevels)
	}

	if cfg.Parallel {
		t.Error("Parallel should be false")
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 5*time.Second)
	}

	if cfg.MaxWorkers != 2 {
		t.Errorf("MaxWorkers = %d, want %d", cfg.MaxWorkers, 2)
	}

	if !cfg.SkipCGO {
		t.Error("SkipCGO should be true")
	}

	if cfg.OutputDir != "/tmp/test" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "/tmp/test")
	}
}

func TestRegisterFlags_InvalidDataSizes(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_, parse := RegisterFlags(fs)

	err := fs.Parse([]string{"-data-sizes", "100,abc,300"})
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	err = parse()
	if err == nil {
		t.Error("parse() error = nil, want error for invalid data-sizes")
	}
}

func TestParseIntSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:  "simple",
			input: "100,200,300",
			want:  []int{100, 200, 300},
		},
		{
			name:  "with spaces",
			input: "100, 200, 300",
			want:  []int{100, 200, 300},
		},
		{
			name:  "single value",
			input: "500",
			want:  []int{500},
		},
		{
			name:  "empty values ignored",
			input: "100,,200",
			want:  []int{100, 200},
		},
		{
			name:    "invalid integer",
			input:   "100,abc,200",
			wantErr: true,
		},
		{
			name:    "invalid float",
			input:   "100,200.5,300",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntSlice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !intSliceEqual(got, tt.want) {
				t.Errorf("parseIntSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple",
			input: "L,M,Q,H",
			want:  []string{"L", "M", "Q", "H"},
		},
		{
			name:  "with spaces",
			input: "L, M, Q, H",
			want:  []string{"L", "M", "Q", "H"},
		},
		{
			name:  "single value",
			input: "markdown",
			want:  []string{"markdown"},
		},
		{
			name:  "empty values ignored",
			input: "markdown,,html",
			want:  []string{"markdown", "html"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStringSlice(tt.input)
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("parseStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidErrorLevel(t *testing.T) {
	validLevels := []string{"L", "M", "Q", "H"}
	for _, level := range validLevels {
		if !isValidErrorLevel(level) {
			t.Errorf("isValidErrorLevel(%q) = false, want true", level)
		}
	}

	invalidLevels := []string{"l", "m", "q", "h", "X", "1", ""}
	for _, level := range invalidLevels {
		if isValidErrorLevel(level) {
			t.Errorf("isValidErrorLevel(%q) = true, want false", level)
		}
	}
}

// Helper functions

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
