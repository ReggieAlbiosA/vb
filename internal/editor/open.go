package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/hook"
)

// Open launches the configured editor on filePath.
// If filePath does not exist, it is created (empty) before opening.
// editor is read from vault config (cfg.Editor); falls back to $EDITOR env var if blank.
// After the editor exits successfully, hook.OnSave is called with lens and lintOnSave.
func Open(filePath string, editor string, lens string, lintOnSave bool) error {
	if err := ensureFile(filePath); err != nil {
		return fmt.Errorf("creating %s: %w", filePath, err)
	}

	ed := resolveEditor(editor)
	if ed == "" {
		return errors.New("no editor configured — set 'editor' in .vb/config.toml or $EDITOR")
	}

	cmd := exec.Command(ed, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	hook.OnSave(filePath, lens, lintOnSave)
	return nil
}

// ensureFile creates filePath (and parent dirs) if it does not already exist.
func ensureFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return nil // already exists
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte{}, 0o644)
}

// resolveEditor returns the editor to use: config value first, then $EDITOR env var.
func resolveEditor(cfgEditor string) string {
	if cfgEditor != "" {
		return cfgEditor
	}
	return os.Getenv("EDITOR")
}
