package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

// Save appends a command entry to USED.md inside topicDir.
// File is created if it does not exist. Existing content is never overwritten.
func Save(topicDir, command, description string) error {
	usedPath := filepath.Join(topicDir, "USED.md")

	f, err := os.OpenFile(usedPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening USED.md: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "- %s — %s\n", command, description)
	return err
}
