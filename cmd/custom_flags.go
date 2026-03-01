package cmd

import (
	"os"

	"github.com/ReggieAlbiosA/vb/internal/config"
	"github.com/ReggieAlbiosA/vb/internal/registry"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

// registerCustomLenses loads the vault config (best-effort) and registers any
// custom lenses as boolean flags on the given command.
//
// Resolution: cwd walk first, then default registry vault as fallback.
// The --vault flag cannot be used here because Cobra's init() runs before
// flag parsing.
func registerCustomLenses(cmd *cobra.Command) {
	// Try cwd walk first (backward-compatible).
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	ctx, cwdErr := vault.Resolve(cwd)
	if cwdErr != nil {
		// Not inside a vault — try the default registry vault.
		reg, regErr := registry.Load()
		if regErr != nil || reg.Default == "" {
			return
		}
		path, lookErr := reg.Lookup(reg.Default)
		if lookErr != nil {
			return
		}
		ctx, cwdErr = vault.Resolve(path)
		if cwdErr != nil {
			return
		}
	}

	cfg, err := config.Load(ctx.VaultRoot)
	if err != nil {
		return
	}

	for flag, filename := range cfg.CustomLenses {
		// Register in the resolver so ResolveLens/ActiveLens see it.
		resolver.RegisterCustomLens(flag, filename)

		// Register the cobra flag if it doesn't already exist (avoid double-register).
		if cmd.Flags().Lookup(flag) == nil {
			cmd.Flags().Bool(flag, false, "custom lens: "+filename)
		}
	}
}
