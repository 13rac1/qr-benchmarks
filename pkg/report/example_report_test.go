package report_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/13rac1/qr-library-test/internal/matrix"
	"github.com/13rac1/qr-library-test/pkg/report"
)

// TestExampleReport demonstrates the markdown reporter with realistic data.
// This test creates an example report file that can be manually inspected.
func TestExampleReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping example report generation in short mode")
	}

	tmpDir := t.TempDir()
	reporter := report.NewMarkdownReporter(tmpDir)

	// Create realistic test data simulating skip2+gozxing incompatibility pattern
	results := createRealisticMatrix()

	err := reporter.Generate(results)
	if err != nil {
		t.Fatalf("failed to generate report: %v", err)
	}

	// Verify report was created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatalf("no report files generated")
	}

	// Read and log the report content for manual inspection
	reportPath := filepath.Join(tmpDir, entries[0].Name())
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	t.Logf("Generated report:\n%s", string(content))
	t.Logf("Report file: %s", reportPath)
}

// createRealisticMatrix creates test results that simulate the skip2+gozxing
// fractional module size incompatibility pattern described in the PRD.
func createRealisticMatrix() *matrix.CompatibilityMatrix {
	dataSizes := []int{500, 550, 600, 650, 750, 800}
	pixelSizes := []int{320, 400, 440, 450, 460, 480, 512, 560}

	// Define failure pattern (440, 450 are problematic)
	failures := map[string]bool{
		"500_440": true,
		"550_440": true,
		"550_460": true,
		"600_450": true,
		"600_560": true,
		"650_400": true,
		"650_440": true,
		"750_400": true,
		"750_460": true,
		"750_512": true,
		"800_320": true,
	}

	var results []matrix.TestResult

	for _, dataSize := range dataSizes {
		for _, pixelSize := range pixelSizes {
			key := makeKey(dataSize, pixelSize)
			isFail := failures[key]

			// Simulate QR version detection (would be implemented in later commit)
			version := estimateVersion(dataSize)
			moduleCount := 17 + 4*version
			modulePixelSize := float64(pixelSize) / float64(moduleCount+4)
			isFractional := modulePixelSize != float64(int(modulePixelSize))

			result := matrix.TestResult{
				EncoderName:        "skip2/go-qrcode",
				DecoderName:        "gozxing",
				DataSize:           dataSize,
				PixelSize:          pixelSize,
				QRVersion:          version,
				ModuleCount:        moduleCount,
				ModulePixelSize:    modulePixelSize,
				IsFractionalModule: isFractional,
				EncodeTime:         time.Duration(10+dataSize/50) * time.Millisecond,
				DecodeTime:         time.Duration(5+dataSize/100) * time.Millisecond,
				Success:            !isFail,
				DataMatches:        !isFail,
			}

			if isFail {
				result.Error = &decodeError{msg: "decoder failed to read QR code"}
			}

			results = append(results, result)
		}
	}

	return &matrix.CompatibilityMatrix{
		Results:    results,
		Encoders:   []string{"skip2/go-qrcode"},
		Decoders:   []string{"gozxing"},
		DataSizes:  dataSizes,
		PixelSizes: pixelSizes,
	}
}

func makeKey(dataSize, pixelSize int) string {
	return string(rune('0'+dataSize/100)) + string(rune('0'+(dataSize%100)/10)) + string(rune('0'+dataSize%10)) +
		"_" +
		string(rune('0'+pixelSize/100)) + string(rune('0'+(pixelSize%100)/10)) + string(rune('0'+pixelSize%10))
}

func estimateVersion(dataSize int) int {
	// Rough estimation for binary mode with medium error correction
	if dataSize <= 32 {
		return 1
	}
	if dataSize <= 53 {
		return 2
	}
	if dataSize <= 78 {
		return 3
	}
	if dataSize <= 106 {
		return 4
	}
	if dataSize <= 134 {
		return 5
	}
	if dataSize <= 154 {
		return 6
	}
	if dataSize <= 192 {
		return 7
	}
	if dataSize <= 230 {
		return 8
	}
	if dataSize <= 271 {
		return 9
	}
	if dataSize <= 321 {
		return 10
	}
	if dataSize <= 367 {
		return 11
	}
	if dataSize <= 425 {
		return 12
	}
	if dataSize <= 458 {
		return 13
	}
	if dataSize <= 520 {
		return 14
	}
	if dataSize <= 586 {
		return 15
	}
	if dataSize <= 644 {
		return 16
	}
	if dataSize <= 718 {
		return 17
	}
	if dataSize <= 792 {
		return 18
	}
	return 19
}

type decodeError struct {
	msg string
}

func (e *decodeError) Error() string {
	return e.msg
}
