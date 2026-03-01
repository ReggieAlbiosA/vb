# Native Desktop GUI — Replace browser with webview window

## Context

The `--gui` flag previously opened a browser tab via `xdg-open` (HTTP server + browser). This was broken on Wayland/Hyprland and unreliable across platforms. This enhancement replaces the browser approach with a native desktop window using platform-native web engines, plus multi-tab support so each `vb <topic> <flag> --gui` invocation adds a tab to the same running window.

## Usage

```bash
# Opens a native desktop window with rendered markdown
vb disk --why --gui

# Second invocation adds a tab to the existing window
vb disk --arch --gui

# Close the window — both processes exit cleanly
```

## What changed

| Before | After |
|--------|-------|
| `--gui` launched HTTP server + `xdg-open` | `--gui` opens a native webview window |
| Broken on Wayland/Hyprland | Works on X11, Wayland, macOS, Windows |
| Each invocation opened a new browser tab | Multi-tab: subsequent invocations add tabs via IPC |
| Process exited after browser fetched page | Primary window blocks until closed by user |
| No fallback on display failure | Graceful fallback to terminal renderer |

## Architecture

- **Linux**: WebKit2GTK (requires `libwebkit2gtk-4.1-dev`)
- **macOS**: WebKit (built-in, zero deps)
- **Windows**: WebView2/Edge (ships with Windows 10/11)

Multi-tab support via Unix socket IPC (Linux/macOS):
1. First `--gui` invocation: starts GUI window + IPC listener (blocks until window closed)
2. Subsequent invocations: connect to IPC socket, send tab data as JSON, exit immediately
3. Stale socket handling: if dial fails, remove socket file, caller becomes primary

## Implementation

### `internal/render/render.go`
- `GUIRendererFactory` signature changed from `func() Renderer` to `func(filename string) Renderer`
- `File()` extracts `filepath.Base(path)` and passes it to the factory for tab titles

### `internal/webview/renderer.go`
- `WebviewRenderer` gets a `Filename` field (injected via factory)
- `Render()` flow: check display → build tab content → try IPC send → start primary window → fall back to terminal

### `internal/webview/server.go`
- Removed: `openWindow()`, `htmlHandler()`, `launchBrowser()`, `hasDisplay()`
- Kept: `markdownToHTML()`, `buildHTML()` (backward compat/tests)
- Added: `buildTabContent()` — produces inner HTML for one tab without full page wrapper

### `internal/webview/tabs.go` (new)
- `TabData` struct: Title, Lens, BodyHTML, IsMermaid, Theme
- `buildTabShell(initialTab)` — full HTML page with tab bar, CSS, mermaid CDN, JS tab management
- `tabAddJS(id, tab)` — JS string for `webview.Eval()` to add a new tab dynamically
- `themeColors(theme)` — dark/light CSS color values

### `internal/webview/window.go` (new)
- `windowAPI` interface — abstracts webview for testability
- `webviewAdapter` wraps real `webview.WebView`
- `hasDisplay()` — moved here, expanded for macOS/Windows
- `startPrimaryWindow(initialTab)` — `runtime.LockOSThread()`, create webview, load tab shell, start IPC listener, `w.Run()` blocks

### `internal/webview/ipc.go` (new)
- `socketPath()` — platform-specific: `$XDG_RUNTIME_DIR/vb-gui.sock` (Linux), `$TMPDIR/vb-gui.sock` (macOS)
- `tryIPCSend(tab)` — dial socket, send JSON, read response
- `startIPCListener(w)` — accept connections, decode tab data, inject via `w.Dispatch()`
- Stale socket cleanup on failed dial

### `Makefile` (new)
- `build` / `build-gui` / `test` / `test-gui` targets
- GUI build requires `CGO_ENABLED=1 -tags gui`

### `go.mod`
- Added `github.com/webview/webview_go`
- `replace` directive to lvlrt fork for webkit2gtk-4.1 support (upstream PR #62 pending)

## Tests

- `internal/webview/tabs_test.go`: theme colors, buildTabShell (dark/light/mermaid/cli-tools), tabAddJS (basic, mermaid, XSS, quote escaping)
- `internal/webview/ipc_test.go`: socket path (default/override/linux), IPC round-trip (single/multi-tab), no listener, stale socket cleanup, invalid data
- `internal/webview/server_test.go`: updated factory test, added buildTabContent tests (markdown/mermaid/empty), removed tests for deleted functions
- `internal/webview/renderer_test.go`: updated WebviewRenderer with Filename field, hasDisplay tests moved here
- `internal/render/renderer_test.go`: updated factory mock signature

## Build prerequisites

```bash
# Linux
sudo apt install libwebkit2gtk-4.1-dev

# Then build
make build-gui
```
