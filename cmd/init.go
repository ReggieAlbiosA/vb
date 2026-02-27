package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a vault in the current directory",
	Long: `Creates a .vb/ directory in the current working directory.

This marks the current directory as a vault root — identical to how
'git init' marks a directory as a git repository. The .vb/ folder
contains config.toml and index.json.

The vault (topic folders like hardware/disk/) can live here or
anywhere else — set knowledge_path in .vb/config.toml.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	vbDir := filepath.Join(cwd, ".vb")

	// Guard: already initialized.
	if _, err := os.Stat(vbDir); err == nil {
		return fmt.Errorf("vault already initialized in this directory (%s)", cwd)
	}

	// Create .vb/
	if err := os.MkdirAll(vbDir, 0755); err != nil {
		return fmt.Errorf("creating .vb/: %w", err)
	}

	// Write .vb/config.toml
	configPath := filepath.Join(vbDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(config.DefaultTOML()), 0644); err != nil {
		return fmt.Errorf("writing config.toml: %w", err)
	}

	// Write .vb/index.json (empty scaffold)
	emptyIndex := struct {
		Topics map[string]string `json:"topics"`
	}{Topics: map[string]string{}}
	indexData, _ := json.MarshalIndent(emptyIndex, "", "  ")
	indexPath := filepath.Join(vbDir, "index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("writing index.json: %w", err)
	}

	fmt.Printf("✓ Vault initialized at %s\n", cwd)
	fmt.Println("  .vb/config.toml  — edit to set knowledge_path, editor, theme")
	fmt.Println("  .vb/index.json   — auto-managed, run `vb reindex` to rebuild")
	return nil
}
