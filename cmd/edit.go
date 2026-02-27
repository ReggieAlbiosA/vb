package cmd

import (
	"os"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/config"
	"github.com/ReggieAlbiosA/vb/internal/editor"
	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

var (
	editFlagWhy        bool
	editFlagImportance bool
	editFlagCLITools   bool
	editFlagArch       bool
	editFlagGotchas    bool
	editFlagRefs       bool
)

var editCmd = &cobra.Command{
	Use:   "edit <topic>",
	Short: "Open a lens file in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func init() {
	editCmd.Flags().BoolVar(&editFlagWhy, "why", false, "why this topic exists in your stack")
	editCmd.Flags().BoolVar(&editFlagImportance, "importance", false, "importance and impact of this topic")
	editCmd.Flags().BoolVar(&editFlagCLITools, "cli-tools", false, "CLI tools for this topic")
	editCmd.Flags().BoolVar(&editFlagArch, "arch", false, "architecture overview")
	editCmd.Flags().BoolVar(&editFlagGotchas, "gotchas", false, "gotchas and pitfalls")
	editCmd.Flags().BoolVar(&editFlagRefs, "refs", false, "reference links")
}

func runEdit(cmd *cobra.Command, args []string) error {
	topic := args[0]

	lens, err := resolver.ActiveLens(cmd.Flags())
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := vault.Resolve(cwd)
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

	// editor.Open creates the file if it doesn't exist — skip resolver.Bind().
	filePath := filepath.Join(topicDir, lensFile)

	cfg, err := config.Load(ctx.VaultRoot)
	if err != nil {
		return err
	}

	return editor.Open(filePath, cfg.Editor, lens, cfg.LintOnSave)
}
