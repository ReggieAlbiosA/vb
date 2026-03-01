//go:build gui

package webview

import (
	"fmt"

	"github.com/ReggieAlbiosA/vb/internal/render"
)

func init() {
	render.GUIRendererFactory = func(filename string) render.Renderer {
		return &WebviewRenderer{Filename: filename}
	}
}

// WebviewRenderer satisfies render.Renderer by opening content in a native
// webview window with tab support via IPC.
type WebviewRenderer struct {
	Filename string // e.g. "WHY.md" — used as tab title
}

// Render converts Markdown content to HTML and opens it in a native webview.
// If an existing vb GUI window is running, adds a tab via IPC and returns
// immediately. Otherwise starts a new primary window (blocking until closed).
// Falls back to terminal output if no display is available.
func (r *WebviewRenderer) Render(content []byte, lens string, theme string) (string, error) {
	if !hasDisplay() {
		t := &render.TerminalRenderer{}
		return t.Render(content, lens, theme)
	}

	tab, err := buildTabContent(content, lens, r.Filename, theme)
	if err != nil {
		return "", fmt.Errorf("building tab content: %w", err)
	}

	// Try to send tab to an existing window via IPC.
	if err := tryIPCSend(tab); err == nil {
		// Tab added to existing window — exit immediately.
		return "", nil
	}

	// No existing window — become the primary window (blocks until closed).
	if err := startPrimaryWindow(tab); err != nil {
		// Graceful fallback: display issue at runtime.
		t := &render.TerminalRenderer{}
		return t.Render(content, lens, theme)
	}

	return "", nil
}

// NewRenderer returns a WebviewRenderer (gui build tag active).
func NewRenderer() render.Renderer {
	return &WebviewRenderer{}
}
