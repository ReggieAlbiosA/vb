package cmd

import (
	"github.com/spf13/cobra/doc"
)

// GenerateManPage writes man pages for all commands to the given directory.
// Called by the generate_man.go build script during release.
func GenerateManPage(dir string) {
	header := &doc.GenManHeader{
		Title:   "VB",
		Section: "1",
		Source:  "vb",
		Manual:  "vb Manual",
	}
	doc.GenManTree(rootCmd, header, dir) //nolint:errcheck
}
