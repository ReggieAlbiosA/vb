package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the parsed .vb/config.toml values.
type Config struct {
	KnowledgePath string            `mapstructure:"knowledge_path"`
	Editor        string            `mapstructure:"editor"`
	Theme         string            `mapstructure:"theme"`
	LintOnSave    bool              `mapstructure:"lint_on_save"`
	CustomLenses  map[string]string `mapstructure:"custom_lenses"`
}

// Default returns the config values written by vb init.
func Default() Config {
	return Config{
		KnowledgePath: ".",
		Editor:        "nano",
		Theme:         "dark",
		LintOnSave:    false,
		CustomLenses:  map[string]string{},
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
	if cfg.CustomLenses == nil {
		cfg.CustomLenses = map[string]string{}
	}
	return cfg, nil
}

// SaveCustomLens adds a custom lens to the vault config and writes it back.
func SaveCustomLens(vaultRoot, flag, filename string) error {
	configPath := filepath.Join(vaultRoot, ".vb", "config.toml")

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	raw := string(content)

	// Check if [custom_lenses] section already exists.
	if strings.Contains(raw, "[custom_lenses]") {
		// Append the new lens after the section header.
		entry := fmt.Sprintf("%s = %q\n", flag, filename)
		raw = strings.Replace(raw, "[custom_lenses]\n", "[custom_lenses]\n"+entry, 1)
	} else {
		// Add the section at the end.
		raw += fmt.Sprintf("\n[custom_lenses]\n%s = %q\n", flag, filename)
	}

	return os.WriteFile(configPath, []byte(raw), 0o644)
}
