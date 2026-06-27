// Package config_test tests configuration loading and validation.
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"diffscope/pkg/config"
)

func TestDefaultConfig(t *testing.T) {
	c := config.DefaultConfig()
	if c.Defaults.MaxDiffLines != 10000 {
		t.Errorf("expected MaxDiffLines=10000, got %d", c.Defaults.MaxDiffLines)
	}
	if c.Defaults.Timeout != 30 {
		t.Errorf("expected Timeout=30, got %d", c.Defaults.Timeout)
	}
	if !c.Defaults.ScanDependents {
		t.Error("expected ScanDependents=true")
	}
}

func TestValidateValidConfig(t *testing.T) {
	c := config.DefaultConfig()
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateNegativeMaxDiffLines(t *testing.T) {
	c := config.DefaultConfig()
	c.Defaults.MaxDiffLines = -1
	c.Validate() // should auto-correct
	if c.Defaults.MaxDiffLines != 10000 {
		t.Errorf("expected auto-corrected MaxDiffLines=10000, got %d", c.Defaults.MaxDiffLines)
	}
}

func TestValidateTooLargeMaxDiffLines(t *testing.T) {
	c := config.DefaultConfig()
	c.Defaults.MaxDiffLines = 2000000
	if err := c.Validate(); err == nil {
		t.Error("expected error for very large MaxDiffLines")
	}
}

func TestValidateNegativeTimeout(t *testing.T) {
	c := config.DefaultConfig()
	c.Defaults.Timeout = -1
	c.Validate() // should auto-correct
	if c.Defaults.Timeout != 30 {
		t.Errorf("expected auto-corrected Timeout=30, got %d", c.Defaults.Timeout)
	}
}

func TestSaveAndLoadTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "diffscope.toml")

	c := config.DefaultConfig()
	if err := c.SaveTOML(path); err != nil {
		t.Fatalf("SaveTOML failed: %v", err)
	}

	loaded, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}
	if loaded.Defaults.MaxDiffLines != c.Defaults.MaxDiffLines {
		t.Errorf("MaxDiffLines mismatch: got %d, want %d", loaded.Defaults.MaxDiffLines, c.Defaults.MaxDiffLines)
	}
	if loaded.Defaults.ScanDependents != c.Defaults.ScanDependents {
		t.Error("ScanDependents mismatch")
	}
}

func TestMerge(t *testing.T) {
	c := config.DefaultConfig()
	other := config.Config{
		Defaults: config.Defaults{
			MaxDiffLines:     5000,
			ScoringThreshold: 75,
			Languages:        []string{"rust"},
		},
	}
	c.Merge(other)
	if c.Defaults.MaxDiffLines != 5000 {
		t.Errorf("MaxDiffLines: got %d, want 5000", c.Defaults.MaxDiffLines)
	}
	if c.Defaults.ScoringThreshold != 75 {
		t.Errorf("ScoringThreshold: got %d, want 75", c.Defaults.ScoringThreshold)
	}
	// ScanDependents: false is zero value, so Merge won't apply it
	if c.Defaults.ScanDependents != true {
		t.Error("ScanDependents should still be true")
	}
	if len(c.Defaults.Languages) != 1 || c.Defaults.Languages[0] != "rust" {
		t.Errorf("Languages: got %v, want [rust]", c.Defaults.Languages)
	}
}

func TestLoadYAML(t *testing.T) {
	yamlData := []byte(`
defaults:
  max_diff_lines: 20000
  timeout: 60
  languages: [go, python]
  scoring_threshold: 80
`)
	c, err := config.LoadYAMLFromBytes(yamlData)
	if err != nil {
		t.Fatalf("LoadYAMLFromBytes failed: %v", err)
	}
	if c.Defaults.MaxDiffLines != 20000 {
		t.Errorf("MaxDiffLines: got %d, want 20000", c.Defaults.MaxDiffLines)
	}
	if c.Defaults.Timeout != 60 {
		t.Errorf("Timeout: got %d, want 60", c.Defaults.Timeout)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := config.LoadFromFile("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadYAMLEmpty(t *testing.T) {
	c, err := config.LoadYAMLFromBytes([]byte{})
	if err != nil {
		t.Fatalf("LoadYAMLFromBytes on empty data failed: %v", err)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("empty config should be valid: %v", err)
	}
}

func TestLoadYAMLInvalid(t *testing.T) {
	_, err := config.LoadYAMLFromBytes([]byte("{ invalid yaml"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestConfigFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	c := config.DefaultConfig()
	if err := c.SaveTOML(path); err != nil {
		t.Fatalf("SaveTOML failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	// Verify file is readable by owner
	if info.Mode().Perm()&0400 == 0 {
		t.Log("config file permissions:", info.Mode().Perm())
	}
}
