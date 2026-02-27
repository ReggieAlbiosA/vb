package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// resetQueryFlags resets lens flag state after each test to prevent bleed
// across sequential test runs on the shared rootCmd instance.
func resetQueryFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		for _, name := range []string{"why", "importance", "cli-tools", "arch", "used", "gotchas", "refs", "gui"} {
			if f := rootCmd.Flags().Lookup(name); f != nil {
				f.Changed = false
				f.Value.Set("false") //nolint:errcheck
			}
		}
	})
}

// setupVaultWithTopic initialises a vault, creates hardware/disk with WHY.md,
// reindexes, and returns the vault root.
func setupVaultWithTopic(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	if _, err := execCmd(t, dir, "init"); err != nil {
		t.Fatalf("vb init: %v", err)
	}

	topicDir := filepath.Join(dir, "hardware", "disk")
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(topicDir, "WHY.md"), []byte("# Why disk"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := execCmd(t, dir, "reindex"); err != nil {
		t.Fatalf("vb reindex: %v", err)
	}
	return dir
}

// TestQueryCmd_Success: vb disk --why with authored file → no error.
func TestQueryCmd_Success(t *testing.T) {
	resetQueryFlags(t)
	dir := setupVaultWithTopic(t)

	_, err := execCmd(t, dir, "disk", "--why")
	if err != nil {
		t.Fatalf("vb disk --why: %v", err)
	}
}

// TestQueryCmd_TopicNotFound: vb unknown --why → error.
func TestQueryCmd_TopicNotFound(t *testing.T) {
	resetQueryFlags(t)
	dir := setupVaultWithTopic(t)

	_, err := execCmd(t, dir, "unknown", "--why")
	if err == nil {
		t.Fatal("expected error for unknown topic, got nil")
	}
}

// TestQueryCmd_LensFileMissing: vb disk --why with no WHY.md → user-readable error.
func TestQueryCmd_LensFileMissing(t *testing.T) {
	resetQueryFlags(t)
	dir := t.TempDir()

	if _, err := execCmd(t, dir, "init"); err != nil {
		t.Fatalf("vb init: %v", err)
	}

	// Create topic dir with ARCH.md only — no WHY.md.
	topicDir := filepath.Join(dir, "hardware", "disk")
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(topicDir, "ARCH.md"), []byte("# Arch"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := execCmd(t, dir, "reindex"); err != nil {
		t.Fatalf("vb reindex: %v", err)
	}

	_, err := execCmd(t, dir, "disk", "--why")
	if err == nil {
		t.Fatal("expected error for missing WHY.md, got nil")
	}
}

// TestQueryCmd_NoLens: vb disk (no flag) → error.
func TestQueryCmd_NoLens(t *testing.T) {
	resetQueryFlags(t)
	dir := setupVaultWithTopic(t)

	_, err := execCmd(t, dir, "disk")
	if err == nil {
		t.Fatal("expected error for no lens flag, got nil")
	}
}

// TestQueryCmd_GUIModifier: vb disk --why --gui → resolves without error.
func TestQueryCmd_GUIModifier(t *testing.T) {
	resetQueryFlags(t)
	dir := setupVaultWithTopic(t)

	_, err := execCmd(t, dir, "disk", "--why", "--gui")
	if err != nil {
		t.Fatalf("vb disk --why --gui: %v", err)
	}
}

// TestQueryCmd_UsedFlag: vb disk --why --used → renders output and appends entry to USED.md.
func TestQueryCmd_UsedFlag(t *testing.T) {
	resetQueryFlags(t)
	dir := setupVaultWithTopic(t)

	_, err := execCmd(t, dir, "disk", "--why", "--used")
	if err != nil {
		t.Fatalf("vb disk --why --used: %v", err)
	}

	// USED.md must exist inside the topic directory.
	usedPath := filepath.Join(dir, "hardware", "disk", "USED.md")
	data, readErr := os.ReadFile(usedPath)
	if readErr != nil {
		t.Fatalf("USED.md not created: %v", readErr)
	}
	if len(data) == 0 {
		t.Error("USED.md is empty — expected at least one log entry")
	}
}
