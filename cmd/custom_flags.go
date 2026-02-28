package cmd

import (
	"os"

	"github.com/ReggieAlbiosA/vb/internal/config"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

// registerCustomLenses loads the vault config from cwd (best-effort) and
// registers any custom lenses as boolean flags on the given command.
// If not inside a vault, this silently does nothing.
func registerCustomLenses(cmd *cobra.Command) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	ctx, err := vault.Resolve(cwd)
	if err != nil {
		return // not inside a vault — skip
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
