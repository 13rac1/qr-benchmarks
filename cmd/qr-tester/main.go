// qr-tester executes QR code compatibility matrix testing.
//
// This CLI tool tests encoder/decoder compatibility by running all combinations
// of QR encoders and decoders against a matrix of test cases. Results are
// generated as JSON reports showing which combinations work correctly.
//
// Usage:
//
//	qr-tester [flags]
//
// Examples:
//
//	# Run with default settings (all decoders)
//	qr-tester
//
//	# Run tests in parallel with custom output directory
//	qr-tester -parallel=true -output=./test-results
//
//	# Skip CGO decoders and archived libraries
//	qr-tester -skip-cgo=true -skip-archived=true
//
//	# Run with custom test parameters
//	qr-tester -data-sizes=500,600,700 -pixel-sizes=320,480,640
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/13rac1/qr-library-test/internal/config"
	"github.com/13rac1/qr-library-test/internal/decoders"
	"github.com/13rac1/qr-library-test/internal/encoders"
	"github.com/13rac1/qr-library-test/internal/matrix"
	"github.com/13rac1/qr-library-test/internal/testdata"
	"github.com/13rac1/qr-library-test/pkg/report"
)

const version = "1.0.0"

func main() {
	// Check for version flag
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-version", "--version":
			fmt.Printf("qr-tester v%s\n", version)
			return
		}
	}

	// Register flags
	fs := flag.NewFlagSet("qr-tester", flag.ExitOnError)
	cfg, parse := config.RegisterFlags(fs)

	// Add version flag
	showVersion := fs.Bool("version", false, "Print version and exit")

	// Parse flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Flag parse error: %v", err)
	}

	// Handle version
	if *showVersion {
		fmt.Printf("qr-tester v%s\n", version)
		os.Exit(0)
	}

	// Parse config from flags
	if err := parse(); err != nil {
		log.Fatalf("Config parse error: %v", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation error: %v", err)
	}

	// Run tests
	if err := run(cfg); err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}
}

// run executes the complete test matrix and generates reports.
func run(cfg *config.Config) error {
	// Setup encoders (based on config flags)
	encs := encoders.GetAvailableEncoders(cfg)

	// Setup decoders (based on config flags)
	decs := decoders.GetAvailableDecoders(cfg)
	if len(decs) == 0 {
		return fmt.Errorf("no decoders available (check CGO build and skip flags)")
	}

	// Generate test data based on test mode
	var testCases []testdata.TestCase
	switch cfg.TestMode {
	case "comprehensive":
		testCases = testdata.GenerateComprehensiveMatrix()
	case "standard":
		fallthrough
	default:
		testCases = testdata.GeneratePixelSizeMatrix()
	}

	// Create runner
	runner := matrix.NewRunner(cfg, encs, decs, testCases)

	// Calculate and display test count
	totalTests := len(encs) * len(decs) * len(testCases)
	fmt.Printf("Running %d test combinations (%s mode)...\n", totalTests, cfg.TestMode)
	fmt.Printf("  Encoders: %d\n", len(encs))
	fmt.Printf("  Decoders: %d\n", len(decs))
	fmt.Printf("  Test cases: %d\n\n", len(testCases))

	// Run all tests
	results, err := runner.RunAll()
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Generate JSON report
	reporter := report.NewJSONReporter(cfg.OutputDir)
	if err := reporter.Generate(results); err != nil {
		return fmt.Errorf("json report failed: %w", err)
	}

	fmt.Printf("Results written to %s/\n", cfg.OutputDir)
	return nil
}
