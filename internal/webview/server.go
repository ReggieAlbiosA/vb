//go:build gui

package webview

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/yuin/goldmark"
)

// htmlTemplate is the self-contained HTML page template for standalone rendering.
// Kept for backward compatibility with buildHTML (used in tests).
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

// buildTabContent produces a TabData struct for a single tab, converting
// raw Markdown content to inner HTML without a full page wrapper.
func buildTabContent(content []byte, lens string, filename string, theme string) (TabData, error) {
	mermaid := isMermaidFile(content)

	var bodyHTML string
	if mermaid {
		bodyHTML = string(content)
	} else {
		var err error
		bodyHTML, err = markdownToHTML(content)
		if err != nil {
			return TabData{}, err
		}
	}

	return TabData{
		Title:     filename,
		Lens:      lens,
		BodyHTML:  bodyHTML,
		IsMermaid: mermaid,
		Theme:     theme,
	}, nil
}
