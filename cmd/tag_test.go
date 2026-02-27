package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupVaultWithTag creates a vault with a flat disk/ topic whose WHY.md contains the given tag.
// Topic is placed directly under TopicRoot (no category subdirectory) so the tagger
// returns topic = "disk" as the first path component.
func setupVaultWithTag(t *testing.T, tag string) string {
	t.Helper()
	dir := t.TempDir()

	if _, err := execCmd(t, dir, "init"); err != nil {
		t.Fatalf("vb init: %v", err)
	}

	topicDir := filepath.Join(dir, "disk")
	if err := os.MkdirAll(topicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# Why disk\n\nUsed with #" + tag + " for storage."
	if err := os.WriteFile(filepath.Join(topicDir, "WHY.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Reindex so the vault is valid, even though tagger doesn't use the index.
	if _, err := execCmd(t, dir, "reindex"); err != nil {
		t.Fatalf("vb reindex: %v", err)
	}
	return dir
}

// TestTagCmd_Found: vb tag ssh with matching topic → result printed.
func TestTagCmd_Found(t *testing.T) {
	dir := setupVaultWithTag(t, "ssh")

	out, err := execCmd(t, dir, "tag", "ssh")
	if err != nil {
		t.Fatalf("vb tag ssh: %v", err)
	}
	if !strings.Contains(out, "disk") {
		t.Errorf("expected 'disk' in output, got: %q", out)
	}
}

// TestTagCmd_NotFound: unknown tag → "no topics tagged" message, exit 0.
func TestTagCmd_NotFound(t *testing.T) {
	dir := setupVaultWithTag(t, "ssh")

	out, err := execCmd(t, dir, "tag", "unknowntag99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "no topics tagged") {
		t.Errorf("expected 'no topics tagged' in output, got: %q", out)
	}
}

// TestTagCmd_NoArg: no argument → cobra argument validation error.
func TestTagCmd_NoArg(t *testing.T) {
	dir := setupVaultWithTag(t, "ssh")

	_, err := execCmd(t, dir, "tag")
	if err == nil {
		t.Fatal("expected error for missing tag argument, got nil")
	}
}
