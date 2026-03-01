package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()

	if err := Save(dir, "lsblk", "show all block devices"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "USED.md")); os.IsNotExist(err) {
		t.Fatal("USED.md was not created")
	}
}

func TestSave_AppendsNotOverwrites(t *testing.T) {
	dir := t.TempDir()

	if err := Save(dir, "lsblk", "show all block devices"); err != nil {
		t.Fatal(err)
	}
	if err := Save(dir, "df -h", "check disk usage"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "USED.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "lsblk") {
		t.Error("first entry (lsblk) was overwritten")
	}
	if !strings.Contains(content, "df -h") {
		t.Error("second entry (df -h) is missing")
	}
}

func TestSave_Format(t *testing.T) {
	dir := t.TempDir()

	if err := Save(dir, "sudo parted -l", "list all partition tables"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "USED.md"))
	if err != nil {
		t.Fatal(err)
	}

	line := strings.TrimSpace(string(data))
	want := "- sudo parted -l — list all partition tables"
	if line != want {
		t.Errorf("entry = %q, want %q", line, want)
	}
}

func TestSave_WriteError(t *testing.T) {
	dir := t.TempDir()

	// Create a directory named USED.md so os.OpenFile fails.
	if err := os.Mkdir(filepath.Join(dir, "USED.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := Save(dir, "lsblk", "show block devices")
	if err == nil {
		t.Fatal("expected error when USED.md is a directory, got nil")
	}
}
