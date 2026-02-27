package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// resetEditFlags resets editCmd lens flag state between tests.
func resetEditFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		for _, name := range []string{"why", "importance", "cli-tools", "arch", "gotchas", "refs"} {
			if f := editCmd.Flags().Lookup(name); f != nil {
				f.Changed = false
				f.Value.Set("false") //nolint:errcheck
			}
		}
	})
}

// setVaultEditor overwrites .vb/config.toml with the given editor binary so
// cfg.Editor is set directly and resolveEditor never falls through to $EDITOR.
func setVaultEditor(t *testing.T, vaultRoot, editorBin string) {
	t.Helper()
	content := fmt.Sprintf("knowledge_path = \".\"\neditor = %q\ntheme = \"dark\"\nlint_on_save = false\n", editorBin)
	if err := os.WriteFile(filepath.Join(vaultRoot, ".vb", "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestEditCmd_Success: vb edit disk --why with existing topic + WHY.md → no error.
func TestEditCmd_Success(t *testing.T) {
	resetEditFlags(t)
	dir := setupVaultWithTopic(t)
	setVaultEditor(t, dir, "true") // "true" exits 0 immediately

	_, err := execCmd(t, dir, "edit", "disk", "--why")
	if err != nil {
		t.Fatalf("vb edit disk --why: %v", err)
	}
}

// TestEditCmd_TopicNotFound: unknown topic → ErrTopicNotFound.
func TestEditCmd_TopicNotFound(t *testing.T) {
	resetEditFlags(t)
	dir := setupVaultWithTopic(t)
	setVaultEditor(t, dir, "true")

	_, err := execCmd(t, dir, "edit", "unknown", "--why")
	if err == nil {
		t.Fatal("expected error for unknown topic, got nil")
	}
}

// TestEditCmd_NoLens: no lens flag → ErrNoLens.
func TestEditCmd_NoLens(t *testing.T) {
	resetEditFlags(t)
	dir := setupVaultWithTopic(t)
	setVaultEditor(t, dir, "true")

	_, err := execCmd(t, dir, "edit", "disk")
	if err == nil {
		t.Fatal("expected error for no lens flag, got nil")
	}
}

// TestEditCmd_CreatesLensFile: vb edit disk --arch with no ARCH.md → creates file, opens editor.
func TestEditCmd_CreatesLensFile(t *testing.T) {
	resetEditFlags(t)
	dir := setupVaultWithTopic(t) // only WHY.md exists
	setVaultEditor(t, dir, "true")

	// ARCH.md does not exist yet.
	archPath := filepath.Join(dir, "hardware", "disk", "ARCH.md")
	if _, err := os.Stat(archPath); !os.IsNotExist(err) {
		t.Skip("ARCH.md already exists, skipping creation test")
	}

	_, err := execCmd(t, dir, "edit", "disk", "--arch")
	if err != nil {
		t.Fatalf("vb edit disk --arch: %v", err)
	}

	if _, err := os.Stat(archPath); os.IsNotExist(err) {
		t.Error("ARCH.md was not created by editor.Open")
	}
}
