// Package report provides report generation for QR code compatibility test results.
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/13rac1/qr-library-test/internal/matrix"
)

// MarkdownReporter generates markdown reports from test results.
// Each encoder/decoder combination gets its own report file.
type MarkdownReporter struct {
	OutputDir string
}

// NewMarkdownReporter creates a new markdown reporter that writes to the specified directory.
func NewMarkdownReporter(outputDir string) *MarkdownReporter {
	return &MarkdownReporter{
		OutputDir: outputDir,
	}
}

// Generate creates markdown report files for all encoder/decoder combinations in the matrix.
// One file is created per unique encoder+decoder pair.
// Files are named: <encoder>_<decoder>_<timestamp>.md
func (r *MarkdownReporter) Generate(m *matrix.CompatibilityMatrix) error {
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Group results by encoder/decoder pair
	pairs := make(map[string][]matrix.TestResult)
	for _, result := range m.Results {
		key := result.EncoderName + "|" + result.DecoderName
		pairs[key] = append(pairs[key], result)
	}

	// Generate a report for each pair
	timestamp := time.Now().Format("20060102-150405")
	for key, results := range pairs {
		parts := strings.Split(key, "|")
		encoder := parts[0]
		decoder := parts[1]

		content := r.generateReport(results, encoder, decoder)

		// Create safe filename
		filename := fmt.Sprintf("%s_%s_%s.md",
			sanitizeFilename(encoder),
			sanitizeFilename(decoder),
			timestamp)
		filepath := filepath.Join(r.OutputDir, filename)

		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write report %s: %w", filename, err)
		}
	}

	return nil
}

