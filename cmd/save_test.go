package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resetSaveFlags resets the save command's flag state after each test.
func resetSaveFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		flagSaveDesc = ""
	})
}

// TestSaveCmd_Success: saves a command to USED.md.
func TestSaveCmd_Success(t *testing.T) {
	isolateRegistry(t)
	resetSaveFlags(t)
	dir := setupVaultForTopic(t)

	// Create a topic first.
	if _, err := execCmd(t, dir, "topic", "create", "disk"); err != nil {
		t.Fatalf("create topic: %v", err)
	}

	out, err := execCmd(t, dir, "save", "disk", "lsblk", "-d", "show all block devices")
	if err != nil {
		t.Fatalf("vb save error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, "saved") {
		t.Errorf("output missing 'saved': %s", out)
	}

	// Verify USED.md has the entry.
	data, readErr := os.ReadFile(filepath.Join(dir, "disk", "USED.md"))
	if readErr != nil {
		t.Fatalf("reading USED.md: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "lsblk") {
		t.Errorf("USED.md missing command 'lsblk': %s", content)
	}
	if !strings.Contains(content, "show all block devices") {
		t.Errorf("USED.md missing description: %s", content)
	}
}

// TestSaveCmd_TopicNotFound: errors for unknown topic.
func TestSaveCmd_TopicNotFound(t *testing.T) {
	isolateRegistry(t)
	resetSaveFlags(t)
	dir := setupVaultForTopic(t)

	_, err := execCmd(t, dir, "save", "nonexistent", "ls", "-d", "list files")
	if err == nil {
		t.Fatal("expected error for unknown topic, got nil")
	}
}

// TestSaveCmd_MissingDesc: errors when -d flag is missing.
func TestSaveCmd_MissingDesc(t *testing.T) {
	isolateRegistry(t)
	resetSaveFlags(t)
	dir := setupVaultForTopic(t)

	_, err := execCmd(t, dir, "save", "disk", "lsblk")
	if err == nil {
		t.Fatal("expected error for missing -d flag, got nil")
	}
}

// TestSaveCmd_MultipleEntries: multiple saves append to USED.md.
func TestSaveCmd_MultipleEntries(t *testing.T) {
	isolateRegistry(t)
	resetSaveFlags(t)
	dir := setupVaultForTopic(t)

	if _, err := execCmd(t, dir, "topic", "create", "disk"); err != nil {
		t.Fatalf("create topic: %v", err)
	}

	if _, err := execCmd(t, dir, "save", "disk", "lsblk", "-d", "block devices"); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if _, err := execCmd(t, dir, "save", "disk", "df -h", "-d", "disk usage"); err != nil {
		t.Fatalf("second save: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "disk", "USED.md"))
	content := string(data)
	if !strings.Contains(content, "lsblk") {
		t.Error("first entry missing")
	}
	if !strings.Contains(content, "df -h") {
		t.Error("second entry missing")
	}
}
