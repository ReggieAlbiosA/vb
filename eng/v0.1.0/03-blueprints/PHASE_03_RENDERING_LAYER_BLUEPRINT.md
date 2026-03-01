# vb — Phase 03 · Rendering Layer

**Engineering Blueprint**

Output Abstraction · Markdown Styled Terminal Rendering  
Resolution-Independent Presentation Layer

---

## Build Summary

### What
Phase 03 is the output layer of `vb`. It takes the resolved file path from Phase 02 and renders its Markdown content as styled terminal output. It has no knowledge of topics, lenses, or the vault — it only cares about content and presentation.

Three components:
- **Renderer interface** — the abstraction contract both terminal and future GUI renderers implement
- **Terminal renderer** — Glamour adapter that reads raw Markdown and outputs styled terminal text, using the user's `theme` from vault config
- **Lens badge** — a lightweight lens-aware header rendered above the Glamour block (not a border wrapper around it)

### Why
Phase 02 returns a file path. That path means nothing to the user until its content is presented clearly. Phase 03 closes the loop: the user runs `vb disk --why` and sees formatted, readable output — not raw Markdown bytes.

The renderer is defined as an interface so Phase 06's `--gui` modifier can plug in a different renderer (webview) without touching any resolution logic.

### Importance
Phase 03 is the moment `vb` becomes a usable tool — `v0.1` ships after this phase. Every subsequent phase builds on the assumption that `render.File(path, gui)` works correctly. The `theme` config key from Phase 01 becomes active here for the first time.

The decision to feed **raw Markdown directly to Glamour** (rather than pre-parsing with Goldmark first) is a critical correctness constraint. Glamour handles its own parsing internally — double-processing produces broken output.

---

## Package Structure

```
/internal
  /render
    renderer.go   ← Renderer interface + render.File() public entry point
    terminal.go   ← Glamour adapter (reads config theme)
    theme.go      ← lens badge colours (Lip Gloss header only, not border)
```

> Goldmark is **not used directly** — Glamour uses Goldmark internally.  
> Do NOT add a `/markdown/` sub-package or a `ParseMarkdown()` step.  
> Rendering layer must NOT import resolver, vault, or index.  
> Input = resolved file path + gui bool.  
> Output = styled string printed to stdout.

---

## Rendering Flow

```
Input:
  filePath string   ← from Phase 02 resolver
  gui      bool     ← from --gui flag in cmd layer (Phase 06 modifier)

Execution Order:
  1. Read raw Markdown bytes from filePath (os.ReadFile)
  2. Load theme from vault config (cfg.Theme = "dark" | "light")
  3. Feed raw bytes directly to Glamour with config theme
  4. Prepend lens badge (coloured header line via Lip Gloss)
  5. Print final output to stdout
  6. If gui == true: hand off to Phase 06 webview renderer instead of stdout
```

> Goldmark is NOT called explicitly — step 3 feeds **raw Markdown** to Glamour.  
> Glamour handles all parsing internally.

---

## 00 · Public Entry Point — internal/render/renderer.go

This is what `cmd/query.go` calls. Matches the Phase 02 call site: `render.File(filePath, flagGUI)`.

```go
// Renderer is the abstraction both terminal and GUI renderers implement.
// Phase 06 will provide a WebviewRenderer satisfying this interface.
type Renderer interface {
    Render(content []byte, lens string, theme string) (string, error)
}

// File is the public entry point called by cmd/query.go.
// It reads the file at path and dispatches to the appropriate renderer.
func File(path string, lens string, gui bool, theme string) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("reading %s: %w", path, err)
    }

    if gui {
        // Phase 06 — hand off to webview renderer.
        // No-op until Phase 06 is implemented; falls through to terminal.
    }

    r := &TerminalRenderer{}
    out, err := r.Render(content, lens, theme)
    if err != nil {
        return err
    }

    fmt.Print(out)
    return nil
}
```

> **Note:** `cmd/query.go` from Phase 02 passes `flagGUI bool`. Update the call site in Phase 03  
> to also pass `lens` and `cfg.Theme` so the renderer can apply the correct badge and theme.

---

## 01 · Terminal Renderer — internal/render/terminal.go