// generateReport creates the markdown content for one encoder/decoder pair.
func (r *MarkdownReporter) generateReport(results []matrix.TestResult, encoder, decoder string) string {
	var b strings.Builder

	// Header
	b.WriteString("# QR Compatibility Report\n\n")
	b.WriteString(fmt.Sprintf("**Encoder:** %s  \n", encoder))
	b.WriteString(fmt.Sprintf("**Decoder:** %s  \n", decoder))
	b.WriteString(fmt.Sprintf("**Generated:** %s  \n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Summary
	b.WriteString(r.buildSummary(results))
	b.WriteString("\n")

	// 2D Matrix
	b.WriteString("## Compatibility Matrix\n\n")
	b.WriteString(r.build2DMatrix(results))
	b.WriteString("\n")
	b.WriteString("✓ = Successful decode  \n")
	b.WriteString("✗ = Failed decode\n\n")

	// Failure Analysis
	b.WriteString(r.buildFailureAnalysis(results))
	b.WriteString("\n")

	// Module Size Analysis
	b.WriteString(r.buildModuleInfo(results))
	b.WriteString("\n")

	// Timing Analysis
	b.WriteString(r.buildTimingAnalysis(results))

	return b.String()
}

// buildSummary generates the summary section with statistics.
func (r *MarkdownReporter) buildSummary(results []matrix.TestResult) string {
	var b strings.Builder

	total := len(results)
	successful := 0
	var totalEncodeTime, totalDecodeTime time.Duration

	for _, result := range results {
		if result.Success && result.DataMatches {
			successful++
		}
		totalEncodeTime += result.EncodeTime
		totalDecodeTime += result.DecodeTime
	}

	failed := total - successful
	successRate := 0.0
	if total > 0 {
		successRate = float64(successful) * 100.0 / float64(total)
	}

	avgEncodeTime := time.Duration(0)
	avgDecodeTime := time.Duration(0)
	if total > 0 {
		avgEncodeTime = totalEncodeTime / time.Duration(total)
		avgDecodeTime = totalDecodeTime / time.Duration(total)
	}

	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", total))
	b.WriteString(fmt.Sprintf("- **Successful:** %d (%.1f%%)\n", successful, successRate))
	b.WriteString(fmt.Sprintf("- **Failed:** %d (%.1f%%)\n", failed, 100.0-successRate))
	b.WriteString(fmt.Sprintf("- **Average Encode Time:** %s\n", formatDuration(avgEncodeTime)))
	b.WriteString(fmt.Sprintf("- **Average Decode Time:** %s\n", formatDuration(avgDecodeTime)))

	return b.String()
}

// build2DMatrix creates the 2D pixel size × data size matrix table.
func (r *MarkdownReporter) build2DMatrix(results []matrix.TestResult) string {
	// Collect unique data sizes and pixel sizes
	dataSizeSet := make(map[int]bool)
	pixelSizeSet := make(map[int]bool)
	for _, result := range results {
		dataSizeSet[result.DataSize] = true
		pixelSizeSet[result.PixelSize] = true
	}

	// Convert to sorted slices
	dataSizes := make([]int, 0, len(dataSizeSet))
	for size := range dataSizeSet {
		dataSizes = append(dataSizes, size)
	}
	sort.Ints(dataSizes)

	pixelSizes := make([]int, 0, len(pixelSizeSet))
	for size := range pixelSizeSet {
		pixelSizes = append(pixelSizes, size)
	}
	sort.Ints(pixelSizes)

	// Build lookup map: dataSize+pixelSize -> result
	lookup := make(map[string]*matrix.TestResult)
	for i := range results {
		key := fmt.Sprintf("%d_%d", results[i].DataSize, results[i].PixelSize)
		lookup[key] = &results[i]
	}

	var b strings.Builder

	// Header row
	b.WriteString("```\n")
	b.WriteString("Bytes\\Px ")
	for _, px := range pixelSizes {
		b.WriteString(fmt.Sprintf(" %4d", px))
	}
	b.WriteString("\n")

	// Separator
	b.WriteString("--------+")
	b.WriteString(strings.Repeat("-", len(pixelSizes)*5))
	b.WriteString("\n")

	// Data rows
	for _, dataSize := range dataSizes {
		b.WriteString(fmt.Sprintf(" %4d   |", dataSize))
		for _, pixelSize := range pixelSizes {
			key := fmt.Sprintf("%d_%d", dataSize, pixelSize)
			result := lookup[key]
			if result == nil {
				b.WriteString("     ")
			} else if result.Success && result.DataMatches {
				b.WriteString("  ✓  ")
			} else {
				b.WriteString("  ✗  ")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("```\n")
	return b.String()
}

// buildFailureAnalysis generates the failure analysis section.
func (r *MarkdownReporter) buildFailureAnalysis(results []matrix.TestResult) string {
	var b strings.Builder

	// Collect failures
	var failures []matrix.TestResult
	for _, result := range results {
		if !result.Success || !result.DataMatches {
			failures = append(failures, result)
		}
	}

	if len(failures) == 0 {
		b.WriteString("## Failure Analysis\n\n")
		b.WriteString("No failures detected. All tests passed successfully.\n")
		return b.String()
	}

	b.WriteString("## Failure Analysis\n\n")
	b.WriteString(fmt.Sprintf("### Failed Combinations (%d)\n\n", len(failures)))

	// Sort failures by data size, then pixel size
	sort.Slice(failures, func(i, j int) bool {
		if failures[i].DataSize != failures[j].DataSize {
			return failures[i].DataSize < failures[j].DataSize
		}
		return failures[i].PixelSize < failures[j].PixelSize
	})

	for i, f := range failures {
		errorMsg := "Unknown error"
		if f.Error != nil {
			errorMsg = f.Error.Error()
		}
		b.WriteString(fmt.Sprintf("%d. %d bytes @ %dpx - %s\n", i+1, f.DataSize, f.PixelSize, errorMsg))
	}

	// Analyze patterns
	b.WriteString("\n### Patterns\n\n")

	// Count failures by pixel size
	pixelFailures := make(map[int]int)
	pixelTotal := make(map[int]int)
	for _, result := range results {
		pixelTotal[result.PixelSize]++
		if !result.Success || !result.DataMatches {
			pixelFailures[result.PixelSize]++
		}
	}

	// Sort pixel sizes for consistent output
	pixelSizes := make([]int, 0, len(pixelTotal))
	for px := range pixelTotal {
		pixelSizes = append(pixelSizes, px)
	}
	sort.Ints(pixelSizes)

	for _, px := range pixelSizes {
		fails := pixelFailures[px]
		total := pixelTotal[px]
		if fails > 0 {
			rate := float64(fails) * 100.0 / float64(total)
			b.WriteString(fmt.Sprintf("- Pixel size %dpx: %d/%d failures (%.1f%%)\n", px, fails, total, rate))
		}
	}

	// Check for non-monotonic failures
	if r.hasNonMonotonicFailures(results) {
		b.WriteString("- Non-monotonic failures detected (larger sizes succeed when smaller fail)\n")
	}

	return b.String()
}

// buildModuleInfo generates the module size analysis section.
func (r *MarkdownReporter) buildModuleInfo(results []matrix.TestResult) string {
	var b strings.Builder

	b.WriteString("## Module Size Analysis\n\n")

	// Check if any results have module info
	hasModuleInfo := false
	for _, result := range results {
		if result.QRVersion > 0 {
			hasModuleInfo = true
			break
		}
	}

	if !hasModuleInfo {
		b.WriteString("Module size information unavailable (QR version detection not yet implemented).\n\n")
		b.WriteString("Module size analysis will be available after implementing QR version detection.\n")
		b.WriteString("This will help identify fractional module size issues.\n")
		return b.String()
	}

	// Group by pixel size and calculate module info
	pixelInfo := make(map[int]struct {
		modulePixelSize float64
		isFractional    bool
		hasFailure      bool
		hasSuccess      bool
	})

	for _, result := range results {
		if result.QRVersion <= 0 {
			continue
		}

		info := pixelInfo[result.PixelSize]
		info.modulePixelSize = result.ModulePixelSize
		info.isFractional = result.IsFractionalModule

		if result.Success && result.DataMatches {
			info.hasSuccess = true
		} else {
			info.hasFailure = true
		}
		pixelInfo[result.PixelSize] = info
	}

	// Sort pixel sizes
	pixelSizes := make([]int, 0, len(pixelInfo))
	for px := range pixelInfo {
		pixelSizes = append(pixelSizes, px)
	}
	sort.Ints(pixelSizes)

	// Build table
	b.WriteString("| Pixel Size | Module Pixel Size | Type | Status |\n")
	b.WriteString("|------------|-------------------|------|--------|\n")

	fractionalCount := 0
	for _, px := range pixelSizes {
		info := pixelInfo[px]

		typeStr := "Integer"
		if info.isFractional {
			typeStr = "Fractional"
			fractionalCount++
		}

		status := ""
		if info.hasFailure && info.hasSuccess {
			status = "Mixed"
		} else if info.hasFailure {
			status = "⚠️ Problematic"
		} else if info.hasSuccess {
			status = "✓ Working"
		}

		b.WriteString(fmt.Sprintf("| %dpx | %.2f px/module | %s | %s |\n",
			px, info.modulePixelSize, typeStr, status))
	}

	b.WriteString("\n")

	if fractionalCount == len(pixelSizes) {
		b.WriteString("Note: All tested pixel sizes produce fractional module sizes.  \n")
		b.WriteString("This is the root cause of encoder/decoder incompatibility.\n")
	} else if fractionalCount > 0 {
		b.WriteString(fmt.Sprintf("Note: %d of %d tested pixel sizes produce fractional module sizes.  \n",
			fractionalCount, len(pixelSizes)))
		b.WriteString("Fractional module sizes can cause decode failures with certain encoder/decoder combinations.\n")
	}

	return b.String()
}

// buildTimingAnalysis generates the timing analysis section.
func (r *MarkdownReporter) buildTimingAnalysis(results []matrix.TestResult) string {
	var b strings.Builder

	b.WriteString("## Timing Analysis\n\n")

	// Group by data size
	timings := make(map[int]struct {
		encodeTotal time.Duration
		decodeTotal time.Duration
		count       int
	})

	for _, result := range results {
		t := timings[result.DataSize]
		t.encodeTotal += result.EncodeTime
		t.decodeTotal += result.DecodeTime
		t.count++
		timings[result.DataSize] = t
	}

	// Sort data sizes
	dataSizes := make([]int, 0, len(timings))
	for size := range timings {
		dataSizes = append(dataSizes, size)
	}
	sort.Ints(dataSizes)

	// Build table
	b.WriteString("| Data Size | Avg Encode | Avg Decode |\n")
	b.WriteString("|-----------|------------|------------|\n")

	for _, size := range dataSizes {
		t := timings[size]
		avgEncode := t.encodeTotal / time.Duration(t.count)
		avgDecode := t.decodeTotal / time.Duration(t.count)

		b.WriteString(fmt.Sprintf("| %d bytes | %s | %s |\n",
			size, formatDuration(avgEncode), formatDuration(avgDecode)))
	}

	return b.String()
}

// hasNonMonotonicFailures checks if there are non-monotonic failure patterns.
// Returns true if a smaller pixel size succeeds but a larger one fails for the same data size.
func (r *MarkdownReporter) hasNonMonotonicFailures(results []matrix.TestResult) bool {
	// Group by data size
	byDataSize := make(map[int][]matrix.TestResult)
	for _, result := range results {
		byDataSize[result.DataSize] = append(byDataSize[result.DataSize], result)
	}

	// Check each data size for non-monotonic pattern
	for _, group := range byDataSize {
		// Sort by pixel size
		sort.Slice(group, func(i, j int) bool {
			return group[i].PixelSize < group[j].PixelSize
		})

		// Look for pattern: success followed by failure followed by success
		for i := 1; i < len(group)-1; i++ {
			prevSuccess := group[i-1].Success && group[i-1].DataMatches
			currFail := !group[i].Success || !group[i].DataMatches
			nextSuccess := group[i+1].Success && group[i+1].DataMatches

			if prevSuccess && currFail && nextSuccess {
				return true
			}
		}
	}

	return false
}

// sanitizeFilename converts a library name to a safe filename component.
// Replaces slashes and other problematic characters.
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

// formatDuration formats a duration as milliseconds with one decimal place.
func formatDuration(d time.Duration) string {
	ms := float64(d.Microseconds()) / 1000.0
	return fmt.Sprintf("%.1fms", ms)
}
