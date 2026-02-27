package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/vault"
)

// mkVault creates a temporary directory with a .vb/ marker and a config.toml.
func mkVault(t *testing.T, toml string) string {
	t.Helper()
	dir := t.TempDir()
	vbDir := filepath.Join(dir, ".vb")
	if err := os.MkdirAll(vbDir, 0755); err != nil {
		t.Fatal(err)
	}
	if toml == "" {
		toml = "knowledge_path = \".\"\neditor = \"nano\"\ntheme = \"dark\"\nlint_on_save = false\n"
	}
	if err := os.WriteFile(filepath.Join(vbDir, "config.toml"), []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestResolve_FromVaultRoot: resolves correctly when CWD is the vault root itself.
func TestResolve_FromVaultRoot(t *testing.T) {
	vaultRoot := mkVault(t, "")

	ctx, err := vault.Resolve(vaultRoot)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if ctx.VaultRoot != vaultRoot {
		t.Errorf("VaultRoot: got %q, want %q", ctx.VaultRoot, vaultRoot)
	}
	// Default knowledge_path = "." resolves to vaultRoot itself.
	if ctx.TopicRoot != vaultRoot {
		t.Errorf("TopicRoot: got %q, want %q", ctx.TopicRoot, vaultRoot)
	}
}

// TestResolve_FromNestedDir: Stage 1 walks upward from a nested subdirectory.
func TestResolve_FromNestedDir(t *testing.T) {
	vaultRoot := mkVault(t, "")

	nested := filepath.Join(vaultRoot, "hardware", "disk", "subdir")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	ctx, err := vault.Resolve(nested)
	if err != nil {
		t.Fatalf("Resolve() error from nested dir: %v", err)
	}
	if ctx.VaultRoot != vaultRoot {
		t.Errorf("VaultRoot: got %q, want %q", ctx.VaultRoot, vaultRoot)
	}
}

// TestResolve_NoVault: returns ErrNoVault when no .vb/ exists anywhere.
func TestResolve_NoVault(t *testing.T) {
	// Use an isolated temp dir with no .vb/ marker anywhere in its ancestry
	// (t.TempDir() is typically under /tmp which has no .vb/).
	dir := t.TempDir()
	emptyDir := filepath.Join(dir, "no-vault-here")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := vault.Resolve(emptyDir)
	if err == nil {
		t.Fatal("Resolve() expected error for directory with no vault, got nil")
	}
	if err != vault.ErrNoVault {
		t.Errorf("Resolve() error: got %v, want ErrNoVault", err)
	}
}

// TestResolve_AbsoluteKnowledgePath: Stage 2 honours an absolute knowledge_path.
func TestResolve_AbsoluteKnowledgePath(t *testing.T) {
	externalTopics := t.TempDir()

	toml := "knowledge_path = \"" + externalTopics + "\"\neditor = \"nano\"\ntheme = \"dark\"\nlint_on_save = false\n"
	vaultRoot := mkVault(t, toml)

	ctx, err := vault.Resolve(vaultRoot)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if ctx.VaultRoot != vaultRoot {
		t.Errorf("VaultRoot: got %q, want %q", ctx.VaultRoot, vaultRoot)
	}
	if ctx.TopicRoot != externalTopics {
		t.Errorf("TopicRoot: got %q, want %q", ctx.TopicRoot, externalTopics)
	}
}

// TestResolve_RelativeKnowledgePath: relative knowledge_path resolves against VaultRoot.
func TestResolve_RelativeKnowledgePath(t *testing.T) {
	vaultRoot := mkVault(t, "knowledge_path = \"topics\"\neditor = \"nano\"\ntheme = \"dark\"\nlint_on_save = false\n")

	ctx, err := vault.Resolve(vaultRoot)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	want := filepath.Join(vaultRoot, "topics")
	if ctx.TopicRoot != want {
		t.Errorf("TopicRoot: got %q, want %q", ctx.TopicRoot, want)
	}
}
