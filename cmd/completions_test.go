package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCompletions_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	GenerateCompletions(dir)

	expected := []string{"vb.bash", "vb.zsh", "vb.fish", "_vb"}
	for _, name := range expected {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected completion file %s to exist: %v", name, err)
		}
		if info.Size() == 0 {
			t.Errorf("completion file %s is empty", name)
		}
	}
}

func TestGenerateManPage_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	GenerateManPage(dir)

	// GenManTree generates at least vb.1 for the root command.
	path := filepath.Join(dir, "vb.1")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected man page vb.1 to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("man page vb.1 is empty")
	}
}
