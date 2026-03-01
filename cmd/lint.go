package cmd

import (
	"fmt"

	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/linter"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	lintFlagWhy        bool
	lintFlagImportance bool
	lintFlagCLITools   bool
	lintFlagArch       bool
	lintFlagGotchas    bool
	lintFlagRefs       bool
)

var lintCmd = &cobra.Command{
	Use:   "lint <topic>",
	Short: "Validate a lens file against its schema",
	Args:  cobra.ExactArgs(1),
	RunE:  runLint,
}

func init() {
	lintCmd.Flags().BoolVar(&lintFlagWhy, "why", false, "why this topic exists in your stack")
	lintCmd.Flags().BoolVar(&lintFlagImportance, "importance", false, "importance and impact of this topic")
	lintCmd.Flags().BoolVar(&lintFlagCLITools, "cli-tools", false, "CLI tools for this topic")
	lintCmd.Flags().BoolVar(&lintFlagArch, "arch", false, "architecture overview")
	lintCmd.Flags().BoolVar(&lintFlagGotchas, "gotchas", false, "gotchas and pitfalls")
	lintCmd.Flags().BoolVar(&lintFlagRefs, "refs", false, "reference links")

	registerCustomLenses(lintCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	topic := args[0]

	lens, err := resolver.ActiveLens(cmd.Flags())
	if err != nil {
		return err
	}

	ctx, err := resolveVault()
	if err != nil {
		return err
	}

	schema, err := index.Load(ctx.VaultRoot)
	if err != nil {
		return err
	}

	topicDir, err := resolver.ResolveTopic(topic, schema, ctx.TopicRoot)
	if err != nil {
		return err
	}

	lensFile, err := resolver.ResolveLens(lens)
	if err != nil {
		return err
	}

	filePath, err := resolver.Bind(topicDir, lensFile)
	if err != nil {
		return err
	}

	lintErrs, err := linter.Lint(filePath, lens)
	if err != nil {
		return err
	}

	if len(lintErrs) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "✔ %s: schema valid\n", lens)
		return nil
	}

	for _, e := range lintErrs {
		fmt.Fprintf(cmd.ErrOrStderr(), "✘ [%s] %s\n", e.Lens, e.Message)
	}
	return fmt.Errorf("%d schema violation(s) found", len(lintErrs))
}
