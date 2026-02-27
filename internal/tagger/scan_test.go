package tagger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/tagger"
)

// mkTopic creates topicRoot/topicName/filename with content.
func mkTopic(t *testing.T, topicRoot, topicName, filename, content string) {
	t.Helper()
	dir := filepath.Join(topicRoot, topicName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSearch_Found(t *testing.T) {
	topicRoot := t.TempDir()
	mkTopic(t, topicRoot, "disk", "WHY.md", "# Why\n\nUse this with #ssh for remote disks.")

	results, err := tagger.Search(topicRoot, "ssh")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Topic != "disk" {
		t.Errorf("expected topic 'disk', got %q", results[0].Topic)
	}
	if results[0].File != "WHY.md" {
		t.Errorf("expected file 'WHY.md', got %q", results[0].File)
	}
}

func TestSearch_NotFound(t *testing.T) {
	topicRoot := t.TempDir()
	mkTopic(t, topicRoot, "disk", "WHY.md", "# Why\n\nNo tags here.")

	results, err := tagger.Search(topicRoot, "ssh")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	topicRoot := t.TempDir()
	mkTopic(t, topicRoot, "networking", "WHY.md", "# Why\n\nConnects via #SSH protocol.")

	// Search lowercase, content has uppercase.
	results, err := tagger.Search(topicRoot, "ssh")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (case-insensitive), got %d", len(results))
	}
}

func TestSearch_WordBoundary(t *testing.T) {
	topicRoot := t.TempDir()
	// #disks should NOT match a search for #disk.
	mkTopic(t, topicRoot, "storage", "WHY.md", "# Why\n\nManages #disks and partitions.")

	results, err := tagger.Search(topicRoot, "disk")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results (#disk should not match #disks), got %d", len(results))
	}
}

func TestSearch_MultipleTopics(t *testing.T) {
	topicRoot := t.TempDir()
	mkTopic(t, topicRoot, "disk", "WHY.md", "# Disk\n\nRelated to #networking.")
	mkTopic(t, topicRoot, "ssh", "WHY.md", "# SSH\n\nAlso uses #networking.")

	results, err := tagger.Search(topicRoot, "networking")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_SkipsNonMD(t *testing.T) {
	topicRoot := t.TempDir()
	dir := filepath.Join(topicRoot, "disk")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// txt and toml files with the tag — must be skipped.
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("#ssh reference"), 0o644)
	os.WriteFile(filepath.Join(dir, "config.toml"), []byte("# #ssh"), 0o644)

	results, err := tagger.Search(topicRoot, "ssh")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results (non-.md files skipped), got %d", len(results))
	}
}

func TestSearch_ReadError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root — cannot test permission error")
	}
	topicRoot := t.TempDir()
	dir := filepath.Join(topicRoot, "disk")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(path, []byte("#ssh"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(path, 0o644) }) //nolint:errcheck

	_, err := tagger.Search(topicRoot, "ssh")
	if err == nil {
		t.Fatal("expected error for unreadable .md file, got nil")
	}
}

func TestSearch_SkipsVbDir(t *testing.T) {
	// topicRoot is a sibling of .vb/, so walking topicRoot never visits .vb/.
	vaultRoot := t.TempDir()
	topicRoot := filepath.Join(vaultRoot, "topics")
	vbDir := filepath.Join(vaultRoot, ".vb")

	if err := os.MkdirAll(topicRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	// Put a tagged file inside .vb/ — should never appear in results.
	if err := os.MkdirAll(vbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(vbDir, "internal.md"), []byte("#ssh internal"), 0o644)

	// Also a legitimate topic.
	mkTopic(t, topicRoot, "disk", "WHY.md", "# Disk\n\nUses #ssh.")

	results, err := tagger.Search(topicRoot, "ssh")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (from topics only), got %d: %+v", len(results), results)
	}
	if results[0].Topic != "disk" {
		t.Errorf("expected topic 'disk', got %q", results[0].Topic)
	}
}
