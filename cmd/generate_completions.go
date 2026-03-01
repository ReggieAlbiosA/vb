//go:build ignore

// generate_completions.go — run via `go run cmd/generate_completions.go`
// Called by the GoReleaser before hook to produce completion scripts.
package main

import (
	"os"

	"github.com/ReggieAlbiosA/vb/cmd"
)

func main() {
	os.MkdirAll("completions", 0o755) //nolint:errcheck
	cmd.GenerateCompletions("completions")
}
