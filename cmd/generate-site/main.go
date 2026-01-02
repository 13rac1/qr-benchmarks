package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RawTestResult matches the JSON structure from pkg/report/json.go
type RawTestResult struct {
	Encoder            string  `json:"encoder"`
	Decoder            string  `json:"decoder"`
	DataSize           int     `json:"dataSize"`
	PixelSize          int     `json:"pixelSize"`
	ContentType        string  `json:"contentType"`
	Success            bool    `json:"success"`
	ErrorType          string  `json:"errorType,omitempty"`
	ErrorMsg           string  `json:"errorMsg,omitempty"`
	IsCapacityExceeded bool    `json:"isCapacityExceeded,omitempty"`
	EncodeTimeMs       float64 `json:"encodeTimeMs"`
	DecodeTimeMs       float64 `json:"decodeTimeMs"`
	QRVersion          int     `json:"qrVersion,omitempty"`
	ModuleCount        int     `json:"moduleCount,omitempty"`
	ModulePixelSize    float64 `json:"modulePixelSize,omitempty"`
	IsFractionalModule bool    `json:"isFractionalModule,omitempty"`
}

type RawResults struct {
	Timestamp string          `json:"timestamp"`
	Results   []RawTestResult `json:"results"`
}

// Output structures for Hugo

type DecoderBreakdown struct {
	SuccessRate    float64 `json:"successRate"`
	Tests          int     `json:"tests"`
	Successes      int     `json:"successes"`
	CapacitySkips  int     `json:"capacitySkips"`
	EffectiveTests int     `json:"effectiveTests"` // Tests - CapacitySkips
}

type EncoderStats struct {
	Name           string                      `json:"name"`
	SuccessRate    float64                     `json:"successRate"`
	AvgEncodeMs    float64                     `json:"avgEncodeMs"`
	TotalTests     int                         `json:"totalTests"`
	SuccessCount   int                         `json:"successCount"`
	CapacitySkips  int                         `json:"capacitySkips"`
	EffectiveTests int                         `json:"effectiveTests"` // TotalTests - CapacitySkips
	ByDecoder      map[string]DecoderBreakdown `json:"byDecoder"`
}

type EncoderBreakdown struct {
	SuccessRate    float64 `json:"successRate"`
	Tests          int     `json:"tests"`
	Successes      int     `json:"successes"`
	CapacitySkips  int     `json:"capacitySkips"`
	EffectiveTests int     `json:"effectiveTests"` // Tests - CapacitySkips
}

type DecoderStats struct {
	Name           string                      `json:"name"`
	SuccessRate    float64                     `json:"successRate"`
	AvgDecodeMs    float64                     `json:"avgDecodeMs"`
	TotalTests     int                         `json:"totalTests"`
	SuccessCount   int                         `json:"successCount"`
	CapacitySkips  int                         `json:"capacitySkips"`
	EffectiveTests int                         `json:"effectiveTests"` // TotalTests - CapacitySkips
	ByEncoder      map[string]EncoderBreakdown `json:"byEncoder"`
}

type CombinationResult struct {
	Encoder        string  `json:"encoder"`
	Decoder        string  `json:"decoder"`
	SuccessRate    float64 `json:"successRate"`
	Tests          int     `json:"tests"`
	Successes      int     `json:"successes"`
	CapacitySkips  int     `json:"capacitySkips"`
	EffectiveTests int     `json:"effectiveTests"`
	AvgEncodeMs    float64 `json:"avgEncodeMs"`
	AvgDecodeMs    float64 `json:"avgDecodeMs"`
}

type BestCombination struct {
	Encoder     string  `json:"encoder"`
	Decoder     string  `json:"decoder"`
	SuccessRate float64 `json:"successRate"`
}

type CombinationsData struct {
	Matrix []CombinationResult `json:"matrix"`
	Best   BestCombination     `json:"best"`
}

type FailuresByType struct {
	Encode       int `json:"encode"`
	Decode       int `json:"decode"`
	DataMismatch int `json:"dataMismatch"`
}

