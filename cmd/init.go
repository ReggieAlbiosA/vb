package cmd

import (
	"fmt"
	"os"

	"github.com/ReggieAlbiosA/vb/internal/registry"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

var flagInitName string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a vault in the current directory",
	Long: `Creates a .vb/ directory in the current working directory.

This marks the current directory as a vault root — identical to how
'git init' marks a directory as a git repository. The .vb/ folder
contains config.toml and index.json.

The vault (topic folders like hardware/disk/) can live here or
anywhere else — set knowledge_path in .vb/config.toml.

Optionally provide --name to also register the vault in the global registry:
  vb init --name sysknow`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&flagInitName, "name", "n", "",
		"register this vault in the global registry under the given name")
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	if err := vault.Init(cwd); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Vault initialized at %s\n", cwd)
	fmt.Fprintln(cmd.OutOrStdout(), "  .vb/config.toml  — edit to set knowledge_path, editor, theme")
	fmt.Fprintln(cmd.OutOrStdout(), "  .vb/index.json   — auto-managed, run `vb reindex` to rebuild")

	if flagInitName != "" {
		reg, err := registry.Load()
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}
		if err := reg.Add(flagInitName, cwd); err != nil {
			return err
		}
		if len(reg.Vaults) == 1 {
			_ = reg.SetDefault(flagInitName)
		}
		if err := reg.Save(); err != nil {
			return fmt.Errorf("saving registry: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  registered as %q in global vault registry\n", flagInitName)
	}

	return nil
}
