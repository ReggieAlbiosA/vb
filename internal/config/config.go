package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the parsed .vb/config.toml values.
type Config struct {
	KnowledgePath string `mapstructure:"knowledge_path"`
	Editor        string `mapstructure:"editor"`
	Theme         string `mapstructure:"theme"`
	LintOnSave    bool   `mapstructure:"lint_on_save"`
}

// Default returns the config values written by vb init.
func Default() Config {
	return Config{
		KnowledgePath: ".",
		Editor:        "nano",
		Theme:         "dark",
		LintOnSave:    false,
	}
}

// DefaultTOML returns the raw TOML content written by vb init.
func DefaultTOML() string {
	return `knowledge_path = "."
editor         = "nano"
theme          = "dark"
lint_on_save   = false
`
}

// Load reads .vb/config.toml from vaultRoot and returns a Config.
func Load(vaultRoot string) (Config, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(filepath.Join(vaultRoot, ".vb", "config.toml"))

	// Defaults
	v.SetDefault("knowledge_path", ".")
	v.SetDefault("editor", "nano")
	v.SetDefault("theme", "dark")
	v.SetDefault("lint_on_save", false)

	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}
