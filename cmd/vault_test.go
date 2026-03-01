package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/registry"
)

// TestVaultCreate_Success: creates dir, .vb/ scaffold, and registers in global registry.
func TestVaultCreate_Success(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	vaultPath := filepath.Join(t.TempDir(), "myvault")

	out, err := execCmd(t, t.TempDir(), "vault", "create", "test", vaultPath)
	if err != nil {
		t.Fatalf("vault create error: %v\noutput: %s", err, out)
	}

	// .vb/ scaffold exists.
	for _, name := range []string{".vb", ".vb/config.toml", ".vb/index.json"} {
		if _, err := os.Stat(filepath.Join(vaultPath, name)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", name)
		}
	}

	// Registered in global registry.
	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg.Vaults["test"] != vaultPath {
		t.Errorf("registry Vaults[test] = %q, want %q", reg.Vaults["test"], vaultPath)
	}
	// First vault is auto-default.
	if reg.Default != "test" {
		t.Errorf("registry Default = %q, want %q", reg.Default, "test")
	}

	if !strings.Contains(out, "set as default") {
		t.Errorf("output missing 'set as default': %s", out)
	}
}

// TestVaultCreate_AlreadyRegistered: errors when name is already taken.
func TestVaultCreate_AlreadyRegistered(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path1 := filepath.Join(t.TempDir(), "vault1")
	path2 := filepath.Join(t.TempDir(), "vault2")

	if _, err := execCmd(t, t.TempDir(), "vault", "create", "dup", path1); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := execCmd(t, t.TempDir(), "vault", "create", "dup", path2)
	if err == nil {
		t.Fatal("duplicate name: expected error, got nil")
	}
}

// TestVaultCreate_NestedInVault: errors when path is inside an existing vault.
func TestVaultCreate_NestedInVault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	parent := filepath.Join(t.TempDir(), "parent")

	if _, err := execCmd(t, t.TempDir(), "vault", "create", "outer", parent); err != nil {
		t.Fatalf("create outer: %v", err)
	}

	nested := filepath.Join(parent, "sub", "inner")
	_, err := execCmd(t, t.TempDir(), "vault", "create", "inner", nested)
	if err == nil {
		t.Fatal("nested vault: expected error, got nil")
	}
}

// TestVaultList_Empty: shows "no vaults" message.
func TestVaultList_Empty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out, err := execCmd(t, t.TempDir(), "vault", "list")
	if err != nil {
		t.Fatalf("vault list error: %v", err)
	}
	if !strings.Contains(out, "no vaults registered") {
		t.Errorf("expected 'no vaults registered', got: %s", out)
	}
}

