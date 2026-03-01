//go:build gui

package webview

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"strings"
)

// TabData holds the data needed to render a single tab in the webview window.
type TabData struct {
	Title     string // Tab title (e.g. "WHY.md")
	Lens      string // Lens name (e.g. "why")
	BodyHTML  string // Inner HTML content for the tab body
	IsMermaid bool   // Whether content is a Mermaid diagram
	Theme     string // "dark" or "light"
}

// themeColors returns CSS color values for the given theme.
func themeColors(theme string) (bg, fg, tabBg, tabActiveBg, tabFg, tabActiveFg, badgeBg, badgeFg, codeBg, linkColor string) {
	if theme == "light" {
		return "#ffffff", "#1a1a1a",
			"#e8e8e8", "#ffffff",
			"#666666", "#1a1a1a",
			"#e8f0fe", "#1a73e8",
			"#f5f5f5", "#1a73e8"
	}
	// dark (default)
	return "#1e1e1e", "#e0e0e0",
		"#2a2a2a", "#1e1e1e",
		"#888888", "#e0e0e0",
		"#2a3a4a", "#8ab4f8",
		"#2d2d2d", "#8ab4f8"
}

// tabShellTemplate is the full HTML page with tab bar and initial tab content.
const tabShellTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>vb</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: {{.Bg}};
    color: {{.Fg}};
    line-height: 1.6;
  }
  #vb-tab-bar {
    display: flex;
    background: {{.TabBg}};
    border-bottom: 1px solid {{.TabBg}};
    overflow-x: auto;
    flex-shrink: 0;
  }
  .vb-tab {
    padding: 8px 16px;
    cursor: pointer;
    border: none;
    background: {{.TabBg}};
    color: {{.TabFg}};
    font-size: 0.85rem;
    white-space: nowrap;
    border-bottom: 2px solid transparent;
    transition: background 0.15s, color 0.15s;
  }
  .vb-tab:hover { background: {{.TabActiveBg}}; }
  .vb-tab.active {
    background: {{.TabActiveBg}};
    color: {{.TabActiveFg}};
    border-bottom-color: {{.BadgeFg}};
  }
  .vb-tab-content {
    display: none;
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem 1rem;
  }
  .vb-tab-content.active { display: block; }
  .lens-badge {
    display: inline-block;
    padding: 2px 10px;
    border-radius: 4px;
    font-weight: bold;
    font-size: 0.85rem;
    margin-bottom: 1rem;
    background: {{.BadgeBg}};
    color: {{.BadgeFg}};
  }
  code { background: {{.CodeBg}}; padding: 2px 4px; border-radius: 3px; }
  pre { background: {{.CodeBg}}; padding: 1rem; border-radius: 6px; overflow-x: auto; }
  pre code { background: none; padding: 0; }
  a { color: {{.LinkColor}}; }
</style>
{{if .InitialTab.IsMermaid}}
<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<script>
  mermaid.initialize({ startOnLoad: false, theme: '{{if eq .Theme "light"}}default{{else}}dark{{end}}' });
</script>
{{end}}
<script>
var vbMermaidLoaded = {{.InitialTab.IsMermaid}};

function vbSwitchTab(id) {
  document.querySelectorAll('.vb-tab').forEach(function(t) { t.classList.remove('active'); });
  document.querySelectorAll('.vb-tab-content').forEach(function(c) { c.classList.remove('active'); });
  var tab = document.getElementById('vb-tab-btn-' + id);
  var content = document.getElementById('vb-tab-' + id);
  if (tab) tab.classList.add('active');
  if (content) content.classList.add('active');
}

function vbAddTab(id, title, bodyHTML, isMermaid) {
  // Load mermaid if needed and not already loaded
  if (isMermaid && !vbMermaidLoaded) {
    var s = document.createElement('script');
    s.src = 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js';
    s.onload = function() {
      mermaid.initialize({ startOnLoad: false, theme: document.body.style.background === '#ffffff' ? 'default' : 'dark' });
      vbMermaidLoaded = true;
      vbFinishAddTab(id, title, bodyHTML, isMermaid);
    };
    document.head.appendChild(s);
    return;
  }
  vbFinishAddTab(id, title, bodyHTML, isMermaid);
}

