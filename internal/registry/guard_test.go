package registry

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckNesting_Clean(t *testing.T) {
	// A path with no .vb/ anywhere should pass.
	dir := filepath.Join(t.TempDir(), "newvault")
	if err := CheckNesting(dir); err != nil {
		t.Fatalf("CheckNesting() on clean path: %v", err)
	}
}

func TestCheckNesting_AncestorVault(t *testing.T) {
	parent := t.TempDir()
	// Create .vb/ in parent — simulates an existing vault.
	if err := os.MkdirAll(filepath.Join(parent, ".vb"), 0755); err != nil {
		t.Fatal(err)
	}

	child := filepath.Join(parent, "sub", "nested")
	err := CheckNesting(child)
	if err == nil {
		t.Fatal("CheckNesting() inside existing vault: expected error, got nil")
	}
	if !errors.Is(err, ErrNestedVault) {
		t.Errorf("expected ErrNestedVault, got: %v", err)
	}
}

func TestCheckNesting_DescendantVault(t *testing.T) {
	parent := t.TempDir()
	// Create a sub-vault inside the target path.
	subVault := filepath.Join(parent, "inner", ".vb")
	if err := os.MkdirAll(subVault, 0755); err != nil {
		t.Fatal(err)
	}

	err := CheckNesting(parent)
	if err == nil {
		t.Fatal("CheckNesting() wrapping existing vault: expected error, got nil")
	}
	if !errors.Is(err, ErrNestedVault) {
		t.Errorf("expected ErrNestedVault, got: %v", err)
	}
}

func TestCheckNesting_SelfIsVault(t *testing.T) {
	dir := t.TempDir()
	// path itself has .vb/ — this is deferred to vault.Init's guard, not CheckNesting.
	if err := os.MkdirAll(filepath.Join(dir, ".vb"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := CheckNesting(dir); err != nil {
		t.Fatalf("CheckNesting() on self-vault should pass (deferred to Init): %v", err)
	}
}
