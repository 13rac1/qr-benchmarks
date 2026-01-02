// Package encoders provides QR code encoder implementations.
package encoders

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// GozxingEncoder wraps github.com/makiuchi-d/gozxing encoder for QR code generation.
// This encoder uses the gozxing library's QRCodeWriter to generate QR codes.
type GozxingEncoder struct{}

// Name returns the encoder identifier.
func (e *GozxingEncoder) Name() string {
	return "makiuchi-d/gozxing"
}

// Encode generates a QR code image from the input data.
// The gozxing library generates a BitMatrix which is converted to image.Image.
func (e *GozxingEncoder) Encode(data []byte, opts EncodeOptions) (EncodeResult, error) {
	if len(data) == 0 {
		return EncodeResult{}, fmt.Errorf("gozxing: cannot encode empty data")
	}

	// Map error correction level to hint value
	var levelString string
	switch opts.ErrorCorrectionLevel {
	case ErrorCorrectionL:
		levelString = "L"
	case ErrorCorrectionM:
		levelString = "M"
	case ErrorCorrectionQ:
		levelString = "Q"
	case ErrorCorrectionH:
		levelString = "H"
	default:
		return EncodeResult{}, fmt.Errorf("gozxing: invalid error correction level %q", opts.ErrorCorrectionLevel)
	}

	// Create encoding hints
	hints := make(map[gozxing.EncodeHintType]interface{})
	hints[gozxing.EncodeHintType_ERROR_CORRECTION] = levelString

	// First encode at minimal size to detect QR version
	// The gozxing writer scales the QR to pixel size, so we need to encode
	// at module size first to get accurate version detection
	writer := qrcode.NewQRCodeWriter()
	minMatrix, err := writer.Encode(string(data), gozxing.BarcodeFormat_QR_CODE,
		100, 100, hints)
	if err != nil {
		return EncodeResult{}, fmt.Errorf("gozxing: encode failed: %w", err)
	}

	// Calculate version from minimal BitMatrix dimension
	// Gozxing formula: dimension = version*4 + 17
	// Inverse: version = (dimension - 17) / 4
	minDimension := minMatrix.GetWidth()
	version := (minDimension - 17) / 4

	// Now encode at requested pixel size for final image
	bitMatrix, err := writer.Encode(string(data), gozxing.BarcodeFormat_QR_CODE,
		opts.PixelSize, opts.PixelSize, hints)
	if err != nil {
		return EncodeResult{}, fmt.Errorf("gozxing: encode failed: %w", err)
	}

	// Convert BitMatrix to image.Image
	img := bitMatrixToImage(bitMatrix)

	return EncodeResult{
		Image:   img,
		Version: version,
	}, nil
}

// bitMatrixToImage converts a gozxing BitMatrix to an image.Gray.
// Black pixels (true bits) are set to 0, white pixels (false bits) to 255.
func bitMatrixToImage(matrix *gozxing.BitMatrix) image.Image {
	width := matrix.GetWidth()
	height := matrix.GetHeight()

	img := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if matrix.Get(x, y) {
				img.SetGray(x, y, color.Gray{Y: 0}) // Black
			} else {
				img.SetGray(x, y, color.Gray{Y: 255}) // White
			}
		}
	}

	return img
}

// IsCapacityError returns true if the error indicates data exceeds QR capacity.
func (e *GozxingEncoder) IsCapacityError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Data too big")
}
