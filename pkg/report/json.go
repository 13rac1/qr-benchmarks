// Package report provides report generation for QR code compatibility test results.
package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/13rac1/qr-library-test/internal/matrix"
)

// JSONReporter generates JSON data files from test results.
// Outputs raw test results without aggregation.
type JSONReporter struct {
	OutputDir string
}

// NewJSONReporter creates a new JSON reporter that writes to the specified directory.
func NewJSONReporter(outputDir string) *JSONReporter {
	return &JSONReporter{
		OutputDir: outputDir,
	}
}

// RawTestResult represents a single test result in simplified form.
type RawTestResult struct {
	Encoder          string  `json:"encoder"`
	Decoder          string  `json:"decoder"`
	DataSize         int     `json:"dataSize"`
	PixelSize        int     `json:"pixelSize"`
	ContentType      string  `json:"contentType"`
	Success          bool    `json:"success"`
	ErrorType        string  `json:"errorType,omitempty"`   // "encode", "decode", "dataMismatch"
	ErrorMsg         string  `json:"errorMsg,omitempty"`
	EncodeTimeMs     float64 `json:"encodeTimeMs"`
	DecodeTimeMs     float64 `json:"decodeTimeMs"`
	QRVersion        int     `json:"qrVersion,omitempty"`
	ModuleCount      int     `json:"moduleCount,omitempty"`
	ModulePixelSize  float64 `json:"modulePixelSize,omitempty"`
	IsFractionalModule bool  `json:"isFractionalModule,omitempty"`
}

// RawResults contains all test results with metadata.
type RawResults struct {
	Timestamp string          `json:"timestamp"`
	Results   []RawTestResult `json:"results"`
}

// Generate creates a JSON file with raw test results.
func (r *JSONReporter) Generate(m *matrix.CompatibilityMatrix) error {
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	results := RawResults{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Results:   make([]RawTestResult, 0, len(m.Results)),
	}

	for _, result := range m.Results {
		raw := RawTestResult{
			Encoder:            result.EncoderName,
			Decoder:            result.DecoderName,
			DataSize:           result.DataSize,
			PixelSize:          result.PixelSize,
			ContentType:        result.ContentType,
			Success:            result.Error == nil,
			EncodeTimeMs:       toMilliseconds(result.EncodeTime),
			DecodeTimeMs:       toMilliseconds(result.DecodeTime),
			QRVersion:          result.QRVersion,
			ModuleCount:        result.ModuleCount,
			ModulePixelSize:    result.ModulePixelSize,
			IsFractionalModule: result.IsFractionalModule,
		}

		if result.Error != nil {
			raw.ErrorMsg = result.Error.Error()

			var encErr matrix.EncodeError
			if errors.As(result.Error, &encErr) {
				raw.ErrorType = "encode"
			}

			var decErr matrix.DecodeError
			if errors.As(result.Error, &decErr) {
				raw.ErrorType = "decode"
			}

			var dataErr matrix.DataMismatchError
			if errors.As(result.Error, &dataErr) {
				raw.ErrorType = "dataMismatch"
			}
		}

		results.Results = append(results.Results, raw)
	}

	return r.writeJSON(filepath.Join(r.OutputDir, "results.json"), results)
}

// writeJSON writes data to a JSON file with pretty formatting.
func (r *JSONReporter) writeJSON(path string, data interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// toMilliseconds converts a duration to milliseconds as a float.
func toMilliseconds(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000.0
}
