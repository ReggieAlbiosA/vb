//go:build gui

package webview

import (
	"bytes"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
)

// htmlTemplate is the self-contained HTML page template for the webview.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>vb · {{.Lens}}</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    max-width: 800px;
    margin: 2rem auto;
    padding: 0 1rem;
    line-height: 1.6;
    color: {{if eq .Theme "light"}}#1a1a1a{{else}}#e0e0e0{{end}};
    background: {{if eq .Theme "light"}}#ffffff{{else}}#1e1e1e{{end}};
  }
  .lens-badge {
    display: inline-block;
    padding: 2px 10px;
    border-radius: 4px;
    font-weight: bold;
    font-size: 0.85rem;
    margin-bottom: 1rem;
    background: {{if eq .Theme "light"}}#e8f0fe{{else}}#2a3a4a{{end}};
    color: {{if eq .Theme "light"}}#1a73e8{{else}}#8ab4f8{{end}};
  }
  code { background: {{if eq .Theme "light"}}#f5f5f5{{else}}#2d2d2d{{end}}; padding: 2px 4px; border-radius: 3px; }
  pre { background: {{if eq .Theme "light"}}#f5f5f5{{else}}#2d2d2d{{end}}; padding: 1rem; border-radius: 6px; overflow-x: auto; }
  pre code { background: none; padding: 0; }
  a { color: {{if eq .Theme "light"}}#1a73e8{{else}}#8ab4f8{{end}}; }
</style>
{{if .IsMermaid}}
<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<script>
  mermaid.initialize({ startOnLoad: true, theme: '{{if eq .Theme "light"}}default{{else}}dark{{end}}' });
</script>
{{end}}
</head>
<body>
<div class="lens-badge">{{.LensBadge}}</div>
{{if .IsMermaid}}
<pre class="mermaid">
{{.Body}}
</pre>
{{else}}
{{.Body}}
{{end}}
</body>
</html>`

// templateData holds values passed into the HTML template.
type templateData struct {
	Lens      string
	LensBadge string
	Theme     string
	Body      template.HTML
	IsMermaid bool
}

// markdownToHTML converts raw Markdown bytes to HTML using goldmark.
func markdownToHTML(content []byte) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert(content, &buf); err != nil {
		return "", fmt.Errorf("converting markdown: %w", err)
	}
	return buf.String(), nil
}

// buildHTML produces a self-contained HTML page from Markdown content.
func buildHTML(content []byte, lens string, theme string) (string, error) {
	mermaid := isMermaidFile(content)

	var body string
	if mermaid {
		body = string(content)
	} else {
		var err error
		body, err = markdownToHTML(content)
		if err != nil {
			return "", err
		}
	}

	badge := strings.ToUpper(strings.ReplaceAll(lens, "-", " "))

	tmpl, err := template.New("vb").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData{
		Lens:      lens,
		LensBadge: badge,
		Theme:     theme,
		Body:      template.HTML(body),
		IsMermaid: mermaid,
	})
	if err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// htmlHandler returns an http.Handler that serves the HTML string and signals
// done on the first request (via once) so openWindow knows the browser connected.
func htmlHandler(html string, done chan struct{}) http.Handler {
	var once sync.Once
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
		once.Do(func() { close(done) })
	})
}

// openWindow launches a local HTTP server, opens the browser, and blocks until
// the browser has fetched the page — preventing the process from exiting before
// the server can serve the content.
func openWindow(title, html string) error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	done := make(chan struct{})
	addr := ln.Addr().String()
	go http.Serve(ln, htmlHandler(html, done)) //nolint:errcheck

	if err := launchBrowser("http://" + addr); err != nil {
		return err
	}

	<-done // block until browser has fetched the page, then exit cleanly
	return nil
}

// hasDisplay checks whether a graphical display is available on Linux.
func hasDisplay() bool {
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

// launchBrowser opens the given URL in the system's default browser.
// Returns an error if no display is available (headless) or the platform is unsupported.
func launchBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		if !hasDisplay() {
			return fmt.Errorf("no display available (DISPLAY and WAYLAND_DISPLAY unset)")
		}
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
