package render

import (
	"fmt"
	"os"
)

// Renderer is the abstraction both terminal and GUI renderers implement.
type Renderer interface {
	Render(content []byte, lens string, theme string) (string, error)
}

// GUIRendererFactory is set by the webview package (via init()) to provide
// the GUI renderer when compiled with -tags gui. When nil, gui=true falls
// through to the terminal renderer.
var GUIRendererFactory func() Renderer

// File is the public entry point called by cmd/query.go.
// It reads the file at path and dispatches to the appropriate renderer.
func File(path string, lens string, gui bool, theme string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	var r Renderer
	if gui && GUIRendererFactory != nil {
		r = GUIRendererFactory()
	} else {
		r = &TerminalRenderer{}
	}

	out, err := r.Render(content, lens, theme)
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}
