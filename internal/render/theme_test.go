package render_test

import (
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/render"
)

func TestLensBadge_AllLenses(t *testing.T) {
	lenses := []string{"why", "importance", "cli-tools", "arch", "used", "gotchas", "refs"}
	for _, lens := range lenses {
		got := render.LensBadge(lens)
		if got == "" {
			t.Errorf("LensBadge(%q) returned empty string, want non-empty", lens)
		}
	}
}

func TestLensBadge_UnknownLens(t *testing.T) {
	got := render.LensBadge("unknown-lens")
	if got != "" {
		t.Errorf("LensBadge(unknown) = %q, want empty string", got)
	}
}
