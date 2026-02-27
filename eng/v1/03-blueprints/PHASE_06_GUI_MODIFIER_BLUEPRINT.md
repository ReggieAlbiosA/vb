# vb — Phase 06 · GUI Modifier

**Engineering Blueprint**

Optional Visual Layer · Webview Bridge · Mermaid Rendering  
Terminal-First, GUI-Optional Output Adapter

---

## Build Summary

### What
Phase 06 is the optional visual adapter for `vb`. It wires the `--gui` flag — stubbed since Phase 02 — into a real webview renderer that replaces terminal stdout for users running a graphical environment.

Three components:
- **Wails/Lorca bridge** — an isolated `WebviewRenderer` that satisfies Phase 03's `Renderer` interface and launches a lightweight webview window
- **`--gui` router** — the conditional in `render.File()` that diverts output from the terminal renderer to the webview renderer when `gui=true`
- **Mermaid loader** — detects and dynamically renders `ARCH.mmd` files in the webview using the Mermaid.js library

### Why
Phase 03's terminal renderer is the primary output path. The `--gui` modifier exists for `ARCH.mmd` (architecture diagrams) and rich Markdown viewing in environments where a terminal is insufficient. The webview is an adapter, not a replacement — every command works without it.

The `--gui` flag was parsed in Phase 02 and passed through Phase 03 as a no-op stub specifically so Phase 06 could wire it without touching any resolution logic. The interface boundary (`Renderer`) was placed in Phase 03 for the same reason.

### Importance
Phase 06 must be strictly optional — it cannot become a required dependency:

- Headless servers and CI environments must not fail when `--gui` is passed (graceful fallback to terminal)
- The webview adapter must implement `Renderer` — adding a new interface is forbidden
- `internal/webview` must have zero imports from `internal/resolver`, `internal/vault`, or `internal/linter`

The Mermaid loader is the one place where Phase 06 reads a different file type (`.mmd`) rather than `.md` — this is the only exception to the Markdown-only vault rule, and it only applies inside the webview.

---

## Package Structure

```
/internal
  /webview
    renderer.go   ← WebviewRenderer implementing render.Renderer interface
    server.go     ← embedded HTTP server serving HTML + Mermaid.js to webview window
    mermaid.go    ← .mmd file detection + Mermaid payload builder
```

> `WebviewRenderer` must implement `internal/render.Renderer` — no new interface.  
> `internal/webview` must NOT import any internal package except `internal/render` (for the interface type).  
> If Wails is unavailable (headless), `WebviewRenderer.Render()` must fall back to `TerminalRenderer`.  
> Do NOT add Wails as a required build dependency — use a build tag: `//go:build gui`.

---

## Build Tag Strategy

```
Normal build:  go build ./...         ← webview package excluded
GUI build:     go build -tags gui ./... ← webview package included
```

```go
// internal/webview/renderer.go
//go:build gui

package webview
```

```go
// internal/webview/stub.go (no build tag — always compiled)
// Provides the fallback when compiled without -tags gui.

//go:build !gui

package webview

import "github.com/yourname/vb/internal/render"

// NewRenderer returns a terminal renderer as the fallback when
// the gui build tag is not set.
func NewRenderer() render.Renderer {
    return &render.TerminalRenderer{}
}
```

> This pattern means `render.File()` always calls `webview.NewRenderer()` when `gui=true`,  
> but in non-GUI builds it transparently gets a `TerminalRenderer`.  
> No conditional compilation scattered across `render.File()`.

---

## 00 · CLI Wiring Update — cmd/query.go

No new command is needed. Phase 02 already registered `--gui` as a flag on `queryCmd`.  
Phase 06 updates `render.File()` to wire the real renderer:

```go
// cmd/query.go — update render call site
return render.File(filePath, lens, flagGUI, cfg.Theme)
```

`render.File()` is updated in Phase 06 to call `webview.NewRenderer()` when `gui=true`:

```go
// internal/render/renderer.go — updated File()
func File(path string, lens string, gui bool, theme string) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("reading %s: %w", path, err)
    }

    var r Renderer
    if gui {
        r = webview.NewRenderer()   // webview or terminal stub depending on build tag
    } else {
        r = &TerminalRenderer{}
    }

    out, err := r.Render(content, lens, theme)
    if err != nil {
        return err
    }

    fmt.Print(out)
    return nil
}
```

---

## 01 · WebviewRenderer — internal/webview/renderer.go

```go
//go:build gui

package webview

import (
    "fmt"
    "github.com/yourname/vb/internal/render"
)

// WebviewRenderer satisfies render.Renderer by launching a webview window.
type WebviewRenderer struct{}

// Render converts Markdown content to HTML and opens it in a webview window.
// Falls back to terminal output if the display is unavailable.
func (r *WebviewRenderer) Render(content []byte, lens string, theme string) (string, error) {
    html, err := buildHTML(content, lens, theme)
    if err != nil {
        return "", fmt.Errorf("building webview HTML: %w", err)
    }

    if err := openWindow(lens, html); err != nil {
        // Graceful fallback: headless or display unavailable.
        t := &render.TerminalRenderer{}
        return t.Render(content, lens, theme)
    }

    // Webview blocks until window is closed.
    return "", nil
}

// NewRenderer returns a WebviewRenderer (gui build tag active).
func NewRenderer() render.Renderer {
    return &WebviewRenderer{}
}
```

---

## 02 · Embedded Server — internal/webview/server.go

