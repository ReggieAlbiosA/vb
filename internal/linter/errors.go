package linter

import "fmt"

// LintError represents a single schema violation in a lens file.
type LintError struct {
	Lens    string // The lens name (e.g. "why")
	Message string // Human-readable description of the violation
}

func (e LintError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Lens, e.Message)
}
