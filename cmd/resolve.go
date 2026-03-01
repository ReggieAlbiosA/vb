package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ReggieAlbiosA/vb/internal/registry"
	"github.com/ReggieAlbiosA/vb/internal/vault"
)

// flagVault is set by the --vault / -V persistent flag on rootCmd.
var flagVault string

// resolveVault applies the 3-priority vault resolution strategy:
//
//  1. If --vault / -V is set: look up the name in the global registry.
//  2. Walk upward from cwd looking for .vb/ (backward-compatible behaviour).
//  3. Fall back to the default vault in the global registry.
func resolveVault() (vault.Context, error) {
	// Priority 1: explicit --vault flag.
	if flagVault != "" {
		reg, err := registry.Load()
		if err != nil {
			return vault.Context{}, fmt.Errorf("loading registry: %w", err)
		}
		path, err := reg.Lookup(flagVault)
		if err != nil {
			return vault.Context{}, err
		}
		return vault.Resolve(path)
	}

	// Priority 2: cwd walk.
	cwd, err := os.Getwd()
	if err != nil {
		return vault.Context{}, fmt.Errorf("cannot determine current directory: %w", err)
	}
	ctx, cwdErr := vault.Resolve(cwd)
	if cwdErr == nil {
		return ctx, nil
	}
	if !errors.Is(cwdErr, vault.ErrNoVault) {
		return vault.Context{}, cwdErr
	}

	// Priority 3: default vault from registry.
	reg, err := registry.Load()
	if err != nil {
		return vault.Context{}, cwdErr
	}
	if reg.Default == "" {
		return vault.Context{},
			fmt.Errorf("%w\n\nTip: register a default vault with `vb vault create <name> <path>`"+
				"\n     or run `vb init` inside your knowledge base directory", vault.ErrNoVault)
	}
	defaultPath, err := reg.Lookup(reg.Default)
	if err != nil {
		return vault.Context{}, err
	}
	return vault.Resolve(defaultPath)
}
