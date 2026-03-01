package resolver_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/resolver"
)

func TestBind_FileExists(t *testing.T) {
	dir := t.TempDir()
	lensFile := "WHY.md"
	fullPath := filepath.Join(dir, lensFile)
	if err := os.WriteFile(fullPath, []byte("# Why"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolver.Bind(dir, lensFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != fullPath {
		t.Errorf("got %q, want %q", got, fullPath)
	}
}

func TestBind_FileMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := resolver.Bind(dir, "WHY.md")
	if !errors.Is(err, resolver.ErrLensFileMissing) {
		t.Errorf("expected ErrLensFileMissing, got %v", err)
	}
}

func TestBind_ErrorMessageContainsTopic(t *testing.T) {
	// Use a topicDir whose base name is "disk" so the error message names it.
	dir := filepath.Join(t.TempDir(), "disk")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := resolver.Bind(dir, "WHY.md")
	if err == nil {
		t.Fatal("expected error for missing lens file, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "disk") {
		t.Errorf("error message %q does not contain topic name 'disk'", msg)
	}
	if !strings.Contains(msg, "vb edit") {
		t.Errorf("error message %q does not contain 'vb edit' hint", msg)
	}
}
