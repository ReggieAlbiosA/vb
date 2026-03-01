package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vb",
	Short: "verbose — multi-lens terminal knowledge base",
	Long: `vb is a terminal-native knowledge base.
Store and query technical information through topic + lens combinations.

Built-in lenses: --why, --importance, --cli-tools, --arch, --gotchas, --refs
Create custom lenses: vb create <topic> -c "FILENAME.md"

Run 'vb init' in your knowledge base directory to create a vault.
Register vaults globally with 'vb vault create' to use vb from any directory.`,
}

// Execute is the entrypoint called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flag: available on rootCmd and all subcommands.
	rootCmd.PersistentFlags().StringVarP(
		&flagVault, "vault", "V", "",
		"target a specific vault by registry name (overrides cwd detection)",
	)

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(reindexCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(vaultCmd)
	rootCmd.AddCommand(topicCmd)
}
