// Package encoders provides QR code encoder implementations.
package encoders

import (
	"fmt"
	"image"
	"image/color"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// GozxingEncoder wraps github.com/makiuchi-d/gozxing encoder for QR code generation.
// This encoder uses the gozxing library's QRCodeWriter to generate QR codes.
type GozxingEncoder struct{}

// Name returns the encoder identifier.
func (e *GozxingEncoder) Name() string {
	return "gozxing/encoder"
}

// Encode generates a QR code image from the input data.
// The gozxing library generates a BitMatrix which is converted to image.Image.
func (e *GozxingEncoder) Encode(data []byte, opts EncodeOptions) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("gozxing: cannot encode empty data")
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
		return nil, fmt.Errorf("gozxing: invalid error correction level %q", opts.ErrorCorrectionLevel)
	}

	// Create encoding hints
	hints := make(map[gozxing.EncodeHintType]interface{})
	hints[gozxing.EncodeHintType_ERROR_CORRECTION] = levelString

	// Encode to BitMatrix
	writer := qrcode.NewQRCodeWriter()
	bitMatrix, err := writer.Encode(string(data), gozxing.BarcodeFormat_QR_CODE,
		opts.PixelSize, opts.PixelSize, hints)
	if err != nil {
		return nil, fmt.Errorf("gozxing: encode failed: %w", err)
	}

	// Convert BitMatrix to image.Image
	img := bitMatrixToImage(bitMatrix)
	return img, nil
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
