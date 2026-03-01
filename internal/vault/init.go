package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/config"
)

// Init creates the .vb/ scaffold inside path (config.toml + index.json).
// Returns an error if path already contains a .vb/ directory.
func Init(path string) error {
	vbDir := filepath.Join(path, ".vb")

	if _, err := os.Stat(vbDir); err == nil {
		return fmt.Errorf("vault already initialized at %s", path)
	}

	if err := os.MkdirAll(vbDir, 0755); err != nil {
		return fmt.Errorf("creating .vb/: %w", err)
	}

	configPath := filepath.Join(vbDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(config.DefaultTOML()), 0644); err != nil {
		return fmt.Errorf("writing config.toml: %w", err)
	}

	emptyIndex := struct {
		Topics map[string]string `json:"topics"`
	}{Topics: map[string]string{}}
	indexData, _ := json.MarshalIndent(emptyIndex, "", "  ")
	indexPath := filepath.Join(vbDir, "index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("writing index.json: %w", err)
	}

	return nil
}