```go
//go:build gui

package webview

import (
    "embed"
    "html/template"
    "net"
    "net/http"
)

//go:embed assets/mermaid.min.js
var mermaidJS embed.FS

// buildHTML produces a self-contained HTML page from Markdown content.
// Mermaid.js is embedded — no CDN call required.
func buildHTML(content []byte, lens string, theme string) (string, error) {
    // Convert Markdown to HTML using goldmark (structural, not styled).
    // Styling is applied via embedded CSS in the HTML template.
    htmlBody, err := markdownToHTML(content)
    if err != nil {
        return "", err
    }

    isMermaid := isMermaidFile(content)

    tmpl, err := template.New("vb").Parse(htmlTemplate)
    if err != nil {
        return "", err
    }

    // ... render template with htmlBody, lens badge, theme class, isMermaid flag
}

// openWindow launches the OS webview pointing to a local HTTP server.
// Returns an error if DISPLAY is unavailable or webview init fails.
func openWindow(title, html string) error {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        return err
    }
    addr := ln.Addr().String()

    go http.Serve(ln, htmlHandler(html))

    // Use Lorca or Wails depending on availability.
    return launchWebview(title, "http://"+addr)
}
```

---

## 03 · Mermaid Loader — internal/webview/mermaid.go

```go
//go:build gui

package webview

import (
    "bytes"
    "path/filepath"
    "strings"
)

// isMermaidFile returns true if the content appears to be a Mermaid diagram
// (starts with a known diagram type keyword).
var mermaidKeywords = []string{
    "graph ", "flowchart ", "sequenceDiagram", "classDiagram",
    "stateDiagram", "erDiagram", "gantt", "pie ", "gitGraph",
}

func isMermaidFile(content []byte) bool {
    trimmed := bytes.TrimSpace(content)
    for _, kw := range mermaidKeywords {
        if bytes.HasPrefix(trimmed, []byte(kw)) {
            return true
        }
    }
    return false
}

// MermaidExtension returns true if the file has a .mmd extension.
// Used by cmd layer to detect ARCH.mmd before passing content.
func MermaidExtension(path string) bool {
    return strings.ToLower(filepath.Ext(path)) == ".mmd"
}
```

**Mermaid rendering in HTML template:**

```html
<!-- If isMermaid == true, wrap content in <pre class="mermaid"> -->
<!-- Mermaid.js auto-renders on DOMContentLoaded -->
<script type="module">
  import mermaid from './mermaid.min.js';
  mermaid.initialize({ startOnLoad: true, theme: '{{.Theme}}' });
</script>
```

---

## Graceful Fallback Contract

```
$DISPLAY unset (headless/SSH):
  webview.NewRenderer() → WebviewRenderer
  WebviewRenderer.Render() → openWindow() fails
  Falls through to TerminalRenderer.Render()
  Terminal output produced, no error

Non-GUI build (no -tags gui):
  webview.NewRenderer() → TerminalRenderer (stub)
  render.File() calls TerminalRenderer directly
  No webview code compiled at all
```

---

## Validation Checklist

- ✔ `vb disk --arch --gui` in GUI build → webview window opens with styled Markdown
- ✔ `vb disk --arch --gui` in non-GUI build → terminal output (stub fallback)
- ✔ `vb disk --arch --gui` in headless environment → graceful terminal fallback, exits 0
- ✔ `vb disk --arch` (no `--gui`) → terminal output unchanged from Phase 03
- ✔ `ARCH.mmd` content in webview → Mermaid diagram rendered, not raw text
- ✔ Mermaid.js served from embedded asset — no network call
- ✔ `internal/webview` has no import from `internal/resolver`, `internal/vault`, `internal/linter`
- ✔ `WebviewRenderer` satisfies `render.Renderer` interface at compile time
- ✔ Resolution path (Phase 02) is untouched — `--gui` is a rendering modifier only
- ✔ `go build ./...` (without `-tags gui`) completes without any webview dependency errors

---

## Test Coverage Requirements

Thresholds: **≥80% on internal/webview · ≥80% on cmd**

> Note: webview tests only execute in GUI build. Use build tags on test files.

### internal/webview/mermaid_test.go

| Test | Covers |
|---|---|
| `TestIsMermaidFile_GraphKeyword` | `"graph TD"` content → true |
| `TestIsMermaidFile_SequenceDiagram` | `"sequenceDiagram"` content → true |
| `TestIsMermaidFile_MarkdownContent` | Standard Markdown → false |
| `TestIsMermaidFile_EmptyContent` | Empty bytes → false |
| `TestMermaidExtension_MMD` | `.mmd` file path → true |
| `TestMermaidExtension_MD` | `.md` file path → false |

### internal/webview/renderer_test.go (build: gui)

| Test | Covers |
|---|---|
| `TestWebviewRenderer_FallbackOnHeadless` | No display → falls through to terminal output |
| `TestBuildHTML_ValidMarkdown` | Markdown content → non-empty HTML string |
| `TestBuildHTML_MermaidContent` | Mermaid content → HTML with `<pre class="mermaid">` |

### internal/webview/stub_test.go (build: !gui)

| Test | Covers |
|---|---|
| `TestNewRenderer_IsTerminal` | Non-GUI build returns `*TerminalRenderer` |

### cmd/query_test.go — additions

| Test | Covers |
|---|---|
| `TestQueryCmd_GUIFlag_NonGUIBuild` | `--gui` in non-GUI build → terminal output, no error |

---

```
vb engineering blueprint
phase 06 · gui modifier
```