function vbFinishAddTab(id, title, bodyHTML, isMermaid) {
  // Create tab button
  var btn = document.createElement('button');
  btn.id = 'vb-tab-btn-' + id;
  btn.className = 'vb-tab';
  btn.textContent = title;
  btn.onclick = function() { vbSwitchTab(id); };
  document.getElementById('vb-tab-bar').appendChild(btn);

  // Create tab content
  var div = document.createElement('div');
  div.id = 'vb-tab-' + id;
  div.className = 'vb-tab-content';
  div.innerHTML = bodyHTML;
  document.getElementById('vb-tab-panels').appendChild(div);

  // Render mermaid diagrams if needed
  if (isMermaid && typeof mermaid !== 'undefined') {
    mermaid.run({ nodes: div.querySelectorAll('.mermaid') });
  }

  // Switch to new tab
  vbSwitchTab(id);
}
</script>
</head>
<body>
<div id="vb-tab-bar">
  <button id="vb-tab-btn-{{.InitialID}}" class="vb-tab active" onclick="vbSwitchTab('{{.InitialID}}')">{{.InitialTab.Title}}</button>
</div>
<div id="vb-tab-panels">
  <div id="vb-tab-{{.InitialID}}" class="vb-tab-content active">
    <div class="lens-badge">{{.InitialBadge}}</div>
    {{if .InitialTab.IsMermaid}}
    <pre class="mermaid">{{.InitialTab.BodyHTML}}</pre>
    {{else}}
    {{.InitialBody}}
    {{end}}
  </div>
</div>
{{if .InitialTab.IsMermaid}}
<script>mermaid.run();</script>
{{end}}
</body>
</html>`

// tabShellData is the template data for the full tab shell HTML.
type tabShellData struct {
	Theme        string
	Bg           string
	Fg           string
	TabBg        string
	TabActiveBg  string
	TabFg        string
	TabActiveFg  string
	BadgeBg      string
	BadgeFg      string
	CodeBg       string
	LinkColor    string
	InitialID    string
	InitialTab   TabData
	InitialBadge string
	InitialBody  template.HTML
}

// buildTabShell produces a full HTML page with tab bar and the initial tab content.
func buildTabShell(initialTab TabData) (string, error) {
	bg, fg, tabBg, tabActiveBg, tabFg, tabActiveFg, badgeBg, badgeFg, codeBg, linkColor := themeColors(initialTab.Theme)
	badge := strings.ToUpper(strings.ReplaceAll(initialTab.Lens, "-", " "))

	data := tabShellData{
		Theme:        initialTab.Theme,
		Bg:           bg,
		Fg:           fg,
		TabBg:        tabBg,
		TabActiveBg:  tabActiveBg,
		TabFg:        tabFg,
		TabActiveFg:  tabActiveFg,
		BadgeBg:      badgeBg,
		BadgeFg:      badgeFg,
		CodeBg:       codeBg,
		LinkColor:    linkColor,
		InitialID:    "tab0",
		InitialTab:   initialTab,
		InitialBadge: badge,
		InitialBody:  template.HTML(initialTab.BodyHTML),
	}

	tmpl, err := template.New("tabshell").Parse(tabShellTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing tab shell template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing tab shell template: %w", err)
	}

	return buf.String(), nil
}

// tabAddJS returns a JavaScript string that, when eval'd, adds a new tab dynamically.
func tabAddJS(id string, tab TabData) string {
	badge := strings.ToUpper(strings.ReplaceAll(tab.Lens, "-", " "))

	var bodyHTML string
	if tab.IsMermaid {
		bodyHTML = fmt.Sprintf(`<div class="lens-badge">%s</div><pre class="mermaid">%s</pre>`,
			html.EscapeString(badge),
			html.EscapeString(tab.BodyHTML))
	} else {
		bodyHTML = fmt.Sprintf(`<div class="lens-badge">%s</div>%s`,
			html.EscapeString(badge),
			tab.BodyHTML)
	}

	// Escape for JS string literal (single quotes).
	jsBody := strings.ReplaceAll(bodyHTML, `\`, `\\`)
	jsBody = strings.ReplaceAll(jsBody, `'`, `\'`)
	jsBody = strings.ReplaceAll(jsBody, "\n", `\n`)
	jsBody = strings.ReplaceAll(jsBody, "\r", `\r`)

	jsTitle := strings.ReplaceAll(html.EscapeString(tab.Title), `'`, `\'`)

	return fmt.Sprintf("vbAddTab('%s','%s','%s',%t);",
		html.EscapeString(id),
		jsTitle,
		jsBody,
		tab.IsMermaid)
}
