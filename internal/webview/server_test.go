//go:build gui

package webview

import (
	"net/http"
	"net/http/httptest"
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

func TestHtmlHandler_ServesContent(t *testing.T) {
	handler := htmlHandler("<h1>Test</h1>")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content type, got %q", ct)
	}
	if !strings.Contains(rec.Body.String(), "<h1>Test</h1>") {
		t.Errorf("expected body to contain <h1>Test</h1>, got: %s", rec.Body.String())
	}
}

func TestOpenWindow_ListenFails(t *testing.T) {
	// Force headless to test the fallback path in openWindow.
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	err := openWindow("test", "<html></html>")
	if err == nil {
		t.Error("expected error in headless environment, got nil")
	}
}

func TestLaunchBrowser_NoDisplay(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	err := launchBrowser("http://127.0.0.1:0")
	if err == nil {
		t.Error("expected error when DISPLAY is unset")
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
	// The init() in renderer.go registers GUIRendererFactory.
	// Calling the factory exercises the inner closure.
	r := render.GUIRendererFactory()
	if _, ok := r.(*WebviewRenderer); !ok {
		t.Errorf("expected *WebviewRenderer from GUIRendererFactory, got %T", r)
	}
}

func TestHasDisplay_NoDisplay(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	if hasDisplay() {
		t.Error("expected false when no display vars set")
	}
}

func TestHasDisplay_WithDisplay(t *testing.T) {
	t.Setenv("DISPLAY", ":0")
	t.Setenv("WAYLAND_DISPLAY", "")
	if !hasDisplay() {
		t.Error("expected true when DISPLAY is set")
	}
}

func TestHasDisplay_WithWayland(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	if !hasDisplay() {
		t.Error("expected true when WAYLAND_DISPLAY is set")
	}
}

func TestLaunchBrowser_WithFakeDisplay(t *testing.T) {
	// Set DISPLAY so the display check passes, but use a bogus display
	// that won't actually open a browser window.
	t.Setenv("DISPLAY", ":99")
	t.Setenv("WAYLAND_DISPLAY", "")

	// xdg-open will be started but won't successfully open anything with :99.
	// We just care that the code path is exercised (Start() returns nil or error).
	_ = launchBrowser("http://127.0.0.1:0")
}
