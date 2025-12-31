package matrix

import (
	"testing"

	"github.com/13rac1/qr-library-test/internal/config"
	"github.com/13rac1/qr-library-test/internal/decoders"
	"github.com/13rac1/qr-library-test/internal/encoders"
	"github.com/13rac1/qr-library-test/internal/testdata"
)

func TestNewRunner(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}
	cases := []testdata.TestCase{
		{
			Name:        "test-100b-320px",
			Data:        make([]byte, 100),
			DataSize:    100,
			PixelSize:   320,
			ContentType: testdata.ContentBinary,
		},
	}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases)

	if runner == nil {
		t.Fatal("NewRunner() returned nil")
	}

	if len(runner.Encoders) != 1 {
		t.Errorf("Runner has %d encoders, want 1", len(runner.Encoders))
	}

	if len(runner.Decoders) != 1 {
		t.Errorf("Runner has %d decoders, want 1", len(runner.Decoders))
	}

	if len(runner.TestCases) != 1 {
		t.Errorf("Runner has %d test cases, want 1", len(runner.TestCases))
	}

	if runner.Config != cfg {
		t.Error("Runner config does not match provided config")
	}
}

func TestRunner_RunAll_NoEncoders(t *testing.T) {
	cfg := config.DefaultConfig()
	dec := &decoders.GozxingDecoder{}
	cases := []testdata.TestCase{
		{
			Name:        "test-100b-320px",
			Data:        make([]byte, 100),
			DataSize:    100,
			PixelSize:   320,
			ContentType: testdata.ContentBinary,
		},
	}

	runner := NewRunner(cfg, []encoders.Encoder{}, []decoders.Decoder{dec}, cases)

	_, err := runner.RunAll()
	if err == nil {
		t.Error("RunAll() with no encoders should fail")
	}
}

func TestRunner_RunAll_NoDecoders(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	cases := []testdata.TestCase{
		{
			Name:        "test-100b-320px",
			Data:        make([]byte, 100),
			DataSize:    100,
			PixelSize:   320,
			ContentType: testdata.ContentBinary,
		},
	}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{}, cases)

	_, err := runner.RunAll()
	if err == nil {
		t.Error("RunAll() with no decoders should fail")
	}
}

func TestRunner_RunAll_NoTestCases(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, []testdata.TestCase{})

	_, err := runner.RunAll()
	if err == nil {
		t.Error("RunAll() with no test cases should fail")
	}
}

func TestRunner_RunAll_SingleTest(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	// Create simple test data
	data := []byte("Hello, QR Code!")
	cases := []testdata.TestCase{
		{
			Name:        "test-simple",
			Data:        data,
			DataSize:    len(data),
			PixelSize:   320,
			ContentType: testdata.ContentBinary,
		},
	}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases)

	results, err := runner.RunAll()
	if err != nil {
		t.Fatalf("RunAll() failed: %v", err)
	}

	if results == nil {
		t.Fatal("RunAll() returned nil results")
	}

	// Should have 1 result (1 encoder × 1 decoder × 1 test case)
	if len(results.Results) != 1 {
		t.Errorf("RunAll() returned %d results, want 1", len(results.Results))
	}

	// Check result details
	result := results.Results[0]
	if result.EncoderName != enc.Name() {
		t.Errorf("Result encoder name = %q, want %q", result.EncoderName, enc.Name())
	}

	if result.DecoderName != dec.Name() {
		t.Errorf("Result decoder name = %q, want %q", result.DecoderName, dec.Name())
	}

	if result.DataSize != len(data) {
		t.Errorf("Result data size = %d, want %d", result.DataSize, len(data))
	}

	if result.PixelSize != 320 {
		t.Errorf("Result pixel size = %d, want 320", result.PixelSize)
	}

	// Check timing was recorded
	if result.EncodeTime == 0 {
		t.Error("Result encode time not recorded")
	}

	if result.DecodeTime == 0 {
		t.Error("Result decode time not recorded")
	}

	// This simple test should succeed
	if result.Error != nil {
		t.Errorf("Result should succeed, got error: %v", result.Error)
	}
}

