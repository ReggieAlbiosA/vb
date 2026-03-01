package linter

import (
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

// Lint parses the Markdown file at filePath and runs the rule for the given lens.
// Returns a slice of LintErrors (empty slice = valid). Never modifies the file.
// If no rule is defined for the lens, returns nil (always valid).
func Lint(filePath, lens string) ([]LintError, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	rule, ok := LensRules[lens]
	if !ok {
		// No rule defined for this lens — treated as always valid.
		return nil, nil
	}

	// Parse to AST using Goldmark. No ring — structural inspection only.
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(content))

	return rule.Check(doc, content), nil
}
