//go:build gui

package webview

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	webviewlib "github.com/webview/webview_go"
)

// mockWindow implements windowAPI for testing without a real GUI.
type mockWindow struct {
	mu       sync.Mutex
	title    string
	evalCalls []string
	dispatchCalls int
}

func (m *mockWindow) SetTitle(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.title = title
}

func (m *mockWindow) SetSize(w, h int, hint webviewlib.Hint) {}
func (m *mockWindow) SetHtml(html string)                     {}

func (m *mockWindow) Eval(js string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evalCalls = append(m.evalCalls, js)
}

func (m *mockWindow) Dispatch(f func()) {
	m.mu.Lock()
	m.dispatchCalls++
	m.mu.Unlock()
	// Execute immediately in test (no UI thread).
	f()
}

func (m *mockWindow) Run()     {}
func (m *mockWindow) Destroy() {}

func TestSocketPath_Default(t *testing.T) {
	old := socketPathOverride
	socketPathOverride = ""
	defer func() { socketPathOverride = old }()

	path := socketPath()
	if path == "" {
		t.Error("expected non-empty socket path")
	}
	if !strings.Contains(path, "vb-gui.sock") {
		t.Errorf("expected socket path to contain vb-gui.sock, got %q", path)
	}
}

func TestSocketPath_Override(t *testing.T) {
	old := socketPathOverride
	socketPathOverride = "/tmp/test-vb-gui.sock"
	defer func() { socketPathOverride = old }()

	path := socketPath()
	if path != "/tmp/test-vb-gui.sock" {
		t.Errorf("expected override path, got %q", path)
	}
}

func TestSocketPath_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific test")
	}

	old := socketPathOverride
	socketPathOverride = ""
	defer func() { socketPathOverride = old }()

	xdg := os.Getenv("XDG_RUNTIME_DIR")
	if xdg != "" {
		path := socketPath()
		if !strings.HasPrefix(path, xdg) {
			t.Errorf("expected path to start with XDG_RUNTIME_DIR %q, got %q", xdg, path)
		}
	}
}

func TestIPCRoundTrip(t *testing.T) {
	// Use a temp socket path to avoid conflicts.
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test-vb.sock")

	old := socketPathOverride
	socketPathOverride = sockPath
	defer func() { socketPathOverride = old }()

	mock := &mockWindow{}
	stop := startIPCListener(mock)
	defer stop()

	// Give listener time to start.
	time.Sleep(50 * time.Millisecond)

	tab := TabData{
		Title:    "REFS.md",
		Lens:     "refs",
		BodyHTML: "<p>References</p>",
		Theme:    "dark",
	}

	err := tryIPCSend(tab)
	if err != nil {
		t.Fatalf("IPC send failed: %v", err)
	}

	// Give dispatch time to execute.
	time.Sleep(50 * time.Millisecond)

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if mock.dispatchCalls != 1 {
		t.Errorf("expected 1 dispatch call, got %d", mock.dispatchCalls)
	}
	if len(mock.evalCalls) != 1 {
		t.Errorf("expected 1 eval call, got %d", len(mock.evalCalls))
	}
	if mock.title != "vb - REFS.md" {
		t.Errorf("expected title 'vb - REFS.md', got %q", mock.title)
	}
}

func TestIPCRoundTrip_MultipleTabs(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test-vb.sock")

	old := socketPathOverride
	socketPathOverride = sockPath
	defer func() { socketPathOverride = old }()

	mock := &mockWindow{}
	stop := startIPCListener(mock)
	defer stop()

	time.Sleep(50 * time.Millisecond)

	tabs := []TabData{
		{Title: "WHY.md", Lens: "why", BodyHTML: "<p>Why</p>", Theme: "dark"},
		{Title: "ARCH.mmd", Lens: "arch", BodyHTML: "graph TD\n  A-->B", IsMermaid: true, Theme: "dark"},
	}

	for _, tab := range tabs {
		if err := tryIPCSend(tab); err != nil {
			t.Fatalf("IPC send failed for %s: %v", tab.Title, err)
		}
	}

	time.Sleep(50 * time.Millisecond)

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if mock.dispatchCalls != 2 {
		t.Errorf("expected 2 dispatch calls, got %d", mock.dispatchCalls)
	}
	if mock.title != "vb - ARCH.mmd" {
		t.Errorf("expected title 'vb - ARCH.mmd', got %q", mock.title)
	}
}

func TestTryIPCSend_NoListener(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "no-listener.sock")

	old := socketPathOverride
	socketPathOverride = sockPath
	defer func() { socketPathOverride = old }()

	tab := TabData{Title: "WHY.md", Lens: "why", BodyHTML: "<p>test</p>", Theme: "dark"}
	err := tryIPCSend(tab)
	if err == nil {
		t.Error("expected error when no listener is running")
	}
}

func TestTryIPCSend_StaleSocketCleanup(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "stale.sock")

	old := socketPathOverride
	socketPathOverride = sockPath
	defer func() { socketPathOverride = old }()

	// Create a stale socket file.
	if err := os.WriteFile(sockPath, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}

	tab := TabData{Title: "WHY.md", Lens: "why", BodyHTML: "<p>test</p>", Theme: "dark"}
	_ = tryIPCSend(tab)

	// Stale socket should be cleaned up.
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("expected stale socket to be removed")
	}
}

func TestHandleIPCConn_InvalidData(t *testing.T) {
	mock := &mockWindow{}

	// Create a pipe to simulate a connection.
	server, client := net.Pipe()
	defer server.Close()

	go func() {
		// Send invalid JSON.
		client.Write([]byte("not json\n"))

		// Read response.
		var resp IPCResponse
		json.NewDecoder(client).Decode(&resp)
		client.Close()
	}()

	var tabCounter atomic.Int64
	handleIPCConn(server, mock, &tabCounter)

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if len(mock.evalCalls) != 0 {
		t.Error("expected no eval calls for invalid data")
	}
}
