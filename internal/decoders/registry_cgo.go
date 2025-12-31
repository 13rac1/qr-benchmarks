//go:build cgo
// +build cgo

package decoders

// cgoEnabled returns true when CGO is available.
// This file is compiled when CGO is enabled (CGO_ENABLED=1).
func cgoEnabled() bool {
	return true
}
