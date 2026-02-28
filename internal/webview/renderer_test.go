//go:build gui

package webview

import (
	"strings"
	"testing"
)

func TestWebviewRenderer_FallbackOnHeadless(t *testing.T) {
	// Force headless by unsetting display variables.
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	r := &WebviewRenderer{}
	content := []byte("# Why\n\nBecause it matters.")
	out, err := r.Render(content, "why", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fallback produces terminal output (non-empty).
	if out == "" {
		t.Error("expected non-empty terminal fallback output")
	}
}

func TestBuildHTML_ValidMarkdown(t *testing.T) {
	content := []byte("# Hello\n\nWorld.")
	html, err := buildHTML(content, "why", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if html == "" {
		t.Error("expected non-empty HTML")
	}
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Error("expected HTML to contain rendered heading")
	}
	if !strings.Contains(html, "WHY") {
		t.Error("expected HTML to contain lens badge")
	}
}

func TestBuildHTML_MermaidContent(t *testing.T) {
	content := []byte("graph TD\n  A-->B")
	html, err := buildHTML(content, "arch", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Error("expected HTML to contain mermaid class for diagram content")
	}
	if !strings.Contains(html, "mermaid") {
		t.Error("expected HTML to include mermaid script reference")
	}
}
