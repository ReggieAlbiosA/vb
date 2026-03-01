//go:build gui

package webview

import (
	"fmt"
	"os"
	"runtime"

	webviewlib "github.com/webview/webview_go"
)

// windowAPI abstracts the webview operations needed by startPrimaryWindow,
// enabling test doubles without a real GUI.
type windowAPI interface {
	SetTitle(title string)
	SetSize(w, h int, hint webviewlib.Hint)
	SetHtml(html string)
	Eval(js string)
	Dispatch(f func())
	Run()
	Destroy()
}

// webviewAdapter wraps a real webview.WebView to satisfy windowAPI.
type webviewAdapter struct {
	w webviewlib.WebView
}

func (a *webviewAdapter) SetTitle(title string)                       { a.w.SetTitle(title) }
func (a *webviewAdapter) SetSize(w, h int, hint webviewlib.Hint)      { a.w.SetSize(w, h, hint) }
func (a *webviewAdapter) SetHtml(html string)                         { a.w.SetHtml(html) }
func (a *webviewAdapter) Eval(js string)                              { a.w.Eval(js) }
func (a *webviewAdapter) Dispatch(f func())                           { a.w.Dispatch(f) }
func (a *webviewAdapter) Run()                                        { a.w.Run() }
func (a *webviewAdapter) Destroy()                                    { a.w.Destroy() }

// newWindow creates a real webview window. Replaceable in tests.
var newWindow = func() (windowAPI, error) {
	w := webviewlib.New(false)
	if w == nil {
		return nil, fmt.Errorf("failed to create webview window")
	}
	return &webviewAdapter{w: w}, nil
}

// hasDisplay checks whether a graphical display is available.
func hasDisplay() bool {
	switch runtime.GOOS {
	case "linux":
		return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	case "darwin", "windows":
		return true
	default:
		return false
	}
}

// startPrimaryWindow creates the native webview window with the initial tab,
// starts the IPC listener for subsequent tab additions, and blocks until
// the window is closed.
func startPrimaryWindow(initialTab TabData) error {
	runtime.LockOSThread()

	w, err := newWindow()
	if err != nil {
		return fmt.Errorf("creating webview: %w", err)
	}
	defer w.Destroy()

	w.SetTitle("vb - " + initialTab.Title)
	w.SetSize(900, 700, webviewlib.HintNone)

	shellHTML, err := buildTabShell(initialTab)
	if err != nil {
		return fmt.Errorf("building tab shell: %w", err)
	}
	w.SetHtml(shellHTML)

	// Start IPC listener for additional tabs from other vb processes.
	stopIPC := startIPCListener(w)
	defer stopIPC()

	// Blocks until window is closed by user.
	w.Run()

	return nil
}
