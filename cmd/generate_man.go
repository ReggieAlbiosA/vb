//go:build ignore

// generate_man.go — run via `go run cmd/generate_man.go`
// Called by the GoReleaser before hook to produce man pages.
package main

import (
	"os"

	"github.com/ReggieAlbiosA/vb/cmd"
)

func main() {
	os.MkdirAll("man", 0o755) //nolint:errcheck
	cmd.GenerateManPage("man")
}
