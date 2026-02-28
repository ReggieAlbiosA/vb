//go:build !gui

package webview_test

import (
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/render"
	"github.com/ReggieAlbiosA/vb/internal/webview"
)

func TestNewRenderer_IsTerminal(t *testing.T) {
	r := webview.NewRenderer()
	if _, ok := r.(*render.TerminalRenderer); !ok {
		t.Errorf("expected *render.TerminalRenderer, got %T", r)
	}
}
