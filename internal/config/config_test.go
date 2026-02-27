package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.KnowledgePath != "." {
		t.Errorf("KnowledgePath: got %q, want %q", cfg.KnowledgePath, ".")
	}
	if cfg.Editor != "nano" {
		t.Errorf("Editor: got %q, want %q", cfg.Editor, "nano")
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme: got %q, want %q", cfg.Theme, "dark")
	}
	if cfg.LintOnSave != false {
		t.Error("LintOnSave: got true, want false")
	}
}

func TestDefaultTOML_ContainsKeys(t *testing.T) {
	toml := config.DefaultTOML()
	keys := []string{"knowledge_path", "editor", "theme", "lint_on_save"}
	for _, key := range keys {
		if !containsStr(toml, key) {
			t.Errorf("DefaultTOML() missing key %q", key)
		}
	}
}

func TestLoad_ReadsFromFile(t *testing.T) {
	dir := t.TempDir()
	vbDir := filepath.Join(dir, ".vb")
	if err := os.MkdirAll(vbDir, 0755); err != nil {
		t.Fatal(err)
	}

	toml := `knowledge_path = "/custom/path"
editor = "vim"
theme = "light"
lint_on_save = true
`
	if err := os.WriteFile(filepath.Join(vbDir, "config.toml"), []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.KnowledgePath != "/custom/path" {
		t.Errorf("KnowledgePath: got %q, want %q", cfg.KnowledgePath, "/custom/path")
	}
	if cfg.Editor != "vim" {
		t.Errorf("Editor: got %q, want %q", cfg.Editor, "vim")
	}
	if cfg.Theme != "light" {
		t.Errorf("Theme: got %q, want %q", cfg.Theme, "light")
	}
	if cfg.LintOnSave != true {
		t.Error("LintOnSave: got false, want true")
	}
}

func TestLoad_MissingFile_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	// No .vb/config.toml exists — should silently return defaults.
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.KnowledgePath != "." {
		t.Errorf("KnowledgePath: got %q, want default %q", cfg.KnowledgePath, ".")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStrHelper(s, sub))
}

func containsStrHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
