package cmd

import (
	"fmt"
	"os"

	"github.com/ReggieAlbiosA/vb/internal/config"
	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/render"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/ReggieAlbiosA/vb/internal/vault"
	"github.com/spf13/cobra"
)

var (
	flagWhy        bool
	flagImportance bool
	flagCLITools   bool
	flagArch       bool
	flagUsed       bool
	flagGotchas    bool
	flagRefs       bool
	flagGUI        bool
)

func init() {
	// Wire rootCmd to handle topic queries when no subcommand matches.
	// cobra.ArbitraryArgs bypasses legacyArgs in Find(), which otherwise
	// generates "unknown command" for root commands with subcommands.
	rootCmd.Args = cobra.ArbitraryArgs
	rootCmd.RunE = runQuery

	rootCmd.Flags().BoolVar(&flagWhy, "why", false, "why this topic exists in your stack")
	rootCmd.Flags().BoolVar(&flagImportance, "importance", false, "importance and impact of this topic")
	rootCmd.Flags().BoolVar(&flagCLITools, "cli-tools", false, "CLI tools for this topic")
	rootCmd.Flags().BoolVar(&flagArch, "arch", false, "architecture overview")
	rootCmd.Flags().BoolVar(&flagUsed, "used", false, "how this topic is used")
	rootCmd.Flags().BoolVar(&flagGotchas, "gotchas", false, "gotchas and pitfalls")
	rootCmd.Flags().BoolVar(&flagRefs, "refs", false, "reference links")
	rootCmd.Flags().BoolVar(&flagGUI, "gui", false, "open in GUI viewer (modifier, not a lens)")
}

func runQuery(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	if len(args) > 1 {
		return fmt.Errorf("expected 1 topic argument, got %d", len(args))
	}
	topic := args[0]

	// Determine which lens flag was set.
	lens, err := resolver.ActiveLens(cmd.Flags())
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	// Two-stage vault resolution (Phase 01).
	ctx, err := vault.Resolve(cwd)
	if err != nil {
		return err
	}

	// Load config for theme (Phase 03).
	cfg, err := config.Load(ctx.VaultRoot)
	if err != nil {
		return err
	}

	// Load index from existing package.
	schema, err := index.Load(ctx.VaultRoot)
	if err != nil {
		return err
	}

	// Resolve topic → topicDir (using TopicRoot, not VaultRoot).
	topicDir, err := resolver.ResolveTopic(topic, schema, ctx.TopicRoot)
	if err != nil {
		return err
	}

	// Resolve lens flag → filename.
	lensFile, err := resolver.ResolveLens(lens)
	if err != nil {
		return err
	}

	// Validate file exists.
	filePath, err := resolver.Bind(topicDir, lensFile)
	if err != nil {
		return err
	}

	// Hand off to renderer (Phase 03), passing lens, flagGUI, and theme.
	return render.File(filePath, lens, flagGUI, cfg.Theme)
}