type ConditionFailures struct {
	Condition string  `json:"condition"`
	Failures  int     `json:"failures"`
	Total     int     `json:"total"`
	Rate      float64 `json:"rate"`
}

type FailuresData struct {
	ByType           FailuresByType      `json:"byType"`
	ByDataSize       []ConditionFailures `json:"byDataSize"`
	ByPixelSize      []ConditionFailures `json:"byPixelSize"`
	ByContentType    []ConditionFailures `json:"byContentType"`
	FractionalModule ConditionFailures   `json:"fractionalModule"`
	IntegerModule    ConditionFailures   `json:"integerModule"`
}

type SummaryData struct {
	Timestamp       string          `json:"timestamp"`
	TotalTests      int             `json:"totalTests"`
	TotalSuccesses  int             `json:"totalSuccesses"`
	CapacitySkips   int             `json:"capacitySkips"`
	EffectiveTests  int             `json:"effectiveTests"`
	OverallRate     float64         `json:"overallRate"`
	BestEncoder     string          `json:"bestEncoder"`
	BestDecoder     string          `json:"bestDecoder"`
	BestCombination BestCombination `json:"bestCombination"`
	EncoderCount    int             `json:"encoderCount"`
	DecoderCount    int             `json:"decoderCount"`
}

func main() {
	resultsDir := "results"
	outputDir := "website/data"

	if len(os.Args) > 1 {
		resultsDir = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	results, err := loadAllResults(resultsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading results: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "No results found in %s\n", resultsDir)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d test results\n", len(results))

	encoders := computeEncoderStats(results)
	decoders := computeDecoderStats(results)
	combinations := computeCombinations(results)
	failures := computeFailures(results)
	summary := computeSummary(results, encoders, decoders, combinations)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if err := writeJSON(filepath.Join(outputDir, "encoders.json"), encoders); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing encoders.json: %v\n", err)
		os.Exit(1)
	}

	if err := writeJSON(filepath.Join(outputDir, "decoders.json"), decoders); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing decoders.json: %v\n", err)
		os.Exit(1)
	}

	if err := writeJSON(filepath.Join(outputDir, "combinations.json"), combinations); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing combinations.json: %v\n", err)
		os.Exit(1)
	}

	if err := writeJSON(filepath.Join(outputDir, "failures.json"), failures); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing failures.json: %v\n", err)
		os.Exit(1)
	}

	if err := writeJSON(filepath.Join(outputDir, "summary.json"), summary); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing summary.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated Hugo data files in %s\n", outputDir)
}

func loadAllResults(dir string) ([]RawTestResult, error) {
	var allResults []RawTestResult

	// Load from encoders directory
	encodersDir := filepath.Join(dir, "encoders")
	if err := loadResultsFromDir(encodersDir, &allResults); err != nil {
		return nil, err
	}

	// Deduplicate (since we only need one copy of each result)
	seen := make(map[string]bool)
	var unique []RawTestResult
	for _, r := range allResults {
		key := fmt.Sprintf("%s|%s|%d|%d|%s", r.Encoder, r.Decoder, r.DataSize, r.PixelSize, r.ContentType)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, r)
		}
	}

	return unique, nil
}

func loadResultsFromDir(dir string, results *[]RawTestResult) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		var raw RawResults
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		*results = append(*results, raw.Results...)
	}

	return nil
}

