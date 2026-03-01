//go:build !gui

package webview

import "github.com/ReggieAlbiosA/vb/internal/render"

// NewRenderer returns a terminal renderer as the fallback when
// the gui build tag is not set.
func NewRenderer() render.Renderer {
	return &render.TerminalRenderer{}
}
