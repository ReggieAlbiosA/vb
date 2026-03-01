package cmd

import (
	"path/filepath"
)

// GenerateCompletions writes shell completion scripts to the given directory.
// Called by the generate_completions.go build script during release.
func GenerateCompletions(dir string) {
	rootCmd.GenBashCompletionFile(filepath.Join(dir, "vb.bash"))        //nolint:errcheck
	rootCmd.GenZshCompletionFile(filepath.Join(dir, "vb.zsh"))          //nolint:errcheck
	rootCmd.GenFishCompletionFile(filepath.Join(dir, "vb.fish"), false) //nolint:errcheck
	rootCmd.GenPowerShellCompletionFile(filepath.Join(dir, "_vb"))      //nolint:errcheck
}
