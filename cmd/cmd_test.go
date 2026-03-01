package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// isolateRegistry sets XDG_CONFIG_HOME to a temp dir so tests never touch
// the real ~/.config/vb/vaults.json. Returns the isolated config dir.
// Call this ONCE at the start of each test function.
func isolateRegistry(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return dir
}

// execCmd runs the root cobra command with the given args from a specific directory.
// It returns the combined stdout/stderr output and any error.
// Callers MUST call isolateRegistry(t) before using this helper.
func execCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()

	// Chdir to the target directory so os.Getwd() inside commands returns dir.
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(original) })

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	// Reset args after test so state doesn't bleed.
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, execErr := rootCmd.ExecuteC()
	return buf.String(), execErr
}

// TestInitCmd_Success: vb init creates .vb/, config.toml, and index.json.
func TestInitCmd_Success(t *testing.T) {
	isolateRegistry(t)
	dir := t.TempDir()

	_, err := execCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("vb init error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(dir, ".vb"),
		filepath.Join(dir, ".vb", "config.toml"),
		filepath.Join(dir, ".vb", "index.json"),
	}
	for _, f := range expectedFiles {
		if _, statErr := os.Stat(f); os.IsNotExist(statErr) {
			t.Errorf("vb init: expected %s to exist, but it does not", f)
		}
	}
}

// TestInitCmd_AlreadyInitialized: vb init errors when .vb/ already exists.
func TestInitCmd_AlreadyInitialized(t *testing.T) {
	isolateRegistry(t)
	dir := t.TempDir()

	// First init.
	if _, err := execCmd(t, dir, "init"); err != nil {
		t.Fatalf("first vb init error: %v", err)
	}

	// Second init should fail.
	_, err := execCmd(t, dir, "init")
	if err == nil {
		t.Fatal("second vb init: expected error for already-initialized vault, got nil")
	}
}

// TestReindexCmd_Success: vb reindex indexes topics and reports count.
func TestReindexCmd_Success(t *testing.T) {
	isolateRegistry(t)
	dir := t.TempDir()

	// Init vault.
	if _, err := execCmd(t, dir, "init"); err != nil {
		t.Fatalf("vb init error: %v", err)
	}

	// Create a topic.
	topicDir := filepath.Join(dir, "hardware", "disk")
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(topicDir, "WHY.md"), []byte("# Disk"), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := execCmd(t, dir, "reindex")
	if err != nil {
		t.Fatalf("vb reindex error: %v", err)
	}
	_ = out // output verified via index.json content below

	// Confirm index.json has disk entry.
	data, readErr := os.ReadFile(filepath.Join(dir, ".vb", "index.json"))
	if readErr != nil {
		t.Fatalf("reading index.json: %v", readErr)
	}
	if !bytes.Contains(data, []byte(`"disk"`)) {
		t.Errorf("index.json missing 'disk' entry, got: %s", data)
	}
}

// TestReindexCmd_NoVault: vb reindex errors when run outside any vault.
func TestReindexCmd_NoVault(t *testing.T) {
	isolateRegistry(t)
	// Use an isolated temp dir with no .vb/ marker.
	dir := t.TempDir()
	emptyDir := filepath.Join(dir, "no-vault")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := execCmd(t, emptyDir, "reindex")
	if err == nil {
		t.Fatal("vb reindex: expected error outside a vault, got nil")
	}
}
