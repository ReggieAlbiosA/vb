package render_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/render"
)

func TestFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	if err := os.WriteFile(path, []byte("# Why\n\nBecause it matters."), 0644); err != nil {
		t.Fatal(err)
	}

	if err := render.File(path, "why", false, "dark"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFile_FileNotFound(t *testing.T) {
	err := render.File("/nonexistent/path/WHY.md", "why", false, "dark")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestFile_GUIFallthrough(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	if err := os.WriteFile(path, []byte("# Why\n\nBecause it matters."), 0644); err != nil {
		t.Fatal(err)
	}

	// In non-GUI builds, GUIRendererFactory is nil — gui=true falls through to terminal.
	if err := render.File(path, "why", true, "dark"); err != nil {
		t.Fatalf("unexpected error with gui=true: %v", err)
	}
}

func TestFile_GUIRendererFactory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	if err := os.WriteFile(path, []byte("# Why\n\nBecause it matters."), 0644); err != nil {
		t.Fatal(err)
	}

	// Set a custom factory that returns TerminalRenderer (simulates stub build).
	old := render.GUIRendererFactory
	render.GUIRendererFactory = func(filename string) render.Renderer {
		return &render.TerminalRenderer{}
	}
	t.Cleanup(func() { render.GUIRendererFactory = old })

	if err := render.File(path, "why", true, "dark"); err != nil {
		t.Fatalf("unexpected error with GUIRendererFactory set: %v", err)
	}
}
