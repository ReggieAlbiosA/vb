package render

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

// TerminalRenderer renders Markdown as styled terminal output via Glamour.
type TerminalRenderer struct{}

// Render takes raw Markdown bytes and returns styled terminal output.
// Raw bytes are fed directly to Glamour — no pre-parsing step.
func (r *TerminalRenderer) Render(content []byte, lens string, theme string) (string, error) {
	// Use vault config theme ("dark" | "light"), not WithAutoStyle().
	// WithAutoStyle() ignores the user's explicit config preference.
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(theme),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", fmt.Errorf("creating renderer: %w", err)
	}

	// Feed raw Markdown directly — Glamour parses internally via Goldmark.
	rendered, err := renderer.Render(string(content))
	if err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}

	// Prepend lens badge above Glamour output.
	badge := LensBadge(lens)
	return badge + rendered, nil
}
