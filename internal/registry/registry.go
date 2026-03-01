package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrVaultNotFound is returned when a name is absent from the registry.
	ErrVaultNotFound = errors.New("vault not found in registry")
	// ErrVaultAlreadyExists is returned when Add() is called with a duplicate name.
	ErrVaultAlreadyExists = errors.New("vault name already registered")
)

// Registry is the in-memory representation of ~/.config/vb/vaults.json.
type Registry struct {
	Default string            `json:"default"`
	Vaults  map[string]string `json:"vaults"`
}

// RegistryPath returns the absolute path to vaults.json,
// respecting XDG_CONFIG_HOME if set.
func RegistryPath() string {
	return filepath.Join(configDir(), "vaults.json")
}

// configDir returns the vb config directory, following the XDG Base Directory spec.
func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "vb")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "vb")
}

// Load reads vaults.json and returns the Registry.
// If the file does not exist, an empty Registry is returned (not an error).
func Load() (Registry, error) {
	data, err := os.ReadFile(RegistryPath())
	if err != nil {
		if os.IsNotExist(err) {
			return Registry{Vaults: map[string]string{}}, nil
		}
		return Registry{}, fmt.Errorf("reading registry: %w", err)
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return Registry{}, fmt.Errorf("parsing registry: %w", err)
	}
	if reg.Vaults == nil {
		reg.Vaults = map[string]string{}
	}
	return reg, nil
}

// Save writes the Registry to vaults.json, creating the config directory if needed.
func (r Registry) Save() error {
	dir := configDir()
	if dir == "" {
		return fmt.Errorf("cannot determine config directory")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising registry: %w", err)
	}
	return os.WriteFile(RegistryPath(), data, 0644)
}

// Add registers name → absPath. Returns ErrVaultAlreadyExists if name is taken.
func (r *Registry) Add(name, absPath string) error {
	if _, exists := r.Vaults[name]; exists {
		return fmt.Errorf("%w: %q", ErrVaultAlreadyExists, name)
	}
	r.Vaults[name] = absPath
	return nil
}

// Remove unregisters name. Returns ErrVaultNotFound if name is absent.
// Clears Default if the removed vault was the default.
func (r *Registry) Remove(name string) error {
	if _, exists := r.Vaults[name]; !exists {
		return fmt.Errorf("%w: %q", ErrVaultNotFound, name)
	}
	delete(r.Vaults, name)
	if r.Default == name {
		r.Default = ""
	}
	return nil
}

// SetDefault marks name as the default vault.
// Returns ErrVaultNotFound if name is not registered.
func (r *Registry) SetDefault(name string) error {
	if _, exists := r.Vaults[name]; !exists {
		return fmt.Errorf("%w: %q", ErrVaultNotFound, name)
	}
	r.Default = name
	return nil
}

// Lookup returns the absolute path for the named vault.
// Returns ErrVaultNotFound if name is not registered.
func (r Registry) Lookup(name string) (string, error) {
	path, exists := r.Vaults[name]
	if !exists {
		return "", fmt.Errorf("%w: %q", ErrVaultNotFound, name)
	}
	return path, nil
}
