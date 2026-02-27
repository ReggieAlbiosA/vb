package render

import (
	"fmt"
	"os"
)

// Renderer is the abstraction both terminal and GUI renderers implement.
// Phase 06 will provide a WebviewRenderer satisfying this interface.
type Renderer interface {
	Render(content []byte, lens string, theme string) (string, error)
}

// File is the public entry point called by cmd/query.go.
// It reads the file at path and dispatches to the appropriate renderer.
func File(path string, lens string, gui bool, theme string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	if gui {
		// Phase 06 — hand off to webview renderer.
		// No-op until Phase 06 is implemented; falls through to terminal.
	}

	r := &TerminalRenderer{}
	out, err := r.Render(content, lens, theme)
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}
