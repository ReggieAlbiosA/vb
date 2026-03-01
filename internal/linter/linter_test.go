package linter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/linter"
)

func TestLint_ValidWhy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	os.WriteFile(path, []byte("# Why\n\nThis topic exists because it matters."), 0o644)

	errs, err := linter.Lint(path, "why")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no violations, got: %v", errs)
	}
}

func TestLint_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	os.WriteFile(path, []byte{}, 0o644)

	errs, err := linter.Lint(path, "why")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Error("expected violation for empty file, got none")
	}
}

func TestLint_NoRuleForLens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "USED.md")
	os.WriteFile(path, []byte("- 2026-02-27 14:01 UTC  vb disk --why"), 0o644)

	// "used" has no rule — always valid, nil returned.
	errs, err := linter.Lint(path, "used")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errs != nil {
		t.Errorf("expected nil errors for schema-free lens, got: %v", errs)
	}
}

func TestLint_FileNotFound(t *testing.T) {
	_, err := linter.Lint("/nonexistent/path/WHY.md", "why")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLintError_ErrorMethod(t *testing.T) {
	e := linter.LintError{Lens: "why", Message: "must have a paragraph"}
	got := e.Error()
	want := "[why] must have a paragraph"
	if got != want {
		t.Errorf("LintError.Error() = %q, want %q", got, want)
	}
}
