//go:build gui

package webview

import (
	"bytes"
	"path/filepath"
	"strings"
)

// mermaidKeywords lists known Mermaid diagram type keywords.
var mermaidKeywords = []string{
	"graph ", "flowchart ", "sequenceDiagram", "classDiagram",
	"stateDiagram", "erDiagram", "gantt", "pie ", "gitGraph",
}

// isMermaidFile returns true if the content appears to be a Mermaid diagram
// (starts with a known diagram type keyword).
func isMermaidFile(content []byte) bool {
	trimmed := bytes.TrimSpace(content)
	for _, kw := range mermaidKeywords {
		if bytes.HasPrefix(trimmed, []byte(kw)) {
			return true
		}
	}
	return false
}

// MermaidExtension returns true if the file has a .mmd extension.
// Used by cmd layer to detect ARCH.mmd before passing content.
func MermaidExtension(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".mmd"
}