func computeEncoderStats(results []RawTestResult) []EncoderStats {
	type encoderAgg struct {
		totalTests    int
		successes     int
		capacitySkips int
		totalEncMs    float64
		byDecoder     map[string]*struct{ tests, successes, capacitySkips int }
	}

	agg := make(map[string]*encoderAgg)

	for _, r := range results {
		if agg[r.Encoder] == nil {
			agg[r.Encoder] = &encoderAgg{
				byDecoder: make(map[string]*struct{ tests, successes, capacitySkips int }),
			}
		}
		a := agg[r.Encoder]
		a.totalTests++
		a.totalEncMs += r.EncodeTimeMs
		if r.Success {
			a.successes++
		}
		if r.IsCapacityExceeded {
			a.capacitySkips++
		}

		if a.byDecoder[r.Decoder] == nil {
			a.byDecoder[r.Decoder] = &struct{ tests, successes, capacitySkips int }{}
		}
		a.byDecoder[r.Decoder].tests++
		if r.Success {
			a.byDecoder[r.Decoder].successes++
		}
		if r.IsCapacityExceeded {
			a.byDecoder[r.Decoder].capacitySkips++
		}
	}

	var stats []EncoderStats
	for name, a := range agg {
		byDec := make(map[string]DecoderBreakdown)
		for dec, d := range a.byDecoder {
			effectiveTests := d.tests - d.capacitySkips
			rate := 0.0
			if effectiveTests > 0 {
				rate = float64(d.successes) / float64(effectiveTests) * 100
			}
			byDec[dec] = DecoderBreakdown{
				SuccessRate:    rate,
				Tests:          d.tests,
				Successes:      d.successes,
				CapacitySkips:  d.capacitySkips,
				EffectiveTests: effectiveTests,
			}
		}

		effectiveTests := a.totalTests - a.capacitySkips
		rate := 0.0
		if effectiveTests > 0 {
			rate = float64(a.successes) / float64(effectiveTests) * 100
		}
		avgEnc := 0.0
		if a.totalTests > 0 {
			avgEnc = a.totalEncMs / float64(a.totalTests)
		}

		stats = append(stats, EncoderStats{
			Name:           name,
			SuccessRate:    rate,
			AvgEncodeMs:    avgEnc,
			TotalTests:     a.totalTests,
			SuccessCount:   a.successes,
			CapacitySkips:  a.capacitySkips,
			EffectiveTests: effectiveTests,
			ByDecoder:      byDec,
		})
	}

	// Sort by success rate descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].SuccessRate > stats[j].SuccessRate
	})

	return stats
}

func computeDecoderStats(results []RawTestResult) []DecoderStats {
	type decoderAgg struct {
		totalTests    int
		successes     int
		capacitySkips int
		totalDecMs    float64
		byEncoder     map[string]*struct{ tests, successes, capacitySkips int }
	}

	agg := make(map[string]*decoderAgg)

	for _, r := range results {
		if agg[r.Decoder] == nil {
			agg[r.Decoder] = &decoderAgg{
				byEncoder: make(map[string]*struct{ tests, successes, capacitySkips int }),
			}
		}
		a := agg[r.Decoder]
		a.totalTests++
		a.totalDecMs += r.DecodeTimeMs
		if r.Success {
			a.successes++
		}
		if r.IsCapacityExceeded {
			a.capacitySkips++
		}

		if a.byEncoder[r.Encoder] == nil {
			a.byEncoder[r.Encoder] = &struct{ tests, successes, capacitySkips int }{}
		}
		a.byEncoder[r.Encoder].tests++
		if r.Success {
			a.byEncoder[r.Encoder].successes++
		}
		if r.IsCapacityExceeded {
			a.byEncoder[r.Encoder].capacitySkips++
		}
	}

	var stats []DecoderStats
	for name, a := range agg {
		byEnc := make(map[string]EncoderBreakdown)
		for enc, e := range a.byEncoder {
			effectiveTests := e.tests - e.capacitySkips
			rate := 0.0
			if effectiveTests > 0 {
				rate = float64(e.successes) / float64(effectiveTests) * 100
			}
			byEnc[enc] = EncoderBreakdown{
				SuccessRate:    rate,
				Tests:          e.tests,
				Successes:      e.successes,
				CapacitySkips:  e.capacitySkips,
				EffectiveTests: effectiveTests,
			}
		}

		effectiveTests := a.totalTests - a.capacitySkips
		rate := 0.0
		if effectiveTests > 0 {
			rate = float64(a.successes) / float64(effectiveTests) * 100
		}
		avgDec := 0.0
		if a.totalTests > 0 {
			avgDec = a.totalDecMs / float64(a.totalTests)
		}

		stats = append(stats, DecoderStats{
			Name:           name,
			SuccessRate:    rate,
			AvgDecodeMs:    avgDec,
			TotalTests:     a.totalTests,
			SuccessCount:   a.successes,
			CapacitySkips:  a.capacitySkips,
			EffectiveTests: effectiveTests,
			ByEncoder:      byEnc,
		})
	}

	// Sort by success rate descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].SuccessRate > stats[j].SuccessRate
	})

	return stats
}

