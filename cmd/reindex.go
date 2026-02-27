package cmd

import (
	"fmt"
	"os"

	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Rebuild the topic index from your knowledge_path",
	Long: `Walks the knowledge_path defined in .vb/config.toml and rebuilds
.vb/index.json with all detected topic folders.

A topic folder is any directory that directly contains at least one .md file.
Run this after adding, renaming, or removing topic folders.`,
	RunE: runReindex,
}

func runReindex(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	// Two-stage vault resolution.
	ctx, err := vault.Resolve(cwd)
	if err != nil {
		return err
	}

	fmt.Printf("vault root : %s\n", ctx.VaultRoot)
	fmt.Printf("topic root : %s\n", ctx.TopicRoot)

	schema, err := index.Build(ctx.VaultRoot, ctx.TopicRoot)
	if err != nil {
		return err
	}

	count := len(schema.Topics)
	if count == 0 {
		fmt.Println("✓ Index rebuilt — no topics found yet (add topic folders with .md files)")
	} else {
		fmt.Printf("✓ Index rebuilt — %d topic(s) indexed\n", count)
		for name, path := range schema.Topics {
			fmt.Printf("  %-20s %s\n", name, path)
		}
	}
	return nil
}
