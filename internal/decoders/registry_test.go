package decoders

import (
	"testing"

	"github.com/13rac1/qr-library-test/internal/config"
)

func TestGetAvailableDecoders_DefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	decoders := GetAvailableDecoders(cfg)

	// Default config should include all decoders (gozxing, tuotoo, goqr)
	expectedCount := 3
	if len(decoders) != expectedCount {
		t.Errorf("GetAvailableDecoders() returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify we have the expected decoder names
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"gozxing", "tuotoo", "goqr"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAvailableDecoders() missing decoder %q", name)
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
		if dec.Name() == "goqr" {
			t.Error("GetAvailableDecoders() with SkipArchived should not include goqr")
		}
	}

	// Verify we still have gozxing and tuotoo
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"gozxing", "tuotoo"}
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

	// CGO decoders not implemented yet (commit 8)
	// Should still have all 3 pure Go decoders
	expectedCount := 3
	if len(decoders) != expectedCount {
		t.Errorf("GetAvailableDecoders() with SkipCGO returned %d decoders, want %d", len(decoders), expectedCount)
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

	expected := []string{"gozxing", "tuotoo"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAvailableDecoders() missing decoder %q", name)
		}
	}
}

func TestGetAllDecoders(t *testing.T) {
	decoders := GetAllDecoders()

	// Should return all 3 decoders regardless of config
	expectedCount := 3
	if len(decoders) != expectedCount {
		t.Errorf("GetAllDecoders() returned %d decoders, want %d", len(decoders), expectedCount)
	}

	// Verify we have all expected decoders
	names := make(map[string]bool)
	for _, dec := range decoders {
		names[dec.Name()] = true
	}

	expected := []string{"gozxing", "tuotoo", "goqr"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("GetAllDecoders() missing decoder %q", name)
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

	core := []string{"gozxing", "tuotoo"}
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
