package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/config"
	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/spf13/cobra"
)

var flagCreate string

var createCmd = &cobra.Command{
	Use:   "create <topic> -c <FILENAME.md>",
	Short: "Create a custom lens for a topic",
	Long: `Create a custom lens file and register it as a CLI flag.

Filename must be UPPERCASE with underscores (e.g. FILESYSTEM_FORMAT.md).
Only .md and .mmd extensions are accepted.

After creation, use the new flag with any command:
  vb <topic> --filesystem-format
  vb edit <topic> --filesystem-format`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&flagCreate, "create", "c", "", "lens filename to create (e.g. FILESYSTEM_FORMAT.md)")
	createCmd.MarkFlagRequired("create") //nolint:errcheck
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	topic := args[0]
	filename := flagCreate

	// 1. Validate the filename.
	if err := resolver.ValidateLensFilename(filename); err != nil {
		return err
	}

	// 2. Derive the flag name.
	flag := resolver.FilenameToFlag(filename)

	// 3. Check for collision with built-in or existing custom flags.
	if resolver.IsReservedFlag(flag) {
		return fmt.Errorf("--%s is already a built-in flag — cannot create duplicate", flag)
	}

	// 4. Resolve vault + topic.
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

	// 5. Check if custom lens already exists in config.
	cfg, err := config.Load(ctx.VaultRoot)
	if err != nil {
		return err
	}
	if _, exists := cfg.CustomLenses[flag]; exists {
		return fmt.Errorf("--%s is already a custom lens — use 'vb edit %s --%s' to edit it", flag, topic, flag)
	}

	// 6. Create the file (empty scaffold).
	filePath := filepath.Join(topicDir, filename)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("%s already exists at %s", filename, topicDir)
	}
	if err := os.WriteFile(filePath, []byte(""), 0o644); err != nil {
		return fmt.Errorf("creating %s: %w", filename, err)
	}

	// 7. Save the custom lens to vault config.
	if err := config.SaveCustomLens(ctx.VaultRoot, flag, filename); err != nil {
		return fmt.Errorf("saving custom lens to config: %w", err)
	}

	// 8. Register in resolver so it's immediately usable.
	resolver.RegisterCustomLens(flag, filename)

	fmt.Printf("✔ lens --%s created → %s/%s\n", flag, topic, filename)
	return nil
}