// TestVaultList_WithVaults: shows registered vaults with default marker.
func TestVaultList_WithVaults(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	// Seed a registry directly.
	reg := registry.Registry{
		Default: "alpha",
		Vaults: map[string]string{
			"alpha": "/path/alpha",
			"beta":  "/path/beta",
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatal(err)
	}

	out, err := execCmd(t, t.TempDir(), "vault", "list")
	if err != nil {
		t.Fatalf("vault list error: %v", err)
	}
	if !strings.Contains(out, "* alpha") {
		t.Errorf("missing default marker for alpha: %s", out)
	}
	if !strings.Contains(out, "beta") {
		t.Errorf("missing beta: %s", out)
	}
}

// TestVaultUse_Success: sets the default vault.
func TestVaultUse_Success(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	// Seed a registry.
	reg := registry.Registry{
		Default: "alpha",
		Vaults: map[string]string{
			"alpha": "/path/alpha",
			"beta":  "/path/beta",
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatal(err)
	}

	out, err := execCmd(t, t.TempDir(), "vault", "use", "beta")
	if err != nil {
		t.Fatalf("vault use error: %v\noutput: %s", err, out)
	}

	// Verify registry was updated.
	updated, _ := registry.Load()
	if updated.Default != "beta" {
		t.Errorf("Default = %q, want %q", updated.Default, "beta")
	}
}

// TestVaultUse_NotFound: errors when name is not registered.
func TestVaultUse_NotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := execCmd(t, t.TempDir(), "vault", "use", "missing")
	if err == nil {
		t.Fatal("vault use missing: expected error, got nil")
	}
}

// TestVaultRemove_Success: unregisters vault, files unchanged.
func TestVaultRemove_Success(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	vaultPath := filepath.Join(t.TempDir(), "removeme")

	if _, err := execCmd(t, t.TempDir(), "vault", "create", "rm", vaultPath); err != nil {
		t.Fatalf("create: %v", err)
	}

	out, err := execCmd(t, t.TempDir(), "vault", "remove", "rm")
	if err != nil {
		t.Fatalf("vault remove error: %v\noutput: %s", err, out)
	}

	// Registry should be empty.
	reg, _ := registry.Load()
	if _, exists := reg.Vaults["rm"]; exists {
		t.Error("vault still registered after remove")
	}

	// Files should still exist.
	if _, err := os.Stat(filepath.Join(vaultPath, ".vb")); os.IsNotExist(err) {
		t.Error(".vb/ deleted — should have been preserved")
	}
}

// TestVaultRemove_WasDefault: clears default when removing the default vault.
func TestVaultRemove_WasDefault(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	reg := registry.Registry{
		Default: "only",
		Vaults:  map[string]string{"only": "/path/only"},
	}
	if err := reg.Save(); err != nil {
		t.Fatal(err)
	}

	out, err := execCmd(t, t.TempDir(), "vault", "remove", "only")
	if err != nil {
		t.Fatalf("vault remove error: %v", err)
	}

	updated, _ := registry.Load()
	if updated.Default != "" {
		t.Errorf("Default = %q after removing only vault, want empty", updated.Default)
	}
	if !strings.Contains(out, "no default vault set") {
		t.Errorf("output missing default warning: %s", out)
	}
}

// TestVaultRemove_NotFound: errors when name is not registered.
func TestVaultRemove_NotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := execCmd(t, t.TempDir(), "vault", "remove", "ghost")
	if err == nil {
		t.Fatal("vault remove missing: expected error, got nil")
	}
}

// TestInitCmd_WithName: vb init --name registers vault in global registry.
func TestInitCmd_WithName(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	dir := t.TempDir()

	// Reset flag state.
	flagInitName = ""
	t.Cleanup(func() { flagInitName = "" })

	out, err := execCmd(t, dir, "init", "--name", "myknowledge")
	if err != nil {
		t.Fatalf("vb init --name error: %v\noutput: %s", err, out)
	}

	// Vault should be registered.
	reg, _ := registry.Load()
	if reg.Vaults["myknowledge"] != dir {
		t.Errorf("registry missing vault: got %q", reg.Vaults["myknowledge"])
	}
	if reg.Default != "myknowledge" {
		t.Errorf("Default = %q, want %q", reg.Default, "myknowledge")
	}

	if !strings.Contains(out, "registered as") {
		t.Errorf("output missing registration message: %s", out)
	}
}

// TestResolveVault_DefaultFallback: resolveVault() falls through to default vault
// when not inside any vault directory.
func TestResolveVault_DefaultFallback(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	// Create a vault at a known path.
	vaultPath := filepath.Join(t.TempDir(), "fallback")
	if _, err := execCmd(t, t.TempDir(), "vault", "create", "fb", vaultPath); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Create a topic so reindex has something to index.
	topicDir := filepath.Join(vaultPath, "testtopic")
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(topicDir, "WHY.md"), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run reindex from an unrelated directory — should use default vault.
	emptyDir := filepath.Join(t.TempDir(), "nowhere")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	out, err := execCmd(t, emptyDir, "reindex")
	if err != nil {
		t.Fatalf("reindex from outside vault: %v\noutput: %s", err, out)
	}

	// Verify index was rebuilt.
	data, _ := os.ReadFile(filepath.Join(vaultPath, ".vb", "index.json"))
	var idx struct {
		Topics map[string]string `json:"topics"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("parsing index: %v", err)
	}
	if _, exists := idx.Topics["testtopic"]; !exists {
		t.Errorf("index missing testtopic: %s", string(data))
	}
}
