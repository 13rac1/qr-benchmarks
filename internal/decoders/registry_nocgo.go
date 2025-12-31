//go:build !cgo
// +build !cgo

package decoders

// cgoEnabled returns false when CGO is not available.
// This file is compiled when CGO is disabled (default on most systems).
func cgoEnabled() bool {
	return false
}
