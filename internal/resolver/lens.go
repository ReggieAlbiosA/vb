package resolver

import (
	"fmt"

	"github.com/spf13/pflag"
)

// LensToFile maps CLI flag names to their corresponding vault filenames.
// --gui is intentionally absent — it is a rendering modifier, not a lens.
// It is parsed in the cmd layer and passed to the renderer in Phase 03/06.
var LensToFile = map[string]string{
	"why":        "WHY.md",
	"importance": "IMPORTANCE.md",
	"cli-tools":  "CLI_TOOLS.md",
	"arch":       "ARCH.md",
	"used":       "USED.md",
	"gotchas":    "GOTCHAS.md",
	"refs":       "REFS.md",
}

// ResolveLens converts a CLI flag name to its vault filename.
func ResolveLens(flag string) (string, error) {
	file, exists := LensToFile[flag]
	if !exists {
		return "", fmt.Errorf("%w: --%s", ErrInvalidLens, flag)
	}
	return file, nil
}

// ActiveLens inspects the flagset and returns the single active lens name.
// Errors if zero or multiple lenses are set.
func ActiveLens(flags *pflag.FlagSet) (string, error) {
	var active []string
	for name := range LensToFile {
		f := flags.Lookup(name)
		if f != nil && f.Changed {
			active = append(active, name)
		}
	}
	if len(active) == 0 {
		return "", ErrNoLens
	}
	if len(active) > 1 {
		return "", fmt.Errorf("%w: %v", ErrMultipleLenses, active)
	}
	return active[0], nil
}
