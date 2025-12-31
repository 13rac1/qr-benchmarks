package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/13rac1/qr-library-test/internal/matrix"
)

func TestNewMarkdownReporter(t *testing.T) {
	reporter := NewMarkdownReporter("./test-output")

	if reporter.OutputDir != "./test-output" {
		t.Errorf("expected output dir './test-output', got '%s'", reporter.OutputDir)
	}
}

func TestGenerate_CreatesOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "reports")

	reporter := NewMarkdownReporter(outputDir)

	results := createSampleMatrix()
	err := reporter.Generate(results)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("output directory was not created")
	}
}

func TestGenerate_CreatesReportFiles(t *testing.T) {
	tmpDir := t.TempDir()
	reporter := NewMarkdownReporter(tmpDir)

	results := createSampleMatrix()
	err := reporter.Generate(results)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that files were created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}

	if len(entries) == 0 {
		t.Errorf("no report files were created")
	}

	// Check filename format
	found := false
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "skip2-go-qrcode_gozxing_") && strings.HasSuffix(name, ".md") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected report file with pattern 'skip2-go-qrcode_gozxing_*.md' not found")
	}
}

func TestGenerateReport_AllPassing(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := createAllPassingResults()

	content := reporter.generateReport(results, "test-encoder", "test-decoder")

	// Check header
	if !strings.Contains(content, "# QR Compatibility Report") {
		t.Errorf("report missing title")
	}
	if !strings.Contains(content, "**Encoder:** test-encoder") {
		t.Errorf("report missing encoder name")
	}
	if !strings.Contains(content, "**Decoder:** test-decoder") {
		t.Errorf("report missing decoder name")
	}

	// Check summary
	if !strings.Contains(content, "## Summary") {
		t.Errorf("report missing summary section")
	}
	if !strings.Contains(content, "Successful:** 4 (100.0%)") {
		t.Errorf("report has incorrect success rate for all passing tests")
	}
	if !strings.Contains(content, "Failed:** 0 (0.0%)") {
		t.Errorf("report has incorrect failure count for all passing tests")
	}

	// Check matrix
	if !strings.Contains(content, "## Compatibility Matrix") {
		t.Errorf("report missing compatibility matrix section")
	}

	// Check failure analysis
	if !strings.Contains(content, "No failures detected") {
		t.Errorf("report should indicate no failures for all passing tests")
	}
}

func TestGenerateReport_AllFailing(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := createAllFailingResults()

	content := reporter.generateReport(results, "test-encoder", "test-decoder")

	// Check summary
	if !strings.Contains(content, "Successful:** 0 (0.0%)") {
		t.Errorf("report has incorrect success rate for all failing tests")
	}
	if !strings.Contains(content, "Failed:** 4 (100.0%)") {
		t.Errorf("report has incorrect failure count for all failing tests")
	}

	// Check failure analysis
	if !strings.Contains(content, "### Failed Combinations (4)") {
		t.Errorf("report missing failure combinations section")
	}
}

func TestGenerateReport_MixedResults(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := createMixedResults()

	content := reporter.generateReport(results, "skip2/go-qrcode", "gozxing")

	// Check summary
	if !strings.Contains(content, "Successful:** 6 (75.0%)") {
		t.Errorf("report has incorrect success rate for mixed results")
	}
	if !strings.Contains(content, "Failed:** 2 (25.0%)") {
		t.Errorf("report has incorrect failure count for mixed results")
	}

	// Check failure analysis
	if !strings.Contains(content, "### Failed Combinations (2)") {
		t.Errorf("report should show 2 failed combinations")
	}
}

