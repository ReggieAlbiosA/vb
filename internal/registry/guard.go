package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// ErrNestedVault is returned when creating a vault would nest inside or around
// an existing vault.
var ErrNestedVault = fmt.Errorf("nested vaults are not supported")

// CheckNesting verifies that path does not overlap with any existing vault.
//
// 1. Ancestor check: walks from parent(path) upward looking for .vb/.
// 2. Descendant check: walks path's subtree looking for .vb/.
//
// A non-existent path passes both checks.
func CheckNesting(path string) error {
	path = filepath.Clean(path)

	// Ancestor check: walk from parent upward.
	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".vb")); err == nil {
			return fmt.Errorf("%w: %q is inside existing vault at %q",
				ErrNestedVault, path, dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Descendant check: walk path looking for .vb/ in subdirectories.
	var nestErr error
	_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipAll
		}
		if !d.IsDir() || d.Name() != ".vb" {
			return nil
		}
		vaultDir := filepath.Dir(p)
		if vaultDir == path {
			// path itself has .vb/ — deferred to vault.Init's guard.
			return nil
		}
		nestErr = fmt.Errorf("%w: %q contains existing vault at %q",
			ErrNestedVault, path, vaultDir)
		return nestErr
	})

	return nestErr
}
