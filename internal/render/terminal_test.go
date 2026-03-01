package render_test

import (
	"strings"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/render"
)

func TestTerminalRenderer_DarkTheme(t *testing.T) {
	r := &render.TerminalRenderer{}
	_, err := r.Render([]byte("# Dark\n\nContent here."), "why", "dark")
	if err != nil {
		t.Fatalf("dark theme render error: %v", err)
	}
}

func TestTerminalRenderer_LightTheme(t *testing.T) {
	r := &render.TerminalRenderer{}
	_, err := r.Render([]byte("# Light\n\nContent here."), "why", "light")
	if err != nil {
		t.Fatalf("light theme render error: %v", err)
	}
}

func TestTerminalRenderer_OutputContainsBadge(t *testing.T) {
	r := &render.TerminalRenderer{}
	out, err := r.Render([]byte("# Title\n\nBody."), "arch", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ARCH") {
		t.Errorf("output does not contain badge text 'ARCH': %q", out)
	}
}

func TestTerminalRenderer_ContentUnmodified(t *testing.T) {
	content := "# Structure\n\nThis exact word: xyzzy"
	r := &render.TerminalRenderer{}
	out, err := r.Render([]byte(content), "why", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "xyzzy") {
		t.Errorf("rendered output does not contain source text 'xyzzy': %q", out)
	}
}

func TestTerminalRenderer_InvalidTheme(t *testing.T) {
	r := &render.TerminalRenderer{}
	_, err := r.Render([]byte("# Title"), "why", "invalid-style-that-does-not-exist")
	if err == nil {
		t.Fatal("expected error for invalid theme, got nil")
	}
}
