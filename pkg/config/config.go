// Package config handles loading and validating DiffScope configuration.
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Config represents the full configuration for DiffScope.
type Config struct {
	Defaults Defaults `yaml:"defaults" toml:"defaults"`
}

// Defaults holds the default settings for analysis.
type Defaults struct {
	MaxDiffLines     int      `yaml:"max_diff_lines" toml:"max_diff_lines" default:"10000"`
	Timeout          int      `yaml:"timeout" toml:"timeout" default:"30"`
	Languages        []string `yaml:"languages" toml:"languages" default:"[go,python,typescript]"`
	ScoringThreshold int      `yaml:"scoring_threshold" toml:"scoring_threshold" default:"50"`
	ScanDependents   bool     `yaml:"scan_dependents" toml:"scan_dependents" default:"true"`
	ExcludePatterns  []string `yaml:"exclude_patterns" toml:"exclude_patterns"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Defaults: Defaults{
			MaxDiffLines:     10000,
			Timeout:          30,
			Languages:        []string{"go", "python", "typescript"},
			ScoringThreshold: 50,
			ScanDependents:   true,
			ExcludePatterns:  []string{},
		},
	}
}

// LoadFromFile loads configuration from a TOML file.
func LoadFromFile(path string) (Config, error) {
	var c Config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return Config{}, fmt.Errorf("load config from %s: %w", path, err)
	}
	return c, nil
}

// LoadYAMLFromBytes loads configuration from YAML bytes.
func LoadYAMLFromBytes(data []byte) (Config, error) {
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse YAML config: %w", err)
	}
	return c, nil
}

// SaveTOML writes the config to a TOML file.
func (c *Config) SaveTOML(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	return nil
}

// Validate checks that the configuration values are reasonable.
func (c *Config) Validate() error {
	if c.Defaults.MaxDiffLines < 1 {
		c.Defaults.MaxDiffLines = DefaultConfig().Defaults.MaxDiffLines
	}
	if c.Defaults.MaxDiffLines > 1000000 {
		return fmt.Errorf("max_diff_lines too large (max 1,000,000)")
	}
	if c.Defaults.Timeout < 1 {
		c.Defaults.Timeout = DefaultConfig().Defaults.Timeout
	}
	return nil
}

// Merge applies settings from another Config into this one (non-zero overrides).
func (c *Config) Merge(other Config) {
	if other.Defaults.MaxDiffLines > 0 {
		c.Defaults.MaxDiffLines = other.Defaults.MaxDiffLines
	}
	if other.Defaults.Timeout > 0 {
		c.Defaults.Timeout = other.Defaults.Timeout
	}
	if len(other.Defaults.Languages) > 0 {
		c.Defaults.Languages = other.Defaults.Languages
	}
	if other.Defaults.ScoringThreshold > 0 {
		c.Defaults.ScoringThreshold = other.Defaults.ScoringThreshold
	}
	if other.Defaults.ScanDependents {
		c.Defaults.ScanDependents = true
	}
	if len(other.Defaults.ExcludePatterns) > 0 {
		c.Defaults.ExcludePatterns = append(c.Defaults.ExcludePatterns, other.Defaults.ExcludePatterns...)
	}
}