func TestRunner_RunAll_MultipleTests(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	// Create multiple test cases
	cases := []testdata.TestCase{
		{
			Name:        "test-100b-320px",
			Data:        generateTestData(100),
			DataSize:    100,
			PixelSize:   320,
			ContentType: testdata.ContentBinary,
		},
		{
			Name:        "test-100b-480px",
			Data:        generateTestData(100),
			DataSize:    100,
			PixelSize:   480,
			ContentType: testdata.ContentBinary,
		},
		{
			Name:        "test-200b-320px",
			Data:        generateTestData(200),
			DataSize:    200,
			PixelSize:   320,
			ContentType: testdata.ContentBinary,
		},
	}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases)

	results, err := runner.RunAll()
	if err != nil {
		t.Fatalf("RunAll() failed: %v", err)
	}

	// Should have 3 results (1 encoder × 1 decoder × 3 test cases)
	if len(results.Results) != 3 {
		t.Errorf("RunAll() returned %d results, want 3", len(results.Results))
	}

	// Verify matrix metadata
	if len(results.Encoders) != 1 {
		t.Errorf("Matrix has %d encoders, want 1", len(results.Encoders))
	}

	if len(results.Decoders) != 1 {
		t.Errorf("Matrix has %d decoders, want 1", len(results.Decoders))
	}

	// Should have 2 unique data sizes (100, 200)
	if len(results.DataSizes) != 2 {
		t.Errorf("Matrix has %d data sizes, want 2", len(results.DataSizes))
	}

	// Should have 2 unique pixel sizes (320, 480)
	if len(results.PixelSizes) != 2 {
		t.Errorf("Matrix has %d pixel sizes, want 2", len(results.PixelSizes))
	}
}

func TestRunner_RunAll_WithPixelSizeMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pixel size matrix test in short mode")
	}

	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	// Generate a subset of the full pixel size matrix for testing
	// Using smaller data sizes to speed up the test
	dataSizes := []int{100, 200}
	pixelSizes := []int{320, 480}

	cases := make([]testdata.TestCase, 0, len(dataSizes)*len(pixelSizes))
	for _, dataSize := range dataSizes {
		data := generateTestData(dataSize)
		for _, pixelSize := range pixelSizes {
			cases = append(cases, testdata.TestCase{
				Name:        formatTestName("binary", dataSize, pixelSize),
				Data:        data,
				DataSize:    dataSize,
				PixelSize:   pixelSize,
				ContentType: testdata.ContentBinary,
			})
		}
	}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases)

	results, err := runner.RunAll()
	if err != nil {
		t.Fatalf("RunAll() failed: %v", err)
	}

	// Should have 4 results (1 encoder × 1 decoder × 4 test cases)
	expectedResults := len(dataSizes) * len(pixelSizes)
	if len(results.Results) != expectedResults {
		t.Errorf("RunAll() returned %d results, want %d", len(results.Results), expectedResults)
	}

	// Track success/failure statistics
	successCount := 0
	failureCount := 0

	for _, result := range results.Results {
		if result.Error == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	t.Logf("Test results: %d successful, %d failed", successCount, failureCount)

	// We expect at least some tests to pass
	if successCount == 0 {
		t.Error("Expected at least some tests to succeed")
	}
}

// generateTestData creates deterministic test data for testing.
func generateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// formatTestName creates a test case identifier.
func formatTestName(contentType string, dataSize, pixelSize int) string {
	return contentType + "-" + formatInt(dataSize) + "b-" + formatInt(pixelSize) + "px"
}

// formatInt converts an integer to a string.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := 0
	temp := n
	for temp > 0 {
		digits++
		temp /= 10
	}

	result := make([]byte, digits)
	for i := digits - 1; i >= 0; i-- {
		result[i] = byte('0' + n%10)
		n /= 10
	}

	if negative {
		return "-" + string(result)
	}
	return string(result)
}
