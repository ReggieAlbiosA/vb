package resolver_test

import (
	"errors"
	"testing"

	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/spf13/pflag"
)

func TestResolveLens_AllValidFlags(t *testing.T) {
	want := map[string]string{
		"why":        "WHY.md",
		"importance": "IMPORTANCE.md",
		"cli-tools":  "CLI_TOOLS.md",
		"arch":       "ARCH.md",
		"gotchas":    "GOTCHAS.md",
		"refs":       "REFS.md",
	}
	for flag, wantFile := range want {
		got, err := resolver.ResolveLens(flag)
		if err != nil {
			t.Errorf("ResolveLens(%q): unexpected error: %v", flag, err)
			continue
		}
		if got != wantFile {
			t.Errorf("ResolveLens(%q) = %q, want %q", flag, got, wantFile)
		}
	}
}

func TestResolveLens_InvalidFlag(t *testing.T) {
	_, err := resolver.ResolveLens("invalid-flag")
	if !errors.Is(err, resolver.ErrInvalidLens) {
		t.Errorf("expected ErrInvalidLens, got %v", err)
	}
}

func TestActiveLens_SingleFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("why", false, "")
	if err := fs.Parse([]string{"--why"}); err != nil {
		t.Fatal(err)
	}

	got, err := resolver.ActiveLens(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "why" {
		t.Errorf("got %q, want %q", got, "why")
	}
}

func TestActiveLens_NoFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	// Register all lenses so Lookup doesn't return nil, but none are Changed.
	for name := range resolver.LensToFile {
		fs.Bool(name, false, "")
	}

	_, err := resolver.ActiveLens(fs)
	if !errors.Is(err, resolver.ErrNoLens) {
		t.Errorf("expected ErrNoLens, got %v", err)
	}
}

func TestActiveLens_MultipleFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("why", false, "")
	fs.Bool("arch", false, "")
	if err := fs.Parse([]string{"--why", "--arch"}); err != nil {
		t.Fatal(err)
	}

	_, err := resolver.ActiveLens(fs)
	if !errors.Is(err, resolver.ErrMultipleLenses) {
		t.Errorf("expected ErrMultipleLenses, got %v", err)
	}
}

func TestActiveLens_UsedNotALens(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	for name := range resolver.LensToFile {
		fs.Bool(name, false, "")
	}
	fs.Bool("used", false, "")
	if err := fs.Parse([]string{"--used", "--why"}); err != nil {
		t.Fatal(err)
	}

	got, err := resolver.ActiveLens(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// --used is not in LensToFile — only --why must be returned.
	if got != "why" {
		t.Errorf("expected lens 'why' (--used ignored), got %q", got)
	}
}

func TestActiveLens_GUINotALens(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	// Register all lenses plus gui.
	for name := range resolver.LensToFile {
		fs.Bool(name, false, "")
	}
	fs.Bool("gui", false, "")
	if err := fs.Parse([]string{"--gui", "--why"}); err != nil {
		t.Fatal(err)
	}

	got, err := resolver.ActiveLens(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// gui is not in LensToFile — only why must be returned.
	if got != "why" {
		t.Errorf("expected lens 'why' (gui ignored), got %q", got)
	}
}
