package matrix

import (
	"testing"

	"github.com/13rac1/qr-library-test/internal/config"
	"github.com/13rac1/qr-library-test/internal/decoders"
	"github.com/13rac1/qr-library-test/internal/encoders"
	"github.com/13rac1/qr-library-test/internal/testdata"
)

// TestDebug_SingleCase helps debug a single test case to understand failures.
func TestDebug_SingleCase(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	// Create a simple test case with alphanumeric data
	data := []byte("TESTING123456789")
	cases := []testdata.TestCase{
		{
			Name:        "debug-test",
			Data:        data,
			DataSize:    len(data),
			PixelSize:   320,
			ContentType: testdata.ContentAlphanumeric,
		},
	}

	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases)

	results, err := runner.RunAll()
	if err != nil {
		t.Fatalf("RunAll() failed: %v", err)
	}

	result := results.Results[0]
	t.Logf("Result: Error=%v", result.Error)
	t.Logf("EncodeTime: %v, DecodeTime: %v", result.EncodeTime, result.DecodeTime)
	t.Logf("QRVersion: %d, ModuleCount: %d", result.QRVersion, result.ModuleCount)
	t.Logf("ModulePixelSize: %.2f, IsFractional: %v", result.ModulePixelSize, result.IsFractionalModule)

	if result.Error != nil {
		t.Errorf("Test failed: %v", result.Error)
	}
}

// TestDebug_BinaryData tests with the binary data from GeneratePixelSizeMatrix.
func TestDebug_BinaryData(t *testing.T) {
	cfg := config.DefaultConfig()
	enc := &encoders.Skip2Encoder{}
	dec := &decoders.GozxingDecoder{}

	// Use actual binary data from the generator
	cases := testdata.GeneratePixelSizeMatrix()
	if len(cases) == 0 {
		t.Fatal("No test cases generated")
	}

	// Test just the first case
	runner := NewRunner(cfg, []encoders.Encoder{enc}, []decoders.Decoder{dec}, cases[:1])

	results, err := runner.RunAll()
	if err != nil {
		t.Fatalf("RunAll() failed: %v", err)
	}

	result := results.Results[0]
	t.Logf("Test case: %s", cases[0].Name)
	t.Logf("Data size: %d bytes, Pixel size: %d", cases[0].DataSize, cases[0].PixelSize)
	t.Logf("Result: Error=%v", result.Error)
	if result.Error != nil {
		t.Logf("Error details: %v", result.Error)
		t.Logf("Note: Binary data encoding/decoding failed (this may be expected)")
	}
}