func TestBuild2DMatrix_Formatting(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := createMatrixResults()

	matrix := reporter.build2DMatrix(results)

	// Check header row
	if !strings.Contains(matrix, "Bytes\\Px") {
		t.Errorf("matrix missing header label")
	}
	if !strings.Contains(matrix, " 320") {
		t.Errorf("matrix missing pixel size 320")
	}
	if !strings.Contains(matrix, " 480") {
		t.Errorf("matrix missing pixel size 480")
	}

	// Check data rows
	if !strings.Contains(matrix, " 500") {
		t.Errorf("matrix missing data size 500")
	}
	if !strings.Contains(matrix, " 600") {
		t.Errorf("matrix missing data size 600")
	}

	// Check symbols
	if !strings.Contains(matrix, "✓") {
		t.Errorf("matrix missing success symbol")
	}
	if !strings.Contains(matrix, "✗") {
		t.Errorf("matrix missing failure symbol")
	}

	// Check code block
	if !strings.HasPrefix(matrix, "```\n") {
		t.Errorf("matrix should start with code block marker")
	}
	if !strings.HasSuffix(strings.TrimSpace(matrix), "```") {
		t.Errorf("matrix should end with code block marker")
	}
}

func TestBuildSummary_Statistics(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := []matrix.TestResult{
		{Success: true, DataMatches: true, EncodeTime: 10 * time.Millisecond, DecodeTime: 5 * time.Millisecond},
		{Success: true, DataMatches: true, EncodeTime: 20 * time.Millisecond, DecodeTime: 10 * time.Millisecond},
		{Success: false, DataMatches: false, EncodeTime: 15 * time.Millisecond, DecodeTime: 8 * time.Millisecond},
	}

	summary := reporter.buildSummary(results)

	if !strings.Contains(summary, "Total Tests:** 3") {
		t.Errorf("summary has incorrect total count")
	}
	if !strings.Contains(summary, "Successful:** 2 (66.7%)") {
		t.Errorf("summary has incorrect success statistics")
	}
	if !strings.Contains(summary, "Failed:** 1 (33.3%)") {
		t.Errorf("summary has incorrect failure statistics")
	}

	// Check average times (15ms encode, 7.7ms decode)
	if !strings.Contains(summary, "Average Encode Time:") {
		t.Errorf("summary missing average encode time")
	}
	if !strings.Contains(summary, "Average Decode Time:") {
		t.Errorf("summary missing average decode time")
	}
}

func TestBuildModuleInfo_WithoutVersionDetection(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := []matrix.TestResult{
		{QRVersion: -1, ModuleCount: 0}, // Version detection not implemented
		{QRVersion: -1, ModuleCount: 0},
	}

	info := reporter.buildModuleInfo(results)

	if !strings.Contains(info, "Module size information unavailable") {
		t.Errorf("should indicate module info is unavailable when version detection not implemented")
	}
	if !strings.Contains(info, "QR version detection not yet implemented") {
		t.Errorf("should explain why module info is unavailable")
	}
}

func TestBuildModuleInfo_WithVersionDetection(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := []matrix.TestResult{
		{
			PixelSize:          320,
			QRVersion:          10,
			ModuleCount:        57,
			ModulePixelSize:    5.25,
			IsFractionalModule: true,
			Success:            true,
			DataMatches:        true,
		},
		{
			PixelSize:          480,
			QRVersion:          10,
			ModuleCount:        57,
			ModulePixelSize:    7.87,
			IsFractionalModule: true,
			Success:            false,
			DataMatches:        false,
		},
	}

	info := reporter.buildModuleInfo(results)

	// Check table structure
	if !strings.Contains(info, "| Pixel Size | Module Pixel Size | Type | Status |") {
		t.Errorf("module info missing table header")
	}

	// Check data
	if !strings.Contains(info, "320px") {
		t.Errorf("module info missing 320px entry")
	}
	if !strings.Contains(info, "5.25 px/module") {
		t.Errorf("module info missing calculated module pixel size")
	}
	if !strings.Contains(info, "Fractional") {
		t.Errorf("module info should indicate fractional module sizes")
	}

	// Check status indicators
	if !strings.Contains(info, "✓ Working") {
		t.Errorf("module info should show working status for successful tests")
	}
	if !strings.Contains(info, "⚠️ Problematic") {
		t.Errorf("module info should show problematic status for failed tests")
	}

	// Check note about fractional modules
	if !strings.Contains(info, "All tested pixel sizes produce fractional module sizes") {
		t.Errorf("module info should note when all sizes are fractional")
	}
}

