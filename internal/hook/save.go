package hook

import (
	"fmt"
	"os"

	"github.com/ReggieAlbiosA/vb/internal/linter"
)

// OnSave is called by editor.Open() after the editor process exits.
// It runs lint only if lintOnSave is true.
// Lint errors are printed as warnings — they do not fail the edit command.
func OnSave(filePath, lens string, lintOnSave bool) {
	if !lintOnSave {
		return
	}

	lintErrs, err := linter.Lint(filePath, lens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lint: %v\n", err)
		return
	}

	if len(lintErrs) == 0 {
		fmt.Printf("✔ %s: schema valid\n", lens)
		return
	}

	for _, e := range lintErrs {
		fmt.Fprintf(os.Stderr, "⚠ [%s] %s\n", e.Lens, e.Message)
	}
}