func computeCombinations(results []RawTestResult) CombinationsData {
	type combAgg struct {
		tests         int
		successes     int
		capacitySkips int
		encMs         float64
		decMs         float64
	}

	agg := make(map[string]*combAgg)

	for _, r := range results {
		key := r.Encoder + "|" + r.Decoder
		if agg[key] == nil {
			agg[key] = &combAgg{}
		}
		a := agg[key]
		a.tests++
		a.encMs += r.EncodeTimeMs
		a.decMs += r.DecodeTimeMs
		if r.Success {
			a.successes++
		}
		if r.IsCapacityExceeded {
			a.capacitySkips++
		}
	}

	var matrix []CombinationResult
	var best CombinationResult

	for key, a := range agg {
		parts := splitKey(key)
		effectiveTests := a.tests - a.capacitySkips
		rate := 0.0
		if effectiveTests > 0 {
			rate = float64(a.successes) / float64(effectiveTests) * 100
		}
		avgEnc := 0.0
		avgDec := 0.0
		if a.tests > 0 {
			avgEnc = a.encMs / float64(a.tests)
			avgDec = a.decMs / float64(a.tests)
		}

		cr := CombinationResult{
			Encoder:        parts[0],
			Decoder:        parts[1],
			SuccessRate:    rate,
			Tests:          a.tests,
			Successes:      a.successes,
			CapacitySkips:  a.capacitySkips,
			EffectiveTests: effectiveTests,
			AvgEncodeMs:    avgEnc,
			AvgDecodeMs:    avgDec,
		}
		matrix = append(matrix, cr)

		if rate > best.SuccessRate {
			best = cr
		}
	}

	// Sort by success rate descending
	sort.Slice(matrix, func(i, j int) bool {
		return matrix[i].SuccessRate > matrix[j].SuccessRate
	})

	return CombinationsData{
		Matrix: matrix,
		Best: BestCombination{
			Encoder:     best.Encoder,
			Decoder:     best.Decoder,
			SuccessRate: best.SuccessRate,
		},
	}
}

