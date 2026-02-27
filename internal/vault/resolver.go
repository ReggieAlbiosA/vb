package vault

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/config"
)

// Context is the result of a successful vault resolution.
// VaultRoot is the directory containing .vb/.
// TopicRoot is the resolved knowledge_path — where topic folders live.
// These may or may not be the same directory.
type Context struct {
	VaultRoot string
	TopicRoot string
}

// ErrNoVault is returned when no .vb/ marker is found walking upward.
var ErrNoVault = errors.New("no vault found — run `vb init` in your knowledge base directory")

// Resolve performs two-stage vault discovery starting from startDir.
//
// Stage 1: walk parent directories until .vb/ is found.
// Stage 2: read knowledge_path from config.toml; resolve it relative to VaultRoot.
//
// Stage 1 is pure filesystem stat — no config reads, no index access.
// Stage 2 reads config only — no filesystem walk of topic folders yet.
func Resolve(startDir string) (Context, error) {
	// Stage 1 — find the .vb/ marker.
	vaultRoot, err := findMarker(startDir)
	if err != nil {
		return Context{}, err
	}

	// Stage 2 — resolve knowledge_path from config.
	cfg, err := config.Load(vaultRoot)
	if err != nil {
		return Context{}, err
	}

	topicRoot := cfg.KnowledgePath
	if !filepath.IsAbs(topicRoot) {
		// Relative paths are resolved against the vault root.
		topicRoot = filepath.Join(vaultRoot, topicRoot)
	}
	topicRoot = filepath.Clean(topicRoot)

	return Context{
		VaultRoot: vaultRoot,
		TopicRoot: topicRoot,
	}, nil
}

// findMarker walks from dir upward until it finds a directory containing .vb/.
func findMarker(dir string) (string, error) {
	dir = filepath.Clean(dir)
	for {
		info, err := os.Stat(filepath.Join(dir, ".vb"))
		if err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return "", ErrNoVault
		}
		dir = parent
	}
}
