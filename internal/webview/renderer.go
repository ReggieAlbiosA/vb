//go:build gui

package webview

import (
	"fmt"

	"github.com/ReggieAlbiosA/vb/internal/render"
)

func init() {
	render.GUIRendererFactory = func() render.Renderer {
		return &WebviewRenderer{}
	}
}

// WebviewRenderer satisfies render.Renderer by launching a webview window.
type WebviewRenderer struct{}

// Render converts Markdown content to HTML and opens it in a webview window.
// Falls back to terminal output if the display is unavailable.
func (r *WebviewRenderer) Render(content []byte, lens string, theme string) (string, error) {
	html, err := buildHTML(content, lens, theme)
	if err != nil {
		return "", fmt.Errorf("building webview HTML: %w", err)
	}

	if err := openWindow(lens, html); err != nil {
		// Graceful fallback: headless or display unavailable.
		t := &render.TerminalRenderer{}
		return t.Render(content, lens, theme)
	}

	// Webview blocks until browser is opened — output handled by browser.
	return "", nil
}

// NewRenderer returns a WebviewRenderer (gui build tag active).
func NewRenderer() render.Renderer {
	return &WebviewRenderer{}
}
