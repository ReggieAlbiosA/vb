package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// resetLintFlags resets lintCmd lens flag state between tests.
func resetLintFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		for _, name := range []string{"why", "importance", "cli-tools", "arch", "gotchas", "refs"} {
			if f := lintCmd.Flags().Lookup(name); f != nil {
				f.Changed = false
				f.Value.Set("false") //nolint:errcheck
			}
		}
	})
}

// setupVaultWithLensFile creates a vault with disk/WHY.md containing content.
func setupVaultWithLensFile(t *testing.T, lensFile, content string) string {
	t.Helper()
	dir := t.TempDir()

	if _, err := execCmd(t, dir, "init"); err != nil {
		t.Fatalf("vb init: %v", err)
	}

	topicDir := filepath.Join(dir, "hardware", "disk")
	if err := os.MkdirAll(topicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(topicDir, lensFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := execCmd(t, dir, "reindex"); err != nil {
		t.Fatalf("vb reindex: %v", err)
	}
	return dir
}

// TestLintCmd_Valid: valid WHY.md → exits 0 with schema valid message.
func TestLintCmd_Valid(t *testing.T) {
	resetLintFlags(t)
	dir := setupVaultWithLensFile(t, "WHY.md", "# Why\n\nBecause it matters.")

	out, err := execCmd(t, dir, "lint", "disk", "--why")
	if err != nil {
		t.Fatalf("vb lint disk --why: %v", err)
	}
	_ = out
}

// TestLintCmd_Violation: WHY.md with only a heading (no paragraph) → exits non-zero.
func TestLintCmd_Violation(t *testing.T) {
	resetLintFlags(t)
	dir := setupVaultWithLensFile(t, "WHY.md", "# Heading only\n")

	_, err := execCmd(t, dir, "lint", "disk", "--why")
	if err == nil {
		t.Fatal("expected non-zero exit for schema violation, got nil")
	}
}

// TestLintCmd_TopicNotFound: unknown topic → resolver error.
func TestLintCmd_TopicNotFound(t *testing.T) {
	resetLintFlags(t)
	dir := setupVaultWithLensFile(t, "WHY.md", "# Why\n\nContent.")

	_, err := execCmd(t, dir, "lint", "unknown", "--why")
	if err == nil {
		t.Fatal("expected error for unknown topic, got nil")
	}
}

// TestLintCmd_LensFileMissing: known topic, lens file not authored → binder error.
func TestLintCmd_LensFileMissing(t *testing.T) {
	resetLintFlags(t)
	// Only WHY.md exists — REFS.md does not.
	dir := setupVaultWithLensFile(t, "WHY.md", "# Why\n\nContent.")

	_, err := execCmd(t, dir, "lint", "disk", "--refs")
	if err == nil {
		t.Fatal("expected error for missing lens file, got nil")
	}
}

// TestLintCmd_NoLens: no lens flag → ErrNoLens.
func TestLintCmd_NoLens(t *testing.T) {
	resetLintFlags(t)
	dir := setupVaultWithLensFile(t, "WHY.md", "# Why\n\nContent.")

	_, err := execCmd(t, dir, "lint", "disk")
	if err == nil {
		t.Fatal("expected error for no lens flag, got nil")
	}
}
