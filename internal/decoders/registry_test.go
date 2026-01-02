package decoders

import (
	"testing"

	"github.com/13rac1/qr-library-test/internal/config"
)

func TestGetAvailableDecoders_DefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	decoders := GetAvailableDecoders(cfg)

	// Default config should include all decoders (gozxing, tuotoo, goqr)
	// Plus goquirc if CGO is enabled
	expectedCount := 3
	if cgoEnabled() {
		expectedCount = 4
	}
	if len(decoders) != expectedCount {
		t.Errorf("GetAvailableDecoders() returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify we have the expected decoder names
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"makiuchi-d/gozxing", "tuotoo/qrcode", "liyue201/goqr"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAvailableDecoders() missing decoder %q", name)
		}
	}

	// Verify goquirc is included if CGO is enabled
	if cgoEnabled() {
		if !names["kdar/goquirc"] {
			t.Error("GetAvailableDecoders() should include kdar/goquirc when CGO is enabled")
		}
	}
}

func TestGetAvailableDecoders_SkipArchived(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SkipArchived = true

	decoders := GetAvailableDecoders(cfg)

	// Should only have gozxing and tuotoo (no goqr)
	expectedCount := 2
	if len(decoders) != expectedCount {
		t.Errorf("GetAvailableDecoders() with SkipArchived returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify goqr is excluded
	for _, dec := range decoders {
		if dec.Name() == "liyue201/goqr" {
			t.Error("GetAvailableDecoders() with SkipArchived should not include liyue201/goqr")
		}
	}

	// Verify we still have gozxing and tuotoo
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"makiuchi-d/gozxing", "tuotoo/qrcode"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAvailableDecoders() missing decoder %q", name)
		}
	}
}

func TestGetAvailableDecoders_SkipCGO(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SkipCGO = true

	decoders := GetAvailableDecoders(cfg)

	// With SkipCGO, should only have pure Go decoders (gozxing, tuotoo, goqr)
	expectedCount := 3
	if len(decoders) != expectedCount {
		t.Errorf("GetAvailableDecoders() with SkipCGO returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify goquirc is excluded even if CGO is available
	for _, dec := range decoders {
		if dec.Name() == "kdar/goquirc" {
			t.Error("GetAvailableDecoders() with SkipCGO should not include kdar/goquirc")
		}
	}
}

func TestGetAvailableDecoders_SkipBoth(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SkipArchived = true
	cfg.SkipCGO = true

	decoders := GetAvailableDecoders(cfg)

	// Should only have gozxing and tuotoo
	expectedCount := 2
	if len(decoders) != expectedCount {
		t.Errorf("GetAvailableDecoders() with both skip flags returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify we have the expected decoders
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"makiuchi-d/gozxing", "tuotoo/qrcode"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAvailableDecoders() missing decoder %q", name)
		}
	}
}

func TestGetAllDecoders(t *testing.T) {
	decoders := GetAllDecoders()

	// Should return all 3 decoders regardless of config
	// Plus goquirc if CGO is enabled
	expectedCount := 3
	if cgoEnabled() {
		expectedCount = 4
	}
	if len(decoders) != expectedCount {
		t.Errorf("GetAllDecoders() returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify we have all expected decoders
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"makiuchi-d/gozxing", "tuotoo/qrcode", "liyue201/goqr"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAllDecoders() missing decoder %q", name)
		}
	}

	// Verify goquirc is included if CGO is enabled
	if cgoEnabled() {
		if !names["kdar/goquirc"] {
			t.Error("GetAllDecoders() should include kdar/goquirc when CGO is enabled")
		}
	} else {
		if names["kdar/goquirc"] {
			t.Error("GetAllDecoders() should not include kdar/goquirc when CGO is disabled")
		}
	}
}

func TestGetAvailableDecoders_AlwaysIncludesCoreDecoders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SkipArchived = true
	cfg.SkipCGO = true

	decoders := GetAvailableDecoders(cfg)

	// Even with all skip flags, should always have core decoders
	if len(decoders) < 2 {
		t.Errorf("GetAvailableDecoders() with all skip flags returned %d decoders, want at least 2", len(decoders))
	}

	// Verify core decoders are present
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	core := []string{"makiuchi-d/gozxing", "tuotoo/qrcode"}
	for _, name := range core {
		if !names[name] {
			t.Errorf("GetAvailableDecoders() should always include core decoder %q", name)
		}
	}
}

func TestDecoderRegistry_NoNilDecoders(t *testing.T) {
	cfg := config.DefaultConfig()

	decoders := GetAvailableDecoders(cfg)

	for _, dec := range decoders {
		if dec == nil {
			t.Error("GetAvailableDecoders() returned nil decoder")
		}
	}

	allDecoders := GetAllDecoders()

	for _, dec := range allDecoders {
		if dec == nil {
			t.Error("GetAllDecoders() returned nil decoder")
		}
	}
}

func TestCgoEnabled(t *testing.T) {
	// This test verifies that cgoEnabled() returns the correct value
	// based on build tags. The actual value depends on whether the
	// code is built with CGO_ENABLED=1 or CGO_ENABLED=0.
	result := cgoEnabled()

	// Log the result for visibility
	t.Logf("cgoEnabled() = %v", result)

	// When built with CGO_ENABLED=1 and -tags cgo, result should be true
	// When built with CGO_ENABLED=0 or without cgo tag, result should be false
	// This test just verifies the function is callable and returns a boolean
	if result {
		t.Log("CGO is enabled - goquirc decoder will be available")
	} else {
		t.Log("CGO is disabled - goquirc decoder will not be available")
	}
}
