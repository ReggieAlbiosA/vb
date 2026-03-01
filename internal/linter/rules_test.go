package linter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/linter"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertNoViolations(t *testing.T, path, lens string) {
	t.Helper()
	errs, err := linter.Lint(path, lens)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no violations, got: %v", errs)
	}
}

func assertViolation(t *testing.T, path, lens string) {
	t.Helper()
	errs, err := linter.Lint(path, lens)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if len(errs) == 0 {
		t.Error("expected at least one violation, got none")
	}
}

func TestImportanceRule_HasParagraph(t *testing.T) {
	path := writeTemp(t, "IMPORTANCE.md", "# Importance\n\nThis topic is critical for production.")
	assertNoViolations(t, path, "importance")
}

func TestImportanceRule_NoParagraph(t *testing.T) {
	path := writeTemp(t, "IMPORTANCE.md", "# Importance only heading\n")
	assertViolation(t, path, "importance")
}

func TestWhyRule_HasParagraph(t *testing.T) {
	path := writeTemp(t, "WHY.md", "# Why\n\nThis is a paragraph explaining the topic.")
	assertNoViolations(t, path, "why")
}

func TestWhyRule_NoParagraph(t *testing.T) {
	// Heading only — no paragraph body.
	path := writeTemp(t, "WHY.md", "# Why disk\n")
	assertViolation(t, path, "why")
}

func TestCLIToolsRule_HasCodeBlock(t *testing.T) {
	path := writeTemp(t, "CLI_TOOLS.md", "# Tools\n\n```bash\nlsblk\n```\n")
	assertNoViolations(t, path, "cli-tools")
}

func TestCLIToolsRule_NoCodeBlock(t *testing.T) {
	path := writeTemp(t, "CLI_TOOLS.md", "# Tools\n\nJust some prose, no code.\n")
	assertViolation(t, path, "cli-tools")
}

func TestArchRule_HasHeading(t *testing.T) {
	path := writeTemp(t, "ARCH.md", "# Architecture\n\nSome overview text.\n")
	assertNoViolations(t, path, "arch")
}

func TestArchRule_NoHeading(t *testing.T) {
	path := writeTemp(t, "ARCH.md", "Just prose, no headings.\n")
	assertViolation(t, path, "arch")
}

func TestGotchasRule_HasListItem(t *testing.T) {
	path := writeTemp(t, "GOTCHAS.md", "# Gotchas\n\n- Watch out for X\n- Also Y\n")
	assertNoViolations(t, path, "gotchas")
}

func TestGotchasRule_NoListItem(t *testing.T) {
	path := writeTemp(t, "GOTCHAS.md", "# Gotchas\n\nJust prose, no list.\n")
	assertViolation(t, path, "gotchas")
}

func TestRefsRule_HasLink(t *testing.T) {
	path := writeTemp(t, "REFS.md", "# Refs\n\n- [Official docs](https://example.com)\n")
	assertNoViolations(t, path, "refs")
}

func TestRefsRule_NoLink(t *testing.T) {
	path := writeTemp(t, "REFS.md", "# Refs\n\nNo links here, just prose.\n")
	assertViolation(t, path, "refs")
}
