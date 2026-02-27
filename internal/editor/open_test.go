package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	if err := os.WriteFile(path, []byte("# Why"), 0o644); err != nil {
		t.Fatal(err)
	}

	// "true" exits 0 immediately without modifying the file.
	if err := Open(path, "true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("existing file was deleted by editor")
	}
}

func TestOpen_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ARCH.md")
	// Do not pre-create the file.

	if err := Open(path, "true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("missing file was not created before editor launch")
	}
}

func TestOpen_CreatesMissingDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newdir", "subdir", "GOTCHAS.md")
	// Parent dirs do not exist.

	if err := Open(path, "true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file and parent dirs were not created")
	}
}

func TestOpen_NoEditor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EDITOR", "")

	err := Open(path, "")
	if err == nil {
		t.Fatal("expected error for no editor configured, got nil")
	}
}

func TestResolveEditor_ConfigFirst(t *testing.T) {
	t.Setenv("EDITOR", "nano")
	got := resolveEditor("vim")
	if got != "vim" {
		t.Errorf("resolveEditor(%q) = %q, want %q", "vim", got, "vim")
	}
}

func TestResolveEditor_EnvFallback(t *testing.T) {
	t.Setenv("EDITOR", "nano")
	got := resolveEditor("")
	if got != "nano" {
		t.Errorf("resolveEditor(%q) = %q, want %q", "", got, "nano")
	}
}

func TestEnsureFile_MkdirError(t *testing.T) {
	dir := t.TempDir()
	// Create a file named "parent" — MkdirAll(parent) fails because it is not a directory.
	parentAsFile := filepath.Join(dir, "parent")
	if err := os.WriteFile(parentAsFile, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	// Attempt to create a child path through the file — MkdirAll must fail.
	path := filepath.Join(parentAsFile, "GOTCHAS.md")

	err := ensureFile(path)
	if err == nil {
		t.Fatal("expected error when parent path is a file, got nil")
	}
}
