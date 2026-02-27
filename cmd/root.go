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

Run 'vb init' in your knowledge base directory to create a vault.`,
}

// Execute is the entrypoint called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(reindexCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(lintCmd)
}
