//go:build gui

package webview

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

// IPCResponse is sent back to the connecting client after a tab is added.
type IPCResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// socketPath returns the platform-specific IPC socket/pipe path.
// Overridable via socketPathOverride for testing.
var socketPathOverride string

func socketPath() string {
	if socketPathOverride != "" {
		return socketPathOverride
	}

	switch runtime.GOOS {
	case "linux":
		if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
			return dir + "/vb-gui.sock"
		}
		return "/tmp/vb-gui.sock"
	case "darwin":
		if dir := os.Getenv("TMPDIR"); dir != "" {
			return dir + "vb-gui.sock"
		}
		return "/tmp/vb-gui.sock"
	default:
		return "/tmp/vb-gui.sock"
	}
}

// tryIPCSend attempts to connect to an existing vb GUI window via IPC and send
// tab data. Returns nil on success (tab was added to existing window).
// Returns an error if no window is running or the connection fails.
func tryIPCSend(tab TabData) error {
	path := socketPath()

	conn, err := net.Dial("unix", path)
	if err != nil {
		// Connection failed — no running window or stale socket.
		// Try to clean up a stale socket file.
		_ = os.Remove(path)
		return fmt.Errorf("no running vb gui window: %w", err)
	}
	defer conn.Close()

	// Send tab data as JSON.
	if err := json.NewEncoder(conn).Encode(tab); err != nil {
		return fmt.Errorf("sending tab data: %w", err)
	}

	// Read response.
	var resp IPCResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("reading IPC response: %w", err)
	}

	if resp.Status != "ok" {
		return fmt.Errorf("IPC error: %s", resp.Message)
	}

	return nil
}

// startIPCListener creates a Unix socket listener that accepts tab data
// from other vb processes and injects them into the running webview window.
// Returns a stop function that closes the listener and removes the socket.
func startIPCListener(w windowAPI) func() {
	path := socketPath()

	// Remove any stale socket from a previous run.
	_ = os.Remove(path)

	ln, err := net.Listen("unix", path)
	if err != nil {
		// Non-fatal: IPC won't work but the window still functions.
		fmt.Fprintf(os.Stderr, "vb: warning: IPC listener failed: %v\n", err)
		return func() {}
	}

	var tabCounter atomic.Int64
	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					continue
				}
			}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				handleIPCConn(c, w, &tabCounter)
			}(conn)
		}
	}()

	return func() {
		close(done)
		ln.Close()
		_ = os.Remove(path)
		wg.Wait()
	}
}

// handleIPCConn processes a single IPC connection: decodes tab data,
// injects it into the webview, and sends a response.
func handleIPCConn(conn net.Conn, w windowAPI, counter *atomic.Int64) {
	var tab TabData
	if err := json.NewDecoder(conn).Decode(&tab); err != nil {
		resp := IPCResponse{Status: "error", Message: "invalid tab data"}
		json.NewEncoder(conn).Encode(resp) //nolint:errcheck
		return
	}

	id := fmt.Sprintf("tab%d", counter.Add(1))
	js := tabAddJS(id, tab)

	// Dispatch to the UI thread (required by webview).
	w.Dispatch(func() {
		w.Eval(js)
		w.SetTitle("vb - " + tab.Title)
	})

	resp := IPCResponse{Status: "ok"}
	json.NewEncoder(conn).Encode(resp) //nolint:errcheck
}
