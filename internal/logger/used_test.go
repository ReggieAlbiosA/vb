package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppend_CreatesFile(t *testing.T) {
	dir := t.TempDir()

	if err := Append(dir, "disk", "why"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "USED.md")); os.IsNotExist(err) {
		t.Fatal("USED.md was not created")
	}
}

func TestAppend_AppendsNotOverwrites(t *testing.T) {
	dir := t.TempDir()

	if err := Append(dir, "disk", "why"); err != nil {
		t.Fatal(err)
	}
	if err := Append(dir, "disk", "arch"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "USED.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "--why") {
		t.Error("first entry (--why) was overwritten")
	}
	if !strings.Contains(content, "--arch") {
		t.Error("second entry (--arch) is missing")
	}
}

func TestAppend_EntryFormat(t *testing.T) {
	dir := t.TempDir()

	if err := Append(dir, "disk", "why"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "USED.md"))
	if err != nil {
		t.Fatal(err)
	}

	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "- ") {
		t.Errorf("entry does not start with '- ': %q", line)
	}
	if !strings.Contains(line, "vb disk --why") {
		t.Errorf("entry missing 'vb disk --why': %q", line)
	}
	if !strings.Contains(line, "UTC") {
		t.Errorf("entry missing 'UTC' timestamp: %q", line)
	}
}

func TestFormatEntry_Timestamp(t *testing.T) {
	ts, _ := time.Parse("2006-01-02 15:04 UTC", "2026-02-27 14:01 UTC")
	e := Entry{Timestamp: ts, Topic: "disk", Lens: "why"}
	got := formatEntry(e)
	want := "- 2026-02-27 14:01 UTC  vb disk --why"
	if got != want {
		t.Errorf("formatEntry() = %q, want %q", got, want)
	}
}

func TestAppend_WriteError(t *testing.T) {
	dir := t.TempDir()

	// Create a directory named USED.md so os.OpenFile fails.
	if err := os.Mkdir(filepath.Join(dir, "USED.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := Append(dir, "disk", "why")
	if err == nil {
		t.Fatal("expected error when USED.md is a directory, got nil")
	}
}
