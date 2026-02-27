package linter

import (
	"github.com/yuin/goldmark/ast"
)

// LensRule is the validation contract every lens rule implements.
type LensRule interface {
	Check(doc ast.Node, source []byte) []LintError
}

// LensRules maps lens flag names to their validation rules.
// Lenses without a rule entry are treated as schema-free (always valid).
// "used" is excluded — USED.md is an append-only log, not schema-validated.
var LensRules = map[string]LensRule{
	"why":        &WhyRule{},
	"importance": &ImportanceRule{},
	"cli-tools":  &CLIToolsRule{},
	"arch":       &ArchRule{},
	"gotchas":    &GotchasRule{},
	"refs":       &RefsRule{},
}

// WhyRule: must have at least one non-empty paragraph.
type WhyRule struct{}

func (r *WhyRule) Check(doc ast.Node, src []byte) []LintError {
	if !hasNodeType(doc, ast.KindParagraph, src) {
		return []LintError{{Lens: "why", Message: "WHY.md must contain at least one paragraph"}}
	}
	return nil
}

// ImportanceRule: must have at least one non-empty paragraph.
type ImportanceRule struct{}

func (r *ImportanceRule) Check(doc ast.Node, src []byte) []LintError {
	if !hasNodeType(doc, ast.KindParagraph, src) {
		return []LintError{{Lens: "importance", Message: "IMPORTANCE.md must contain at least one paragraph"}}
	}
	return nil
}

// CLIToolsRule: must have at least one fenced code block.
type CLIToolsRule struct{}

func (r *CLIToolsRule) Check(doc ast.Node, src []byte) []LintError {
	if !hasFencedCodeBlock(doc) {
		return []LintError{{Lens: "cli-tools", Message: "CLI_TOOLS.md must contain at least one fenced code block"}}
	}
	return nil
}

// ArchRule: must have at least one heading.
type ArchRule struct{}

func (r *ArchRule) Check(doc ast.Node, src []byte) []LintError {
	if !hasNodeType(doc, ast.KindHeading, src) {
		return []LintError{{Lens: "arch", Message: "ARCH.md must contain at least one heading"}}
	}
	return nil
}

// GotchasRule: must have at least one list item.
type GotchasRule struct{}

func (r *GotchasRule) Check(doc ast.Node, src []byte) []LintError {
	if !hasNodeType(doc, ast.KindListItem, src) {
		return []LintError{{Lens: "gotchas", Message: "GOTCHAS.md must contain at least one list item"}}
	}
	return nil
}

// RefsRule: must have at least one link.
type RefsRule struct{}

func (r *RefsRule) Check(doc ast.Node, src []byte) []LintError {
	if !hasNodeType(doc, ast.KindLink, src) {
		return []LintError{{Lens: "refs", Message: "REFS.md must contain at least one link"}}
	}
	return nil
}

// hasNodeType returns true if the AST contains at least one node of the given kind.
func hasNodeType(doc ast.Node, kind ast.NodeKind, src []byte) bool {
	found := false
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == kind {
			found = true
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}

// hasFencedCodeBlock returns true if the AST contains at least one fenced code block.
func hasFencedCodeBlock(doc ast.Node) bool {
	found := false
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindFencedCodeBlock {
			found = true
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}
