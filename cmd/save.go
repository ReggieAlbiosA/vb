package cmd

import (
	"fmt"

	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/logger"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/spf13/cobra"
)

var flagSaveDesc string

var saveCmd = &cobra.Command{
	Use:   `save <topic> "<command>"`,
	Short: "Save a command to a topic's USED.md",
	Long: `Save a shell command with a description to a topic's USED.md file.
Build a personal command cookbook for each topic.

Examples:
  vb save partition "lsblk" -d "show all block devices in tree view"
  vb save disk "sudo smartctl -H /dev/sda" -d "quick SMART health check"
  vb save partition..fs "df -hT" -d "show filesystem types and usage"

Read saved commands back with:
  vb partition --used`,
	Args: cobra.ExactArgs(2),
	RunE: runSave,
}

func init() {
	saveCmd.Flags().StringVarP(&flagSaveDesc, "desc", "d", "", "description of what this command does")
	saveCmd.MarkFlagRequired("desc") //nolint:errcheck
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	topic, command := args[0], args[1]

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

	if err := logger.Save(topicDir, command, flagSaveDesc); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "saved: %s — %s\n", command, flagSaveDesc)
	return nil
}
