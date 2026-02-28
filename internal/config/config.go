package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// appYAML mirrors the structure of etc/app.yaml.
type appYAML struct {
	Paths struct {
		AssetPackage      string `yaml:"asset_package"`
		ServiceCatalogCSV string `yaml:"service_catalog_csv"`
		OutputFrames      string `yaml:"output_frames"`
	} `yaml:"paths"`
	Legend struct {
		OffsetX  float64 `yaml:"offset_x"`
		OffsetY  float64 `yaml:"offset_y"`
		IconSize int     `yaml:"icon_size"`
		FontSize int     `yaml:"font_size"`
	} `yaml:"legend"`
}

// LegendConfig holds resolved legend defaults.
type LegendConfig struct {
	OffsetX  float64
	OffsetY  float64
	IconSize int
	FontSize int
}

// Config holds application-wide configuration resolved from etc/app.yaml.
type Config struct {
	ProjectRoot   string
	AssetDir_     string // absolute path to Asset-Package
	OutFramesDir  string // absolute path to generated frames output dir
	SvcCatalogCSV string // absolute path to service-catalog.csv
	Legend        LegendConfig
}

// New loads etc/app.yaml from the project root and returns a resolved Config.
// All paths fall back to sensible defaults when the config file is absent.
func New() *Config {
	root := findProjectRoot()

	def := appYAML{}
	def.Paths.AssetPackage = "etc/resources/aws/svg"
	def.Paths.ServiceCatalogCSV = "etc/resources/aws/service-catalog.csv"
	def.Paths.OutputFrames = "output/aws-frames"
	def.Legend.OffsetX = 120
	def.Legend.OffsetY = 0
	def.Legend.IconSize = 32
	def.Legend.FontSize = 12

	yamlPath := filepath.Join(root, "etc", "resources", "aws", "app.yaml")
	if data, err := os.ReadFile(yamlPath); err == nil {
		_ = yaml.Unmarshal(data, &def)
	}

	abs := func(rel string) string {
		if filepath.IsAbs(rel) {
			return rel
		}
		return filepath.Join(root, rel)
	}

	return &Config{
		ProjectRoot:   root,
		AssetDir_:     abs(def.Paths.AssetPackage),
		OutFramesDir:  abs(def.Paths.OutputFrames),
		SvcCatalogCSV: abs(def.Paths.ServiceCatalogCSV),
		Legend: LegendConfig{
			OffsetX:  def.Legend.OffsetX,
			OffsetY:  def.Legend.OffsetY,
			IconSize: def.Legend.IconSize,
			FontSize: def.Legend.FontSize,
		},
	}
}

// AssetDir returns the absolute path to the Asset-Package directory.
func (c *Config) AssetDir() string { return c.AssetDir_ }

// OutputFramesDir returns the absolute path to the frames output directory.
func (c *Config) OutputFramesDir() string { return c.OutFramesDir }

// ServiceCatalogCSVPath returns the absolute path to service-catalog.csv.
func (c *Config) ServiceCatalogCSVPath() string { return c.SvcCatalogCSV }

// findProjectRoot walks up from cwd until it finds go.mod, then returns that dir.
func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}
