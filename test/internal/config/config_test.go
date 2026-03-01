package config_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/internal/config"
)

func TestNew_ReturnsNonNilConfig(t *testing.T) {
	cfg := config.New()
	if cfg == nil {
		t.Fatal("config.New() returned nil")
	}
}

func TestNew_DefaultItemIconSize(t *testing.T) {
	cfg := config.New()
	if cfg.ItemIconSize <= 0 {
		t.Errorf("ItemIconSize = %v, want > 0", cfg.ItemIconSize)
	}
}

func TestNew_DefaultLegendIconSize(t *testing.T) {
	cfg := config.New()
	if cfg.Legend.IconSize <= 0 {
		t.Errorf("Legend.IconSize = %v, want > 0", cfg.Legend.IconSize)
	}
}

func TestNew_DefaultLegendFontSize(t *testing.T) {
	cfg := config.New()
	if cfg.Legend.FontSize <= 0 {
		t.Errorf("Legend.FontSize = %v, want > 0", cfg.Legend.FontSize)
	}
}

func TestNew_AssetDirIsSet(t *testing.T) {
	cfg := config.New()
	if cfg.AssetDir() == "" {
		t.Error("AssetDir() should not be empty")
	}
}

func TestNew_ServiceCatalogCSVPathIsSet(t *testing.T) {
	cfg := config.New()
	if cfg.ServiceCatalogCSVPath() == "" {
		t.Error("ServiceCatalogCSVPath() should not be empty")
	}
}

func TestNew_OutputFramesDirIsSet(t *testing.T) {
	cfg := config.New()
	if cfg.OutputFramesDir() == "" {
		t.Error("OutputFramesDir() should not be empty")
	}
}

func TestNew_ProjectRootIsSet(t *testing.T) {
	cfg := config.New()
	if cfg.ProjectRoot == "" {
		t.Error("ProjectRoot should not be empty")
	}
}
