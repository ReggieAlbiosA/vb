//go:build gui

package webview

import "testing"

func TestIsMermaidFile_GraphKeyword(t *testing.T) {
	if !isMermaidFile([]byte("graph TD\n  A-->B")) {
		t.Error("expected true for 'graph TD' content")
	}
}

func TestIsMermaidFile_SequenceDiagram(t *testing.T) {
	if !isMermaidFile([]byte("sequenceDiagram\n  Alice->>Bob: Hello")) {
		t.Error("expected true for 'sequenceDiagram' content")
	}
}

func TestIsMermaidFile_MarkdownContent(t *testing.T) {
	if isMermaidFile([]byte("# Heading\n\nSome markdown content.")) {
		t.Error("expected false for standard Markdown content")
	}
}

func TestIsMermaidFile_EmptyContent(t *testing.T) {
	if isMermaidFile([]byte{}) {
		t.Error("expected false for empty content")
	}
}

func TestMermaidExtension_MMD(t *testing.T) {
	if !MermaidExtension("/topics/disk/ARCH.mmd") {
		t.Error("expected true for .mmd extension")
	}
}

func TestMermaidExtension_MD(t *testing.T) {
	if MermaidExtension("/topics/disk/ARCH.md") {
		t.Error("expected false for .md extension")
	}
}