```go
type TerminalRenderer struct{}

// Render takes raw Markdown bytes and returns styled terminal output.
// Raw bytes are fed directly to Glamour — no pre-parsing step.
func (r *TerminalRenderer) Render(content []byte, lens string, theme string) (string, error) {
    // Use vault config theme ("dark" | "light"), not WithAutoStyle().
    // WithAutoStyle() ignores the user's explicit config preference.
    renderer, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle(theme),
        glamour.WithWordWrap(100),
    )
    if err != nil {
        return "", fmt.Errorf("creating renderer: %w", err)
    }

    // Feed raw Markdown directly — Glamour parses internally via Goldmark.
    rendered, err := renderer.Render(string(content))
    if err != nil {
        return "", fmt.Errorf("rendering markdown: %w", err)
    }

    // Prepend lens badge above Glamour output.
    badge := LensBadge(lens)
    return badge + rendered, nil
}
```

---

## 02 · Lens Badge — internal/render/theme.go

A lightweight coloured header line above the Glamour block.  
**Not a border wrapper** — wrapping Glamour's ANSI output in Lip Gloss borders causes visual conflicts between the two styling systems.

```go
// LensBadge returns a single coloured header line identifying the active lens.
// It sits above the Glamour-rendered content, not wrapped around it.
func LensBadge(lens string) string {
    style := lipgloss.NewStyle().
        Bold(true).
        Padding(0, 1)

    switch lens {
    case "why":
        return style.Foreground(lipgloss.Color("10")).Render("◆ WHY") + "\n"
    case "importance":
        return style.Foreground(lipgloss.Color("11")).Render("◆ IMPORTANCE") + "\n"
    case "cli-tools":
        return style.Foreground(lipgloss.Color("14")).Render("◆ CLI TOOLS") + "\n"
    case "arch":
        return style.Foreground(lipgloss.Color("12")).Render("◆ ARCH") + "\n"
    case "used":
        return style.Foreground(lipgloss.Color("13")).Render("◆ USED") + "\n"
    case "gotchas":
        return style.Foreground(lipgloss.Color("9")).Render("◆ GOTCHAS") + "\n"
    case "refs":
        return style.Foreground(lipgloss.Color("8")).Render("◆ REFS") + "\n"
    default:
        return ""
    }
}
```

> `lipgloss.Style.Render()` is only used on the badge — a short string with no ANSI content.  
> It is never called on Glamour's already-styled output.

---

## Validation Checklist

- ✔ `vb disk --why` prints coloured `◆ WHY` badge + Glamour-rendered WHY.md
- ✔ `vb disk --arch` prints `◆ ARCH` badge with different colour — same content, different badge
- ✔ Raw `.md` content unchanged — no auto-modification by renderer
- ✔ `theme = "dark"` in config → Glamour dark style applied (not auto-detected)
- ✔ `theme = "light"` in config → Glamour light style applied
- ✔ Resolver logic untouched — `internal/render/` imports nothing from `internal/resolver/`
- ✔ No filesystem lookup inside render layer — only `os.ReadFile` at the entry point
- ✔ `gui = true` path is a no-op stub that falls through to terminal (Phase 06 wires it)

---

## Test Coverage Requirements

Thresholds: **≥90% on internal/render · ≥80% on cmd coverage maintained**

### internal/render/renderer_test.go

| Test | Covers |
|---|---|
| `TestFile_Success` | Valid path + lens → prints styled output without error |
| `TestFile_FileNotFound` | Non-existent path → error before render |
| `TestFile_GUIFallthrough` | `gui=true` falls through to terminal renderer (Phase 06 stub) |

### internal/render/terminal_test.go

| Test | Covers |
|---|---|
| `TestTerminalRenderer_DarkTheme` | `theme="dark"` → renders without error |
| `TestTerminalRenderer_LightTheme` | `theme="light"` → renders without error |
| `TestTerminalRenderer_OutputContainsBadge` | Badge prepended above rendered content |
| `TestTerminalRenderer_ContentUnmodified` | Source Markdown structure preserved in output |

### internal/render/theme_test.go

| Test | Covers |
|---|---|
| `TestLensBadge_AllLenses` | All 7 lenses return non-empty badge string |
| `TestLensBadge_UnknownLens` | Unknown lens returns empty string (no panic) |

> `cmd/` coverage is maintained by Phase 02's existing `cmd/query_test.go`.  
> Phase 03 updates the `render.File()` call site in `cmd/query.go` — rerun existing cmd tests to confirm.

---

```
vb engineering blueprint
phase 03 · rendering abstraction
```