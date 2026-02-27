package resolver

import (
	"fmt"
	"os"
	"path/filepath"
)

// Bind constructs the absolute file path and verifies it exists.
// Returns a user-readable error if the lens file hasn't been authored yet.
func Bind(topicDir string, lensFile string) (string, error) {
	fullPath := filepath.Join(topicDir, lensFile)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		topic := filepath.Base(topicDir)
		return "", fmt.Errorf("%w: %q not authored yet for topic %q — run: vb edit %s --%s",
			ErrLensFileMissing, lensFile, topic, topic, lensNameFor(lensFile))
	}
	return fullPath, nil
}

// lensNameFor reverse-looks up the flag name for a given vault filename.
func lensNameFor(filename string) string {
	for flag, file := range LensToFile {
		if file == filename {
			return flag
		}
	}
	return filename
}
