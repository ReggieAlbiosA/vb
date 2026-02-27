package hook_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/hook"
)

// captureOutput redirects os.Stdout and os.Stderr during f(), returning what was written.
func captureOutput(t *testing.T, f func()) (stdout, stderr string) {
	t.Helper()

	// Capture stdout.
	oldStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = wOut

	// Capture stderr.
	oldStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = wErr

	f()

	wOut.Close()
	wErr.Close()

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut) //nolint:errcheck
	io.Copy(&bufErr, rErr) //nolint:errcheck

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return bufOut.String(), bufErr.String()
}

func TestOnSave_LintOnSaveFalse(t *testing.T) {
	// lintOnSave=false → returns immediately; passing a non-existent path is safe.
	_, stderr := captureOutput(t, func() {
		hook.OnSave("/nonexistent/path.md", "why", false)
	})
	if stderr != "" {
		t.Errorf("expected no stderr output when lintOnSave=false, got: %q", stderr)
	}
}

func TestOnSave_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	os.WriteFile(path, []byte("# Why\n\nBecause it matters."), 0o644)

	stdout, _ := captureOutput(t, func() {
		hook.OnSave(path, "why", true)
	})

	if !strings.Contains(stdout, "schema valid") {
		t.Errorf("expected '✔ why: schema valid' in stdout, got: %q", stdout)
	}
}

func TestOnSave_LintReadError(t *testing.T) {
	// Non-existent file + lintOnSave=true → linter returns error → printed to stderr.
	_, stderr := captureOutput(t, func() {
		hook.OnSave("/nonexistent/path/WHY.md", "why", true)
	})
	if stderr == "" {
		t.Error("expected stderr output for lint read error, got nothing")
	}
}

func TestOnSave_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WHY.md")
	os.WriteFile(path, []byte("# Heading only — no paragraph"), 0o644)

	_, stderr := captureOutput(t, func() {
		hook.OnSave(path, "why", true)
	})

	if !strings.Contains(stderr, "⚠") {
		t.Errorf("expected '⚠' warning in stderr for invalid file, got: %q", stderr)
	}
}
