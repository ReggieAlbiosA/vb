package resolver

import (
	"fmt"
	"path/filepath"

	"github.com/ReggieAlbiosA/vb/internal/index"
)

// ResolveTopic looks up topic in the loaded index and returns
// the absolute path to the topic directory.
//
// Note: paths in index.json are relative to TopicRoot (not VaultRoot).
// TopicRoot comes from VaultContext resolved in Phase 01.
func ResolveTopic(topic string, schema index.Schema, topicRoot string) (string, error) {
	relPath, exists := schema.Topics[topic]
	if !exists {
		return "", fmt.Errorf("%w: %q", ErrTopicNotFound, topic)
	}
	return filepath.Join(topicRoot, relPath), nil // ← TopicRoot, NOT VaultRoot
}
