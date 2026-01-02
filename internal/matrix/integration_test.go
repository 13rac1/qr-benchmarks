package matrix

import (
	"testing"

	"github.com/13rac1/qr-library-test/internal/config"
	"github.com/13rac1/qr-library-test/internal/decoders"
	"github.com/13rac1/qr-library-test/internal/encoders"
	"github.com/13rac1/qr-library-test/internal/testdata"
)

// TestIntegration_Skip2Gozxing tests the complete encode→decode cycle with skip2 + gozxing.
// This test demonstrates the full matrix testing capability.
func TestIntegration_Skip2Gozxing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	// Use a subset of the full pixel size matrix for integration testing
	cases := testdata.GeneratePixelSizeMatrix()

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases)

	results, err := runner.RunAll()
	if err != nil {
		t.Fatalf("RunAll() failed: %v", err)
	}

	if results == nil {
		t.Fatal("RunAll() returned nil results")
	}

	// Should have 96 results (1 encoder × 1 decoder × 96 test cases)
	// 6 data sizes × 8 pixel sizes × 2 content types (alphanumeric + UTF-8)
	expectedResults := 96
	if len(results.Results) != expectedResults {
		t.Errorf("RunAll() returned %d results, want %d", len(results.Results), expectedResults)
	}

	// Verify matrix metadata
	if len(results.Encoders) != 1 || results.Encoders[0] != "skip2/go-qrcode" {
		t.Errorf("Unexpected encoders: %v", results.Encoders)
	}

	if len(results.Decoders) != 1 || results.Decoders[0] != "makiuchi-d/gozxing" {
		t.Errorf("Unexpected decoders: %v", results.Decoders)
	}

	// Count successes and failures
	successCount := 0
	failureCount := 0
	fractionalFailures := 0

	for _, result := range results.Results {
		if result.Error == nil {
			successCount++
		} else {
			failureCount++
			if result.IsFractionalModule {
				fractionalFailures++
			}
		}
	}

	t.Logf("Integration test results:")
	t.Logf("  Total tests: %d", len(results.Results))
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", failureCount)
	t.Logf("  Fractional module failures: %d", fractionalFailures)

	// We expect some tests to pass and some to fail
	// This validates that the matrix runner correctly executes and captures results
	if len(results.Results) == 0 {
		t.Error("Expected some test results")
	}

	// Verify all results have timing information
	for i, result := range results.Results {
		if result.EncodeTime == 0 {
			t.Errorf("Result %d: encode time not recorded", i)
		}
		if result.DecodeTime == 0 && result.Error == nil {
			t.Errorf("Result %d: decode time not recorded for successful decode", i)
		}
	}
}
