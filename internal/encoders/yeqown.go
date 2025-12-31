// Package encoders provides QR code encoder implementations.
package encoders

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"

	qrc "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// YeqownEncoder wraps github.com/yeqown/go-qrcode/v2 for QR code generation.
// This encoder uses a builder/writer pattern to generate QR codes.
type YeqownEncoder struct{}

// Name returns the encoder identifier.
func (e *YeqownEncoder) Name() string {
	return "yeqown/go-qrcode"
}

// bufferCloser wraps bytes.Buffer to implement io.WriteCloser.
type bufferCloser struct {
	*bytes.Buffer
}

// Close implements io.Closer interface (no-op for buffer).
func (bc *bufferCloser) Close() error {
	return nil
}

// Encode generates a QR code image from the input data.
// The yeqown/go-qrcode library uses a writer pattern to generate images.
func (e *YeqownEncoder) Encode(data []byte, opts EncodeOptions) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("yeqown: cannot encode empty data")
	}

	// Map error correction level to qrc package constants
	// Note: We use a variable to hold the EncodeOption since ecLevel type is unexported
	var levelOption qrc.EncodeOption
	switch opts.ErrorCorrectionLevel {
	case ErrorCorrectionL:
		levelOption = qrc.WithErrorCorrectionLevel(qrc.ErrorCorrectionLow)
	case ErrorCorrectionM:
		levelOption = qrc.WithErrorCorrectionLevel(qrc.ErrorCorrectionMedium)
	case ErrorCorrectionQ:
		levelOption = qrc.WithErrorCorrectionLevel(qrc.ErrorCorrectionQuart)
	case ErrorCorrectionH:
		levelOption = qrc.WithErrorCorrectionLevel(qrc.ErrorCorrectionHighest)
	default:
		return nil, fmt.Errorf("yeqown: invalid error correction level %q", opts.ErrorCorrectionLevel)
	}

	// Create QR code with options
	qrCode, err := qrc.NewWith(string(data), levelOption)
	if err != nil {
		return nil, fmt.Errorf("yeqown: QR code creation failed: %w", err)
	}

	// Write to buffer using standard writer
	buf := &bufferCloser{Buffer: new(bytes.Buffer)}
	writer := standard.NewWithWriter(buf,
		standard.WithQRWidth(uint8(opts.PixelSize/qrCode.Dimension())),
		standard.WithBgTransparent(),
	)

	if err := qrCode.Save(writer); err != nil {
		return nil, fmt.Errorf("yeqown: save failed: %w", err)
	}

	// Decode PNG bytes to image.Image
	img, _, err := image.Decode(buf.Buffer)
	if err != nil {
		return nil, fmt.Errorf("yeqown: PNG decode failed: %w", err)
	}

	return img, nil
}
