package index_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/index"
)

// mkVaultDir creates a minimal vault structure (just a .vb/ dir) for index tests.
func mkVaultDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".vb"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeFile creates a file at path, creating parent dirs as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestBuild_Empty: no topic folders → empty index.
func TestBuild_Empty(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	schema, err := index.Build(vaultRoot, vaultRoot)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if len(schema.Topics) != 0 {
		t.Errorf("Topics: got %d entries, want 0", len(schema.Topics))
	}
}

// TestBuild_WithTopics: directories with .md files are indexed.
func TestBuild_WithTopics(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	writeFile(t, filepath.Join(vaultRoot, "hardware", "disk", "WHY.md"), "# Disk")
	writeFile(t, filepath.Join(vaultRoot, "networking", "ssh", "WHY.md"), "# SSH")

	schema, err := index.Build(vaultRoot, vaultRoot)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if len(schema.Topics) != 2 {
		t.Errorf("Topics: got %d, want 2. Topics: %v", len(schema.Topics), schema.Topics)
	}
	if schema.Topics["disk"] != "hardware/disk" {
		t.Errorf("disk path: got %q, want %q", schema.Topics["disk"], "hardware/disk")
	}
	if schema.Topics["ssh"] != "networking/ssh" {
		t.Errorf("ssh path: got %q, want %q", schema.Topics["ssh"], "networking/ssh")
	}
}

// TestBuild_SkipsVbDir: .vb/ directory is never indexed regardless of contents.
func TestBuild_SkipsVbDir(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	writeFile(t, filepath.Join(vaultRoot, ".vb", "README.md"), "# internal")
	writeFile(t, filepath.Join(vaultRoot, "hardware", "disk", "WHY.md"), "# Disk")

	schema, err := index.Build(vaultRoot, vaultRoot)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if _, ok := schema.Topics[".vb"]; ok {
		t.Error("Topics: .vb/ was indexed — it must always be skipped")
	}
	if len(schema.Topics) != 1 {
		t.Errorf("Topics: got %d, want 1", len(schema.Topics))
	}
}

// TestBuild_IgnoresDirsWithoutMD: category dirs (no direct .md files) are not indexed.
func TestBuild_IgnoresDirsWithoutMD(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	writeFile(t, filepath.Join(vaultRoot, "hardware", "disk", "WHY.md"), "# Disk")

	schema, err := index.Build(vaultRoot, vaultRoot)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if _, ok := schema.Topics["hardware"]; ok {
		t.Error("Topics: 'hardware' category dir was indexed — only leaf topic dirs should be")
	}
	if _, ok := schema.Topics["disk"]; !ok {
		t.Error("Topics: 'disk' was not indexed")
	}
}

// TestBuild_DecoupledTopicRoot: index built from a different topicRoot than vaultRoot.
func TestBuild_DecoupledTopicRoot(t *testing.T) {
	vaultRoot := mkVaultDir(t)
	topicRoot := t.TempDir()

	writeFile(t, filepath.Join(topicRoot, "devops", "docker", "WHY.md"), "# Docker")

	schema, err := index.Build(vaultRoot, topicRoot)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if schema.Topics["docker"] != "devops/docker" {
		t.Errorf("docker path: got %q, want %q", schema.Topics["docker"], "devops/docker")
	}
}

// TestBuild_WritesIndexJSON: index.json is written to vaultRoot/.vb/.
func TestBuild_WritesIndexJSON(t *testing.T) {
	vaultRoot := mkVaultDir(t)
	writeFile(t, filepath.Join(vaultRoot, "hardware", "disk", "WHY.md"), "# Disk")

	if _, err := index.Build(vaultRoot, vaultRoot); err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	indexPath := filepath.Join(vaultRoot, ".vb", "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("index.json not written to %s", indexPath)
	}
}

// TestBuild_NonExistentTopicRoot: returns error when topicRoot doesn't exist.
func TestBuild_NonExistentTopicRoot(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	_, err := index.Build(vaultRoot, "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("Build() expected error for non-existent topicRoot, got nil")
	}
}

// TestBuild_WriteError: returns error when index.json cannot be written.
func TestBuild_WriteError(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	// Make .vb/index.json a directory — WriteFile will fail trying to write to it.
	indexAsDir := filepath.Join(vaultRoot, ".vb", "index.json")
	if err := os.MkdirAll(indexAsDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := index.Build(vaultRoot, vaultRoot)
	if err == nil {
		t.Fatal("Build() expected error when index.json is a directory, got nil")
	}
}

// TestLoad_RoundTrip: Build then Load returns the same schema.
func TestLoad_RoundTrip(t *testing.T) {
	vaultRoot := mkVaultDir(t)
	writeFile(t, filepath.Join(vaultRoot, "hardware", "disk", "WHY.md"), "# Disk")

	built, err := index.Build(vaultRoot, vaultRoot)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	loaded, err := index.Load(vaultRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Topics["disk"] != built.Topics["disk"] {
		t.Errorf("round-trip mismatch: built %v, loaded %v", built.Topics, loaded.Topics)
	}
}

// TestLoad_FileNotFound: returns error when index.json does not exist.
func TestLoad_FileNotFound(t *testing.T) {
	vaultRoot := mkVaultDir(t)
	// No index.json written — Load should return an error.

	_, err := index.Load(vaultRoot)
	if err == nil {
		t.Fatal("Load() expected error for missing index.json, got nil")
	}
}

// TestLoad_MalformedJSON: returns error when index.json contains invalid JSON.
func TestLoad_MalformedJSON(t *testing.T) {
	vaultRoot := mkVaultDir(t)

	writeFile(t, filepath.Join(vaultRoot, ".vb", "index.json"), "{ this is not valid json }")

	_, err := index.Load(vaultRoot)
	if err == nil {
		t.Fatal("Load() expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing index.json") {
		t.Errorf("Load() error message: got %q, want it to contain 'parsing index.json'", err.Error())
	}
}
