package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ReggieAlbiosA/vb/internal/registry"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage registered vaults",
	Long: `Register, list, and switch between vaults in the global registry.

The global registry lives at ~/.config/vb/vaults.json (or $XDG_CONFIG_HOME/vb/vaults.json).
The default vault is used when vb is run outside any vault directory tree.`,
}

// ── vb vault create <name> <path> ────────────────────────────────────────────

var vaultCreateCmd = &cobra.Command{
	Use:   "create <name> <path>",
	Short: "Create a new vault at path and register it",
	Args:  cobra.ExactArgs(2),
	RunE:  runVaultCreate,
}

func runVaultCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	rawPath := args[1]

	absPath, err := filepath.Abs(rawPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Nesting guard — must run before MkdirAll.
	if err := registry.CheckNesting(absPath); err != nil {
		return err
	}

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Initialize .vb/ scaffold — skip if already a vault (just register it).
	alreadyVault := false
	if err := vault.Init(absPath); err != nil {
		// Check if it's the "already initialized" case — that's fine, just register.
		if _, statErr := os.Stat(filepath.Join(absPath, ".vb")); statErr == nil {
			alreadyVault = true
		} else {
			return err
		}
	}

	// Register in global registry.
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	if err := reg.Add(name, absPath); err != nil {
		return err
	}
	// Auto-set as default if this is the first vault.
	if len(reg.Vaults) == 1 {
		_ = reg.SetDefault(name)
	}
	if err := reg.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}

	if alreadyVault {
		fmt.Fprintf(cmd.OutOrStdout(), "✔ existing vault %q registered from %s\n", name, absPath)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "✔ vault %q created at %s\n", name, absPath)
	}
	if reg.Default == name {
		fmt.Fprintln(cmd.OutOrStdout(), "  (set as default)")
	}
	return nil
}

// ── vb vault list ─────────────────────────────────────────────────────────────

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered vaults",
	Args:  cobra.NoArgs,
	RunE:  runVaultList,
}

func runVaultList(cmd *cobra.Command, args []string) error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	if len(reg.Vaults) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no vaults registered — use `vb vault create` to add one")
		return nil
	}
	names := make([]string, 0, len(reg.Vaults))
	for name := range reg.Vaults {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		marker := "  "
		if name == reg.Default {
			marker = "* "
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s%-20s %s\n", marker, name, reg.Vaults[name])
	}
	return nil
}

// ── vb vault use <name> ──────────────────────────────────────────────────────

var vaultUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default vault",
	Args:  cobra.ExactArgs(1),
	RunE:  runVaultUse,
}

func runVaultUse(cmd *cobra.Command, args []string) error {
	name := args[0]
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	if err := reg.SetDefault(name); err != nil {
		return err
	}
	if err := reg.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✔ default vault set to %q\n", name)
	return nil
}

// ── vb vault remove <name> ───────────────────────────────────────────────────

var vaultRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Unregister a vault (files are not deleted)",
	Args:  cobra.ExactArgs(1),
	RunE:  runVaultRemove,
}

func runVaultRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	if err := reg.Remove(name); err != nil {
		return err
	}
	if err := reg.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✔ vault %q removed from registry (files unchanged)\n", name)
	if reg.Default == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "  (no default vault set — run `vb vault use <name>` to set one)")
	}
	return nil
}

func init() {
	vaultCmd.AddCommand(vaultCreateCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultUseCmd)
	vaultCmd.AddCommand(vaultRemoveCmd)
}
