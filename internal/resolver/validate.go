package resolver

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// validFilename requires UPPERCASE letters and underscores only in the basename.
// Example valid: FILESYSTEM_FORMAT.md, NOTES.mmd
// Example invalid: filesystem.md, My-Notes.md
var validFilename = regexp.MustCompile(`^[A-Z][A-Z0-9_]*\.(md|mmd)$`)

// reservedFlags lists flags that cannot be used as custom lens names.
var reservedFlags = map[string]bool{
	"gui":  true,
	"used": true,
	"help": true,
}

// ValidateLensFilename checks that name is a valid custom lens filename.
// Must be uppercase with underscores, .md or .mmd extension only.
func ValidateLensFilename(name string) error {
	if name == "" {
		return fmt.Errorf("lens filename cannot be empty")
	}
	if !validFilename.MatchString(name) {
		return fmt.Errorf("invalid lens filename %q: must be UPPERCASE_WITH_UNDERSCORES.md or .mmd (e.g. FILESYSTEM_FORMAT.md)", name)
	}
	return nil
}

// FilenameToFlag converts an uppercase lens filename to a CLI flag name.
// FILESYSTEM_FORMAT.md → filesystem-format
// NOTES.mmd → notes
func FilenameToFlag(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return strings.ToLower(strings.ReplaceAll(base, "_", "-"))
}

// IsReservedFlag returns true if the flag name belongs to a built-in
// non-lens flag (gui, used, help) or an existing lens.
func IsReservedFlag(flag string) bool {
	if reservedFlags[flag] {
		return true
	}
	_, exists := LensToFile[flag]
	return exists
}