func TestBuildFailureAnalysis_Patterns(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := []matrix.TestResult{
		{PixelSize: 440, DataSize: 500, Success: false, DataMatches: false},
		{PixelSize: 440, DataSize: 550, Success: false, DataMatches: false},
		{PixelSize: 480, DataSize: 500, Success: true, DataMatches: true},
		{PixelSize: 480, DataSize: 550, Success: true, DataMatches: true},
	}

	analysis := reporter.buildFailureAnalysis(results)

	// Check failure patterns
	if !strings.Contains(analysis, "Pixel size 440px: 2/2 failures (100.0%)") {
		t.Errorf("failure analysis missing pixel size pattern")
	}
}

func TestBuildTimingAnalysis(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())
	results := []matrix.TestResult{
		{DataSize: 500, EncodeTime: 10 * time.Millisecond, DecodeTime: 5 * time.Millisecond},
		{DataSize: 500, EncodeTime: 12 * time.Millisecond, DecodeTime: 6 * time.Millisecond},
		{DataSize: 600, EncodeTime: 15 * time.Millisecond, DecodeTime: 8 * time.Millisecond},
	}

	timing := reporter.buildTimingAnalysis(results)

	// Check table structure
	if !strings.Contains(timing, "| Data Size | Avg Encode | Avg Decode |") {
		t.Errorf("timing analysis missing table header")
	}

	// Check data sizes
	if !strings.Contains(timing, "500 bytes") {
		t.Errorf("timing analysis missing 500 bytes entry")
	}
	if !strings.Contains(timing, "600 bytes") {
		t.Errorf("timing analysis missing 600 bytes entry")
	}

	// Check averages (500 bytes: 11ms encode, 5.5ms decode)
	if !strings.Contains(timing, "11.0ms") {
		t.Errorf("timing analysis has incorrect average encode time for 500 bytes")
	}
	if !strings.Contains(timing, "5.5ms") {
		t.Errorf("timing analysis has incorrect average decode time for 500 bytes")
	}
}

