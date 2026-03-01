package registry

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func isolateRegistry(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func TestRegistryPath_XDG(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	want := filepath.Join(xdg, "vb", "vaults.json")
	got := RegistryPath()
	if got != want {
		t.Errorf("RegistryPath() = %q, want %q", got, want)
	}
}

func TestRegistryPath_HomeDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	got := RegistryPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "vb", "vaults.json")
	if got != want {
		t.Errorf("RegistryPath() = %q, want %q", got, want)
	}
}

func TestLoad_NoFile(t *testing.T) {
	isolateRegistry(t)

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg.Default != "" {
		t.Errorf("Default = %q, want empty", reg.Default)
	}
	if len(reg.Vaults) != 0 {
		t.Errorf("Vaults length = %d, want 0", len(reg.Vaults))
	}
}

func TestSave_RoundTrip(t *testing.T) {
	isolateRegistry(t)

	reg := Registry{
		Default: "alpha",
		Vaults: map[string]string{
			"alpha": "/path/to/alpha",
			"beta":  "/path/to/beta",
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Default != "alpha" {
		t.Errorf("Default = %q, want %q", loaded.Default, "alpha")
	}
	if len(loaded.Vaults) != 2 {
		t.Fatalf("Vaults length = %d, want 2", len(loaded.Vaults))
	}
	if loaded.Vaults["alpha"] != "/path/to/alpha" {
		t.Errorf("alpha path = %q", loaded.Vaults["alpha"])
	}
}

func TestAdd_Success(t *testing.T) {
	reg := Registry{Vaults: map[string]string{}}
	if err := reg.Add("test", "/tmp/test"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if reg.Vaults["test"] != "/tmp/test" {
		t.Errorf("Vaults[test] = %q", reg.Vaults["test"])
	}
}

func TestAdd_Duplicate(t *testing.T) {
	reg := Registry{Vaults: map[string]string{"test": "/tmp/test"}}
	err := reg.Add("test", "/tmp/other")
	if err == nil {
		t.Fatal("Add() duplicate: expected error, got nil")
	}
	if !errors.Is(err, ErrVaultAlreadyExists) {
		t.Errorf("expected ErrVaultAlreadyExists, got: %v", err)
	}
}

func TestRemove_Success(t *testing.T) {
	reg := Registry{Vaults: map[string]string{"test": "/tmp/test"}}
	if err := reg.Remove("test"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}
	if _, exists := reg.Vaults["test"]; exists {
		t.Error("vault still exists after Remove()")
	}
}

func TestRemove_ClearsDefault(t *testing.T) {
	reg := Registry{
		Default: "test",
		Vaults:  map[string]string{"test": "/tmp/test"},
	}
	if err := reg.Remove("test"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}
	if reg.Default != "" {
		t.Errorf("Default = %q after removing default vault, want empty", reg.Default)
	}
}

func TestRemove_NotFound(t *testing.T) {
	reg := Registry{Vaults: map[string]string{}}
	err := reg.Remove("missing")
	if err == nil {
		t.Fatal("Remove() missing: expected error, got nil")
	}
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("expected ErrVaultNotFound, got: %v", err)
	}
}

func TestSetDefault_Success(t *testing.T) {
	reg := Registry{Vaults: map[string]string{"test": "/tmp/test"}}
	if err := reg.SetDefault("test"); err != nil {
		t.Fatalf("SetDefault() error: %v", err)
	}
	if reg.Default != "test" {
		t.Errorf("Default = %q, want %q", reg.Default, "test")
	}
}

func TestSetDefault_NotFound(t *testing.T) {
	reg := Registry{Vaults: map[string]string{}}
	err := reg.SetDefault("missing")
	if err == nil {
		t.Fatal("SetDefault() missing: expected error, got nil")
	}
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("expected ErrVaultNotFound, got: %v", err)
	}
}

func TestLookup_Success(t *testing.T) {
	reg := Registry{Vaults: map[string]string{"test": "/tmp/test"}}
	path, err := reg.Lookup("test")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if path != "/tmp/test" {
		t.Errorf("Lookup() = %q, want %q", path, "/tmp/test")
	}
}

func TestLookup_NotFound(t *testing.T) {
	reg := Registry{Vaults: map[string]string{}}
	_, err := reg.Lookup("missing")
	if err == nil {
		t.Fatal("Lookup() missing: expected error, got nil")
	}
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("expected ErrVaultNotFound, got: %v", err)
	}
}

func TestSave_CreatesDir(t *testing.T) {
	isolateRegistry(t)

	reg := Registry{Vaults: map[string]string{}}
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(RegistryPath()); os.IsNotExist(err) {
		t.Error("vaults.json not created")
	}
}
