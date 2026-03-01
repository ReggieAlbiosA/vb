package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_Success(t *testing.T) {
	dir := t.TempDir()

	if err := Init(dir); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	for _, name := range []string{".vb", ".vb/config.toml", ".vb/index.json"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", path)
		}
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()

	if err := Init(dir); err != nil {
		t.Fatalf("first Init() error: %v", err)
	}

	err := Init(dir)
	if err == nil {
		t.Fatal("second Init(): expected error, got nil")
	}
}

func TestInit_CreatesParentDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "nested", "vault")

	if err := Init(dir); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".vb", "config.toml")); os.IsNotExist(err) {
		t.Error("expected .vb/config.toml to exist in nested path")
	}
}