func TestHasNonMonotonicFailures(t *testing.T) {
	reporter := NewMarkdownReporter(t.TempDir())

	tests := []struct {
		name     string
		results  []matrix.TestResult
		expected bool
	}{
		{
			name: "monotonic failures",
			results: []matrix.TestResult{
				{DataSize: 500, PixelSize: 320, Success: true, DataMatches: true},
				{DataSize: 500, PixelSize: 400, Success: false, DataMatches: false},
				{DataSize: 500, PixelSize: 480, Success: false, DataMatches: false},
			},
			expected: false,
		},
		{
			name: "non-monotonic failures",
			results: []matrix.TestResult{
				{DataSize: 500, PixelSize: 320, Success: true, DataMatches: true},
				{DataSize: 500, PixelSize: 400, Success: false, DataMatches: false},
				{DataSize: 500, PixelSize: 480, Success: true, DataMatches: true},
			},
			expected: true,
		},
		{
			name: "all passing",
			results: []matrix.TestResult{
				{DataSize: 500, PixelSize: 320, Success: true, DataMatches: true},
				{DataSize: 500, PixelSize: 400, Success: true, DataMatches: true},
				{DataSize: 500, PixelSize: 480, Success: true, DataMatches: true},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reporter.hasNonMonotonicFailures(tt.results)
			if result != tt.expected {
				t.Errorf("hasNonMonotonicFailures() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"skip2/go-qrcode", "skip2-go-qrcode"},
		{"github.com/makiuchi-d/gozxing", "github.com-makiuchi-d-gozxing"},
		{"simple-name", "simple-name"},
		{"name with spaces", "name_with_spaces"},
		{"path\\windows", "path-windows"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{10 * time.Millisecond, "10.0ms"},
		{12500 * time.Microsecond, "12.5ms"},
		{1 * time.Second, "1000.0ms"},
		{500 * time.Microsecond, "0.5ms"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
			}
		})
	}
}

// Helper functions to create test data

func createSampleMatrix() *matrix.CompatibilityMatrix {
	return &matrix.CompatibilityMatrix{
		Results: []matrix.TestResult{
			{
				EncoderName: "skip2/go-qrcode",
				DecoderName: "gozxing",
				DataSize:    500,
				PixelSize:   440,
				Success:     true,
				DataMatches: true,
			},
			{
				EncoderName: "skip2/go-qrcode",
				DecoderName: "gozxing",
				DataSize:    500,
				PixelSize:   480,
				Success:     false,
				DataMatches: false,
			},
		},
		Encoders:   []string{"skip2/go-qrcode"},
		Decoders:   []string{"gozxing"},
		DataSizes:  []int{500},
		PixelSizes: []int{440, 480},
	}
}

func createAllPassingResults() []matrix.TestResult {
	return []matrix.TestResult{
		{DataSize: 500, PixelSize: 320, Success: true, DataMatches: true, EncodeTime: 10 * time.Millisecond, DecodeTime: 5 * time.Millisecond},
		{DataSize: 500, PixelSize: 480, Success: true, DataMatches: true, EncodeTime: 11 * time.Millisecond, DecodeTime: 6 * time.Millisecond},
		{DataSize: 600, PixelSize: 320, Success: true, DataMatches: true, EncodeTime: 12 * time.Millisecond, DecodeTime: 7 * time.Millisecond},
		{DataSize: 600, PixelSize: 480, Success: true, DataMatches: true, EncodeTime: 13 * time.Millisecond, DecodeTime: 8 * time.Millisecond},
	}
}

func createAllFailingResults() []matrix.TestResult {
	return []matrix.TestResult{
		{DataSize: 500, PixelSize: 320, Success: false, DataMatches: false, EncodeTime: 10 * time.Millisecond, DecodeTime: 5 * time.Millisecond},
		{DataSize: 500, PixelSize: 480, Success: false, DataMatches: false, EncodeTime: 11 * time.Millisecond, DecodeTime: 6 * time.Millisecond},
		{DataSize: 600, PixelSize: 320, Success: false, DataMatches: false, EncodeTime: 12 * time.Millisecond, DecodeTime: 7 * time.Millisecond},
		{DataSize: 600, PixelSize: 480, Success: false, DataMatches: false, EncodeTime: 13 * time.Millisecond, DecodeTime: 8 * time.Millisecond},
	}
}

func createMixedResults() []matrix.TestResult {
	return []matrix.TestResult{
		{DataSize: 500, PixelSize: 320, Success: true, DataMatches: true, EncodeTime: 10 * time.Millisecond, DecodeTime: 5 * time.Millisecond},
		{DataSize: 500, PixelSize: 440, Success: false, DataMatches: false, EncodeTime: 11 * time.Millisecond, DecodeTime: 6 * time.Millisecond},
		{DataSize: 500, PixelSize: 480, Success: true, DataMatches: true, EncodeTime: 12 * time.Millisecond, DecodeTime: 7 * time.Millisecond},
		{DataSize: 550, PixelSize: 320, Success: true, DataMatches: true, EncodeTime: 11 * time.Millisecond, DecodeTime: 5 * time.Millisecond},
		{DataSize: 550, PixelSize: 440, Success: false, DataMatches: false, EncodeTime: 12 * time.Millisecond, DecodeTime: 6 * time.Millisecond},
		{DataSize: 550, PixelSize: 480, Success: true, DataMatches: true, EncodeTime: 13 * time.Millisecond, DecodeTime: 7 * time.Millisecond},
		{DataSize: 600, PixelSize: 320, Success: true, DataMatches: true, EncodeTime: 12 * time.Millisecond, DecodeTime: 6 * time.Millisecond},
		{DataSize: 600, PixelSize: 480, Success: true, DataMatches: true, EncodeTime: 14 * time.Millisecond, DecodeTime: 8 * time.Millisecond},
	}
}

func createMatrixResults() []matrix.TestResult {
	return []matrix.TestResult{
		{DataSize: 500, PixelSize: 320, Success: true, DataMatches: true},
		{DataSize: 500, PixelSize: 480, Success: false, DataMatches: false},
		{DataSize: 600, PixelSize: 320, Success: true, DataMatches: true},
		{DataSize: 600, PixelSize: 480, Success: true, DataMatches: true},
	}
}
