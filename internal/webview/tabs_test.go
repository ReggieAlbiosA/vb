//go:build gui

package webview

import (
	"strings"
	"testing"
)

func TestThemeColors_Dark(t *testing.T) {
	bg, fg, _, _, _, _, _, _, _, _ := themeColors("dark")
	if bg != "#1e1e1e" {
		t.Errorf("expected dark bg #1e1e1e, got %s", bg)
	}
	if fg != "#e0e0e0" {
		t.Errorf("expected dark fg #e0e0e0, got %s", fg)
	}
}

func TestThemeColors_Light(t *testing.T) {
	bg, fg, _, _, _, _, _, _, _, _ := themeColors("light")
	if bg != "#ffffff" {
		t.Errorf("expected light bg #ffffff, got %s", bg)
	}
	if fg != "#1a1a1a" {
		t.Errorf("expected light fg #1a1a1a, got %s", fg)
	}
}

func TestBuildTabShell_Basic(t *testing.T) {
	tab := TabData{
		Title:    "WHY.md",
		Lens:     "why",
		BodyHTML: "<h1>Hello</h1><p>World.</p>",
		Theme:    "dark",
	}

	html, err := buildTabShell(tab)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"vb-tab-bar",
		"WHY.md",
		"WHY",
		"<h1>Hello</h1>",
		"#1e1e1e",
	}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("expected HTML to contain %q", c)
		}
	}
}

func TestBuildTabShell_Mermaid(t *testing.T) {
	tab := TabData{
		Title:     "ARCH.mmd",
		Lens:      "arch",
		BodyHTML:  "graph TD\n  A-->B",
		IsMermaid: true,
		Theme:     "dark",
	}

	html, err := buildTabShell(tab)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "mermaid.min.js") {
		t.Error("expected mermaid script reference for mermaid tab")
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Error("expected mermaid pre block")
	}
}

func TestBuildTabShell_LightTheme(t *testing.T) {
	tab := TabData{
		Title:    "WHY.md",
		Lens:     "why",
		BodyHTML: "<p>Light content</p>",
		Theme:    "light",
	}

	html, err := buildTabShell(tab)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "#ffffff") {
		t.Error("expected light background color")
	}
}

func TestBuildTabShell_CLIToolsLens(t *testing.T) {
	tab := TabData{
		Title:    "CLI-TOOLS.md",
		Lens:     "cli-tools",
		BodyHTML: "<p>Tools</p>",
		Theme:    "dark",
	}

	html, err := buildTabShell(tab)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "CLI TOOLS") {
		t.Error("expected 'CLI TOOLS' badge (hyphen replaced with space)")
	}
}

func TestTabAddJS_Basic(t *testing.T) {
	tab := TabData{
		Title:    "REFS.md",
		Lens:     "refs",
		BodyHTML: "<p>Reference links</p>",
		Theme:    "dark",
	}

	js := tabAddJS("tab1", tab)

	if !strings.Contains(js, "vbAddTab(") {
		t.Error("expected vbAddTab call")
	}
	if !strings.Contains(js, "tab1") {
		t.Error("expected tab ID in JS")
	}
	if !strings.Contains(js, "REFS.md") {
		t.Error("expected tab title in JS")
	}
	if !strings.Contains(js, "REFS") {
		t.Error("expected lens badge in JS body")
	}
}

func TestTabAddJS_Mermaid(t *testing.T) {
	tab := TabData{
		Title:     "ARCH.mmd",
		Lens:      "arch",
		BodyHTML:  "flowchart LR\n  A-->B",
		IsMermaid: true,
		Theme:     "dark",
	}

	js := tabAddJS("tab2", tab)

	if !strings.Contains(js, "true") {
		t.Error("expected isMermaid=true in JS call")
	}
	if !strings.Contains(js, "mermaid") {
		t.Error("expected mermaid class in body HTML")
	}
}

func TestTabAddJS_XSSEscaping(t *testing.T) {
	tab := TabData{
		Title:    "<script>alert('xss')</script>",
		Lens:     "why",
		BodyHTML: "<p>Safe content</p>",
		Theme:    "dark",
	}

	js := tabAddJS("tab3", tab)

	if strings.Contains(js, "<script>alert") {
		t.Error("expected XSS in title to be escaped")
	}
	if !strings.Contains(js, "&lt;script&gt;") {
		t.Error("expected HTML-escaped title")
	}
}

func TestTabAddJS_QuoteEscaping(t *testing.T) {
	tab := TabData{
		Title:    "it's a test",
		Lens:     "why",
		BodyHTML: "<p>Content with 'quotes'</p>",
		Theme:    "dark",
	}

	js := tabAddJS("tab4", tab)

	// Single quotes in JS string must be escaped.
	if strings.Contains(js, "it's") {
		t.Error("expected single quote in title to be escaped for JS")
	}
}
