package resolver_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
)

func TestResolveTopic_Found(t *testing.T) {
	schema := index.Schema{Topics: map[string]string{"disk": "hardware/disk"}}
	topicRoot := "/vault/topics"

	got, err := resolver.ResolveTopic("disk", schema, topicRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(topicRoot, "hardware/disk")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveTopic_NotFound(t *testing.T) {
	schema := index.Schema{Topics: map[string]string{"disk": "hardware/disk"}}

	_, err := resolver.ResolveTopic("unknown", schema, "/vault")
	if !errors.Is(err, resolver.ErrTopicNotFound) {
		t.Errorf("expected ErrTopicNotFound, got %v", err)
	}
}

func TestResolveTopic_EmptyIndex(t *testing.T) {
	schema := index.Schema{Topics: map[string]string{}}

	_, err := resolver.ResolveTopic("disk", schema, "/vault")
	if !errors.Is(err, resolver.ErrTopicNotFound) {
		t.Errorf("expected ErrTopicNotFound on empty index, got %v", err)
	}
}

func TestResolveTopic_UsesTopicRoot(t *testing.T) {
	schema := index.Schema{Topics: map[string]string{"disk": "hardware/disk"}}
	topicRoot := "/different/topics"

	got, err := resolver.ResolveTopic("disk", schema, topicRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(topicRoot, "hardware/disk")
	if got != want {
		t.Errorf("got %q, want %q — path must be rooted under topicRoot, not vaultRoot", got, want)
	}
}
