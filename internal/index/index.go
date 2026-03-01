package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Schema is the structure of .vb/index.json.
type Schema struct {
	Topics map[string]string `json:"topics"`
}

// Empty returns an empty index scaffold — written by vb init.
func Empty() Schema {
	return Schema{Topics: map[string]string{}}
}

// Build walks topicRoot, detects topic folders, and writes index.json to vaultRoot/.vb/.
//
// A topic folder is a directory that directly contains at least one .md file.
// All path values in the index are relative to topicRoot, keeping the index
// portable even if topicRoot is moved.
func Build(vaultRoot, topicRoot string) (Schema, error) {
	schema := Empty()

	err := filepath.WalkDir(topicRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		// Never descend into .vb/ — regardless of where it sits.
		if d.Name() == ".vb" {
			return filepath.SkipDir
		}

		// Check if this directory directly contains any .md files.
		if hasMDFiles(path) {
			rel, relErr := filepath.Rel(topicRoot, path)
			if relErr != nil {
				return relErr
			}
			// Use forward slashes in the index for cross-platform consistency.
			rel = filepath.ToSlash(rel)
			// Topic name is the leaf directory name.
			topicName := d.Name()
			schema.Topics[topicName] = rel

			// Add ..-joined key for explicit nested addressing.
			dotDotKey := strings.ReplaceAll(rel, "/", "..")
			if dotDotKey != topicName {
				schema.Topics[dotDotKey] = rel
			}
		}
		return nil
	})
	if err != nil {
		return Schema{}, fmt.Errorf("walking topic root: %w", err)
	}

	// Write to VaultRoot/.vb/index.json.
	indexPath := filepath.Join(vaultRoot, ".vb", "index.json")
	if err := write(indexPath, schema); err != nil {
		return Schema{}, err
	}
	return schema, nil
}

// Load reads and parses .vb/index.json from vaultRoot.
func Load(vaultRoot string) (Schema, error) {
	data, err := os.ReadFile(filepath.Join(vaultRoot, ".vb", "index.json"))
	if err != nil {
		return Schema{}, fmt.Errorf("reading index.json: %w", err)
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return Schema{}, fmt.Errorf("parsing index.json: %w", err)
	}
	return s, nil
}

// write serialises schema to a JSON file with consistent formatting.
func write(path string, s Schema) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising index: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing index.json: %w", err)
	}
	return nil
}

// hasMDFiles reports whether dir directly contains at least one .md file.
func hasMDFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".md" {
			return true
		}
	}
	return false
}
