package cmd

import (
	"fmt"
	"os"

	"github.com/ReggieAlbiosA/vb/internal/tagger"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag <name>",
	Short: "Search for a tag across all topics",
	Args:  cobra.ExactArgs(1),
	RunE:  runTag,
}

func runTag(cmd *cobra.Command, args []string) error {
	tagName := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := vault.Resolve(cwd)
	if err != nil {
		return err
	}

	results, err := tagger.Search(ctx.TopicRoot, tagName)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no topics tagged #%s\n", tagName)
		return nil
	}

	for _, r := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s  (%s)\n", r.Topic, r.File)
	}
	return nil
}