func computeFailures(results []RawTestResult) FailuresData {
	var byType FailuresByType

	dataSizeAgg := make(map[int]*struct{ failures, total int })
	pixelSizeAgg := make(map[int]*struct{ failures, total int })
	contentTypeAgg := make(map[string]*struct{ failures, total int })
	var fractionalFailures, fractionalTotal int
	var integerFailures, integerTotal int

	for _, r := range results {
		// Skip capacity exceeded - these are valid rejections, not failures
		if r.IsCapacityExceeded {
			continue
		}

		if !r.Success {
			switch r.ErrorType {
			case "encode":
				byType.Encode++
			case "decode":
				byType.Decode++
			case "dataMismatch":
				byType.DataMismatch++
			}
		}

		// By data size
		if dataSizeAgg[r.DataSize] == nil {
			dataSizeAgg[r.DataSize] = &struct{ failures, total int }{}
		}
		dataSizeAgg[r.DataSize].total++
		if !r.Success {
			dataSizeAgg[r.DataSize].failures++
		}

		// By pixel size
		if pixelSizeAgg[r.PixelSize] == nil {
			pixelSizeAgg[r.PixelSize] = &struct{ failures, total int }{}
		}
		pixelSizeAgg[r.PixelSize].total++
		if !r.Success {
			pixelSizeAgg[r.PixelSize].failures++
		}

		// By content type
		if contentTypeAgg[r.ContentType] == nil {
			contentTypeAgg[r.ContentType] = &struct{ failures, total int }{}
		}
		contentTypeAgg[r.ContentType].total++
		if !r.Success {
			contentTypeAgg[r.ContentType].failures++
		}

		// Fractional vs integer modules
		if r.IsFractionalModule {
			fractionalTotal++
			if !r.Success {
				fractionalFailures++
			}
		} else {
			integerTotal++
			if !r.Success {
				integerFailures++
			}
		}
	}

	var byDataSize []ConditionFailures
	for size, a := range dataSizeAgg {
		rate := 0.0
		if a.total > 0 {
			rate = float64(a.failures) / float64(a.total) * 100
		}
		byDataSize = append(byDataSize, ConditionFailures{
			Condition: fmt.Sprintf("%d bytes", size),
			Failures:  a.failures,
			Total:     a.total,
			Rate:      rate,
		})
	}
	sort.Slice(byDataSize, func(i, j int) bool {
		return byDataSize[i].Rate > byDataSize[j].Rate
	})

	var byPixelSize []ConditionFailures
	for size, a := range pixelSizeAgg {
		rate := 0.0
		if a.total > 0 {
			rate = float64(a.failures) / float64(a.total) * 100
		}
		byPixelSize = append(byPixelSize, ConditionFailures{
			Condition: fmt.Sprintf("%dpx", size),
			Failures:  a.failures,
			Total:     a.total,
			Rate:      rate,
		})
	}
	sort.Slice(byPixelSize, func(i, j int) bool {
		return byPixelSize[i].Rate > byPixelSize[j].Rate
	})

	var byContentType []ConditionFailures
	for ct, a := range contentTypeAgg {
		rate := 0.0
		if a.total > 0 {
			rate = float64(a.failures) / float64(a.total) * 100
		}
		byContentType = append(byContentType, ConditionFailures{
			Condition: ct,
			Failures:  a.failures,
			Total:     a.total,
			Rate:      rate,
		})
	}
	sort.Slice(byContentType, func(i, j int) bool {
		return byContentType[i].Rate > byContentType[j].Rate
	})

	fractionalRate := 0.0
	if fractionalTotal > 0 {
		fractionalRate = float64(fractionalFailures) / float64(fractionalTotal) * 100
	}
	integerRate := 0.0
	if integerTotal > 0 {
		integerRate = float64(integerFailures) / float64(integerTotal) * 100
	}

	return FailuresData{
		ByType:        byType,
		ByDataSize:    byDataSize,
		ByPixelSize:   byPixelSize,
		ByContentType: byContentType,
		FractionalModule: ConditionFailures{
			Condition: "Fractional module size",
			Failures:  fractionalFailures,
			Total:     fractionalTotal,
			Rate:      fractionalRate,
		},
		IntegerModule: ConditionFailures{
			Condition: "Integer module size",
			Failures:  integerFailures,
			Total:     integerTotal,
			Rate:      integerRate,
		},
	}
}

func computeSummary(results []RawTestResult, encoders []EncoderStats, decoders []DecoderStats, combinations CombinationsData) SummaryData {
	total := len(results)
	successes := 0
	capacitySkips := 0
	for _, r := range results {
		if r.Success {
			successes++
		}
		if r.IsCapacityExceeded {
			capacitySkips++
		}
	}

	effectiveTests := total - capacitySkips
	rate := 0.0
	if effectiveTests > 0 {
		rate = float64(successes) / float64(effectiveTests) * 100
	}

	bestEncoder := ""
	if len(encoders) > 0 {
		bestEncoder = encoders[0].Name
	}

	bestDecoder := ""
	if len(decoders) > 0 {
		bestDecoder = decoders[0].Name
	}

	return SummaryData{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		TotalTests:      total,
		TotalSuccesses:  successes,
		CapacitySkips:   capacitySkips,
		EffectiveTests:  effectiveTests,
		OverallRate:     rate,
		BestEncoder:     bestEncoder,
		BestDecoder:     bestDecoder,
		BestCombination: combinations.Best,
		EncoderCount:    len(encoders),
		DecoderCount:    len(decoders),
	}
}

func splitKey(key string) []string {
	for i := 0; i < len(key); i++ {
		if key[i] == '|' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key, ""}
}

func writeJSON(path string, data interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
