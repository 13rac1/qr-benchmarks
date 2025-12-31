// Package testdata provides test data generation and QR code module calculation utilities.
package testdata

import (
	"errors"
	"image"
	"math"
)

// Standard quiet zone size in modules (white border around QR code).
// QR code specification requires minimum 4 modules.
const QuietZoneModules = 4

// DetectQRVersion detects the QR code version from an encoded image.
// QR versions range from 1 to 40, determining the module count.
//
// This is currently a placeholder implementation. Proper version detection
// requires analyzing the encoded image to identify version information patterns.
// For initial testing, QR version can be inferred from data size or calculated
// by the runner after encoding.
//
// Returns -1 and an error indicating detection is not yet implemented.
func DetectQRVersion(img image.Image) (int, error) {
	if img == nil {
		return -1, errors.New("image is nil")
	}
	return -1, errors.New("QR version detection not yet implemented")
}

// CalculateModuleCount returns the number of modules per side for a QR version.
// QR code module count follows the formula: 17 + 4×version.
//
// Examples:
//   - Version 1: 21 modules (17 + 4×1)
//   - Version 10: 57 modules (17 + 4×10)
//   - Version 40: 177 modules (17 + 4×40)
//
// Returns 0 for invalid versions (must be 1-40).
func CalculateModuleCount(version int) int {
	if version < 1 || version > 40 {
		return 0
	}
	return 17 + 4*version
}

// CalculateModulePixelSize calculates the pixel dimension per module.
// This value determines whether an encoder uses fractional or integer module sizing.
//
// Formula: pixelSize / (moduleCount + quietZone)
//
// The quiet zone is the white border around the QR code. The QR specification
// requires a minimum quiet zone of 4 modules on all sides.
//
// Example:
//   - Version 15 (77 modules) at 440px with 4 module quiet zone:
//     440 / (77 + 4) = 5.43 pixels/module (fractional - problematic!)
//   - Version 15 (77 modules) at 480px with 4 module quiet zone:
//     480 / (77 + 4) = 5.93 pixels/module (fractional - problematic!)
//   - Version 10 (57 modules) at 320px with 4 module quiet zone:
//     320 / (57 + 4) = 5.25 pixels/module (fractional)
//
// Fractional module pixel sizes are a known source of decode failures with
// certain encoder/decoder combinations (notably skip2 + gozxing).
func CalculateModulePixelSize(pixelSize, moduleCount, quietZone int) float64 {
	if pixelSize <= 0 || moduleCount <= 0 || quietZone < 0 {
		return 0
	}
	return float64(pixelSize) / float64(moduleCount+quietZone)
}

// IsFractionalModuleSize checks whether a module pixel size is fractional.
// Returns true if the module pixel size has a non-zero fractional component.
//
// Fractional module sizes occur when the image pixel size does not divide evenly
// by the total module count (including quiet zone). This forces the encoder to
// make rounding decisions, which may not align with decoder expectations.
//
// Examples:
//   - 5.0 pixels/module: false (integer)
//   - 5.43 pixels/module: true (fractional)
//   - 6.0 pixels/module: false (integer)
func IsFractionalModuleSize(modulePixelSize float64) bool {
	return modulePixelSize != math.Floor(modulePixelSize)
}

// CalculateOptimalPixelSize finds the smallest pixel size that results in
// integer module dimensions for the given QR version.
//
// This is useful for identifying "safe" pixel sizes that avoid fractional
// module sizing issues. The optimal pixel size is a multiple of (moduleCount + quietZone).
//
// The function finds the smallest multiple that meets a minimum practical size (100px).
//
// Example:
//   - Version 15 (77 modules) with 4 module quiet zone:
//     moduleCount + quietZone = 81
//     Optimal sizes: 81, 162, 243, 324, 405, 486...
//     First ≥ 100px: 162
//
// Returns 0 if moduleCount or quietZone are invalid.
func CalculateOptimalPixelSize(moduleCount, quietZone int) int {
	if moduleCount <= 0 || quietZone < 0 {
		return 0
	}

	totalModules := moduleCount + quietZone
	minSize := 100

	// Find the smallest multiple of totalModules that is at least minSize
	multiplier := (minSize + totalModules - 1) / totalModules
	return totalModules * multiplier
}
