package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupVaultForTopic creates an initialized vault at dir and returns dir.
func setupVaultForTopic(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out, err := execCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("vb init error: %v\noutput: %s", err, out)
	}
	return dir
}

// resetTopicFlags resets flag state that bleeds between Cobra test runs.
func resetTopicFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		flagTopicIn = ""
		flagTopicTree = false
	})
}

// TestTopicCreate_Flat: creates a flat topic with 6 lens files.
func TestTopicCreate_Flat(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	out, err := execCmd(t, dir, "topic", "create", "disk")
	if err != nil {
		t.Fatalf("topic create error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, `"disk"`) {
		t.Errorf("output missing topic name: %s", out)
	}

	// Verify 6 lens files exist.
	topicDir := filepath.Join(dir, "disk")
	for _, f := range scaffoldFiles {
		fp := filepath.Join(topicDir, f)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
	}
}

// TestTopicCreate_Nested: creates a topic nested under an existing parent.
func TestTopicCreate_Nested(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	// First create the parent topic.
	if _, err := execCmd(t, dir, "topic", "create", "partition"); err != nil {
		t.Fatalf("create parent: %v", err)
	}

	// Create nested topic.
	out, err := execCmd(t, dir, "topic", "create", "fs", "--in", "partition")
	if err != nil {
		t.Fatalf("topic create nested error: %v\noutput: %s", err, out)
	}

	// Verify nested topic has lens files.
	nested := filepath.Join(dir, "partition", "fs")
	for _, f := range scaffoldFiles {
		if _, err := os.Stat(filepath.Join(nested, f)); os.IsNotExist(err) {
			t.Errorf("expected %s in nested topic", f)
		}
	}
}

// TestTopicCreate_DeepNested: creates a deeply nested topic using .. addressing.
func TestTopicCreate_DeepNested(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	// Create parent chain.
	if _, err := execCmd(t, dir, "topic", "create", "partition"); err != nil {
		t.Fatalf("create partition: %v", err)
	}
	if _, err := execCmd(t, dir, "topic", "create", "fs", "--in", "partition"); err != nil {
		t.Fatalf("create fs: %v", err)
	}

	// Create deep nested topic using ..-joined parent.
	out, err := execCmd(t, dir, "topic", "create", "mnt", "--in", "partition..fs")
	if err != nil {
		t.Fatalf("deep nested error: %v\noutput: %s", err, out)
	}

	deep := filepath.Join(dir, "partition", "fs", "mnt")
	for _, f := range scaffoldFiles {
		if _, err := os.Stat(filepath.Join(deep, f)); os.IsNotExist(err) {
			t.Errorf("expected %s in deep nested topic", f)
		}
	}
}

// TestTopicCreate_AlreadyExists: errors when topic already has .md files.
func TestTopicCreate_AlreadyExists(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	// Create topic.
	if _, err := execCmd(t, dir, "topic", "create", "disk"); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Second create should fail.
	_, err := execCmd(t, dir, "topic", "create", "disk")
	if err == nil {
		t.Fatal("expected error for duplicate topic, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists': %v", err)
	}
}

// TestTopicCreate_ParentNotFound: errors when --in references unknown topic.
func TestTopicCreate_ParentNotFound(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	_, err := execCmd(t, dir, "topic", "create", "fs", "--in", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown parent, got nil")
	}
}

// TestTopicList_Empty: no topics prints guidance message.
func TestTopicList_Empty(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	out, err := execCmd(t, dir, "topic", "list")
	if err != nil {
		t.Fatalf("topic list error: %v", err)
	}
	if !strings.Contains(out, "no topics") {
		t.Errorf("expected 'no topics' message, got: %s", out)
	}
}

// TestTopicList_Flat: lists topics in flat sorted format.
func TestTopicList_Flat(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	// Create a couple of topics.
	if _, err := execCmd(t, dir, "topic", "create", "disk"); err != nil {
		t.Fatalf("create disk: %v", err)
	}
	if _, err := execCmd(t, dir, "topic", "create", "ssh"); err != nil {
		t.Fatalf("create ssh: %v", err)
	}

	out, err := execCmd(t, dir, "topic", "list")
	if err != nil {
		t.Fatalf("topic list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "disk") {
		t.Errorf("output missing 'disk': %s", out)
	}
	if !strings.Contains(out, "ssh") {
		t.Errorf("output missing 'ssh': %s", out)
	}
}

// TestTopicList_Tree: nested topics displayed as indented tree.
func TestTopicList_Tree(t *testing.T) {
	isolateRegistry(t)
	resetTopicFlags(t)
	dir := setupVaultForTopic(t)

	// Create nested structure: partition > fs.
	if _, err := execCmd(t, dir, "topic", "create", "partition"); err != nil {
		t.Fatalf("create partition: %v", err)
	}
	if _, err := execCmd(t, dir, "topic", "create", "fs", "--in", "partition"); err != nil {
		t.Fatalf("create fs: %v", err)
	}

	out, err := execCmd(t, dir, "topic", "list", "--tree")
	if err != nil {
		t.Fatalf("topic list --tree error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "partition") {
		t.Errorf("tree output missing 'partition': %s", out)
	}
	if !strings.Contains(out, "  fs") {
		t.Errorf("tree output missing indented 'fs': %s", out)
	}
}
