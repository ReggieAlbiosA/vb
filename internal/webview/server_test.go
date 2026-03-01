//go:build gui

package webview

import (
	"strings"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/render"
)

func TestMarkdownToHTML_Valid(t *testing.T) {
	html, err := markdownToHTML([]byte("# Hello\n\nWorld."))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("expected <h1>Hello</h1>, got: %s", html)
	}
}

func TestMarkdownToHTML_Empty(t *testing.T) {
	html, err := markdownToHTML([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if html != "" {
		t.Errorf("expected empty HTML for empty input, got: %s", html)
	}
}

func TestNewRenderer_ReturnsWebview(t *testing.T) {
	r := NewRenderer()
	if _, ok := r.(*WebviewRenderer); !ok {
		t.Errorf("expected *WebviewRenderer, got %T", r)
	}
}

func TestBuildHTML_LightTheme(t *testing.T) {
	content := []byte("# Test\n\nLight theme content.")
	html, err := buildHTML(content, "refs", "light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "#ffffff") {
		t.Error("expected light theme background color in HTML")
	}
	if !strings.Contains(html, "REFS") {
		t.Error("expected REFS lens badge in HTML")
	}
}

func TestBuildHTML_CLIToolsLens(t *testing.T) {
	content := []byte("# Tools\n\nSome CLI tools documentation.\n")
	html, err := buildHTML(content, "cli-tools", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "CLI TOOLS") {
		t.Error("expected 'CLI TOOLS' lens badge in HTML")
	}
	if !strings.Contains(html, "Tools") {
		t.Error("expected heading in rendered HTML")
	}
}

func TestBuildHTML_MermaidFlowchart(t *testing.T) {
	content := []byte("flowchart LR\n  A-->B")
	html, err := buildHTML(content, "arch", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Error("expected mermaid pre block")
	}
}

func TestBuildHTML_MermaidPie(t *testing.T) {
	content := []byte("pie title Pets\n  \"Dogs\" : 30\n  \"Cats\" : 70")
	html, err := buildHTML(content, "arch", "light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Error("expected mermaid pre block for pie chart")
	}
	if !strings.Contains(html, "mermaid.min.js") {
		t.Error("expected mermaid script reference")
	}
}

func TestInitRegistersFactory(t *testing.T) {
	// The init() in renderer.go registers GUIRendererFactory with filename param.
	r := render.GUIRendererFactory("WHY.md")
	wr, ok := r.(*WebviewRenderer)
	if !ok {
		t.Fatalf("expected *WebviewRenderer from GUIRendererFactory, got %T", r)
	}
	if wr.Filename != "WHY.md" {
		t.Errorf("expected Filename 'WHY.md', got %q", wr.Filename)
	}
}

func TestBuildTabContent_Markdown(t *testing.T) {
	content := []byte("# Hello\n\nWorld.")
	tab, err := buildTabContent(content, "why", "WHY.md", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tab.Title != "WHY.md" {
		t.Errorf("expected title WHY.md, got %q", tab.Title)
	}
	if tab.Lens != "why" {
		t.Errorf("expected lens why, got %q", tab.Lens)
	}
	if tab.IsMermaid {
		t.Error("expected IsMermaid=false for markdown content")
	}
	if !strings.Contains(tab.BodyHTML, "<h1>Hello</h1>") {
		t.Errorf("expected rendered markdown HTML, got %q", tab.BodyHTML)
	}
	if tab.Theme != "dark" {
		t.Errorf("expected theme dark, got %q", tab.Theme)
	}
}

func TestBuildTabContent_Mermaid(t *testing.T) {
	content := []byte("graph TD\n  A-->B")
	tab, err := buildTabContent(content, "arch", "ARCH.mmd", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tab.IsMermaid {
		t.Error("expected IsMermaid=true for mermaid content")
	}
	if tab.BodyHTML != string(content) {
		t.Errorf("expected raw mermaid content, got %q", tab.BodyHTML)
	}
}

func TestBuildTabContent_Empty(t *testing.T) {
	tab, err := buildTabContent([]byte{}, "why", "WHY.md", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tab.BodyHTML != "" {
		t.Errorf("expected empty body for empty content, got %q", tab.BodyHTML)
	}
}
