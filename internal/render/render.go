package render

import (
	"fmt"
	"os"
)

// File outputs the content of a file to stdout.
// gui is reserved for Phase 03 — the full rendering layer.
func File(path string, gui bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	fmt.Print(string(data))
	return nil
}
