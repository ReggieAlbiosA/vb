# vb — Phase 05 · Schema Enforcement

**Engineering Blueprint**

Quality Control · Lens-Aware Markdown Linter · AST Inspection  
Vault Schema Validation Layer

---

## Build Summary

### What
Phase 05 is the contract enforcer of `vb`. It inspects vault Markdown files against per-lens schemas using the Goldmark AST — catching structural violations before they silently corrupt the vault.

Three components:
- **`vb lint`** — resolves a topic + lens to a file path, then runs the appropriate lens rule against its parsed AST
- **Lens Rules** — a map of lens name → `LensRule` interface, each enforcing the expected structure for that file type (`WHY.md`, `CLI_TOOLS.md`, etc.)
- **Auto-lint hook** — a post-write function called by Phase 04's `editor.Open()` when `lint_on_save = true` in vault config

### Why
The vault is only reliable if its files respect the schema that `vb` and its users expect. A `WHY.md` that contains no body text, a `CLI_TOOLS.md` that contains no code blocks, a `REFS.md` with no links — these are structurally broken files that query correctly (Phase 02 resolves them) but are content-empty.

Phase 05 is the only phase that validates content, not path. The Goldmark AST is the right tool because it gives structural visibility into headings, paragraphs, code fences, and lists — without rebuilding a Markdown parser.

### Importance
`vb lint` is optional at query time but becomes critical if `lint_on_save = true`. When enabled, every write via `vb edit` will immediately surface schema violations — making the vault self-correcting during authoring.

The linter must **never block reads**. Lint results are advisory: they produce structured output and a non-zero exit code, but they do not modify files and do not prevent `vb disk --why` from succeeding. The linter is a quality gate, not a wall.

---

## Package Structure

```
/internal
  /linter
    linter.go     ← Lint() entry point, rule dispatch
    rules.go      ← LensRule interface + per-lens implementations
    errors.go     ← LintError type (file, lens, message)
  /hook
    save.go       ← post-write hook called by editor.Open() when lint_on_save = true
```

> Goldmark is already a transitive dependency via Glamour (Phase 03). Import it directly.  
> `internal/linter` must NOT import `internal/render`. Parse ≠ render.  
> `internal/hook` imports `internal/linter` only — no resolver, no editor dependency loop.  
> Lint never modifies files. `os.ReadFile` + AST walk only.

---

## 00 · CLI Command Wiring — cmd/lint.go

```go
var lintCmd = &cobra.Command{
    Use:   "lint <topic>",
    Short: "Validate a lens file against its schema",
    Args:  cobra.ExactArgs(1),
    RunE:  runLint,
}

func runLint(cmd *cobra.Command, args []string) error {
    topic := args[0]

    lens, err := resolver.ActiveLens(cmd.Flags())
    if err != nil {
        return err
    }

    ctx, err := vault.Resolve(cwd)
    if err != nil {
        return err
    }

    schema, err := index.Load(ctx.VaultRoot)
    if err != nil {
        return err
    }

    topicDir, err := resolver.ResolveTopic(topic, schema, ctx.TopicRoot)
    if err != nil {
        return err
    }

    lensFile, err := resolver.ResolveLens(lens)
    if err != nil {
        return err
    }

    filePath, err := resolver.Bind(topicDir, lensFile)
    if err != nil {
        return err
    }

    lintErrs, err := linter.Lint(filePath, lens)
    if err != nil {
        return err
    }

    if len(lintErrs) == 0 {
        fmt.Printf("✔ %s: schema valid\n", lens)
        return nil
    }

    for _, e := range lintErrs {
        fmt.Fprintf(os.Stderr, "✘ [%s] %s\n", e.Lens, e.Message)
    }
    return fmt.Errorf("%d schema violation(s) found", len(lintErrs))
}
```

> The `--gui` flag is **not** available on `lint` — lint is terminal-only output.  
> Lint exit code is non-zero when violations exist. CI pipelines can gate on this.

---

## 01 · Linter Entry Point — internal/linter/linter.go

```go
// Lint parses the Markdown file at filePath and runs the rule for the given lens.
// Returns a slice of LintErrors (empty slice = valid). Never modifies the file.
func Lint(filePath, lens string) ([]LintError, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("reading %s: %w", filePath, err)
    }

    rule, ok := LensRules[lens]
    if !ok {
        // No rule defined for this lens — treated as always valid.
        return nil, nil
    }

    // Parse to AST using Goldmark. No rendering — structural inspection only.
    md  := goldmark.New()
    doc := md.Parser().Parse(text.NewReader(content))

    return rule.Check(doc, content), nil
}
```

> `goldmark.New()` with default options is sufficient — no extensions needed for AST walking.  
> The AST walk is read-only. No `Transform()` calls.

---

## 02 · Lens Rules — internal/linter/rules.go

```go
// LensRule is the validation contract every lens rule implements.
type LensRule interface {
    Check(doc ast.Node, source []byte) []LintError
}

// LensRules maps lens flag names to their validation rules.
// Lenses without a rule entry are treated as schema-free (always valid).
var LensRules = map[string]LensRule{
    "why":        &WhyRule{},
    "importance": &ImportanceRule{},
    "cli-tools":  &CLIToolsRule{},
    "arch":       &ArchRule{},
    "gotchas":    &GotchasRule{},
    "refs":       &RefsRule{},
    // "used" is excluded — USED.md is append-only log, not schema-validated.
}
```

### Rule Definitions

```go
// WhyRule: must have at least one non-empty paragraph.
type WhyRule struct{}
func (r *WhyRule) Check(doc ast.Node, src []byte) []LintError {
    if !hasNodeType(doc, ast.KindParagraph, src) {
        return []LintError{{Lens: "why", Message: "WHY.md must contain at least one paragraph"}}
    }
    return nil
}

// ImportanceRule: must have at least one non-empty paragraph.
type ImportanceRule struct{}
func (r *ImportanceRule) Check(doc ast.Node, src []byte) []LintError {
    if !hasNodeType(doc, ast.KindParagraph, src) {
        return []LintError{{Lens: "importance", Message: "IMPORTANCE.md must contain at least one paragraph"}}
    }
    return nil
}

// CLIToolsRule: must have at least one fenced code block.
type CLIToolsRule struct{}
func (r *CLIToolsRule) Check(doc ast.Node, src []byte) []LintError {
    if !hasFencedCodeBlock(doc) {
        return []LintError{{Lens: "cli-tools", Message: "CLI_TOOLS.md must contain at least one fenced code block"}}
    }
    return nil
}

// ArchRule: must have at least one heading.
type ArchRule struct{}
func (r *ArchRule) Check(doc ast.Node, src []byte) []LintError {
    if !hasNodeType(doc, ast.KindHeading, src) {
        return []LintError{{Lens: "arch", Message: "ARCH.md must contain at least one heading"}}
    }
    return nil
}

// GotchasRule: must have at least one list item.
type GotchasRule struct{}
func (r *GotchasRule) Check(doc ast.Node, src []byte) []LintError {
    if !hasNodeType(doc, ast.KindListItem, src) {
        return []LintError{{Lens: "gotchas", Message: "GOTCHAS.md must contain at least one list item"}}
    }
    return nil
}

// RefsRule: must have at least one link.
type RefsRule struct{}
func (r *RefsRule) Check(doc ast.Node, src []byte) []LintError {
    if !hasNodeType(doc, ast.KindLink, src) {
        return []LintError{{Lens: "refs", Message: "REFS.md must contain at least one link"}}
    }
    return nil
}
```

### AST Walk Helpers

```go
// hasNodeType returns true if the AST contains at least one node of the given kind.
func hasNodeType(doc ast.Node, kind ast.NodeKind, src []byte) bool {
    found := false
    ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
        if entering && n.Kind() == kind {
            found = true
            return ast.WalkStop, nil
        }
        return ast.WalkContinue, nil
    })
    return found
}

// hasFencedCodeBlock returns true if the AST contains at least one fenced code block.
func hasFencedCodeBlock(doc ast.Node) bool {
    found := false
    ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
        if entering && n.Kind() == ast.KindFencedCodeBlock {
            found = true
            return ast.WalkStop, nil
        }
        return ast.WalkContinue, nil
    })
    return found
}
```

---

## 03 · Error Model — internal/linter/errors.go

```go
// LintError represents a single schema violation in a lens file.
type LintError struct {
    Lens    string // The lens name (e.g. "why")
    Message string // Human-readable description of the violation
}

func (e LintError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Lens, e.Message)
}
```

---

## 04 · Auto-Lint Hook — internal/hook/save.go

```go
// OnSave is called by editor.Open() after the editor process exits.
// It runs lint only if lint_on_save = true in vault config.
// Lint errors are printed as warnings — they do not fail the edit command.
func OnSave(filePath, lens string, lintOnSave bool) {
    if !lintOnSave {
        return
    }

    lintErrs, err := linter.Lint(filePath, lens)
    if err != nil {
        fmt.Fprintf(os.Stderr, "lint: %v\n", err)
        return
    }

    if len(lintErrs) == 0 {
        fmt.Printf("✔ %s: schema valid\n", lens)
        return
    }

    for _, e := range lintErrs {
        fmt.Fprintf(os.Stderr, "⚠ [%s] %s\n", e.Lens, e.Message)
    }
}
```

> On-save lint results are **warnings** (`⚠`), not errors.  
> The edit command exits 0 even when lint violations are found on save.  
> `vb lint` (explicit) exits non-zero. On-save hook does not.

---

## Validation Checklist

- ✔ `vb lint disk --why` on a valid `WHY.md` → prints `✔ why: schema valid`, exits 0
- ✔ `vb lint disk --why` on an empty `WHY.md` → prints violation, exits non-zero
- ✔ `vb lint disk --cli-tools` on file with no code blocks → violation
- ✔ `vb lint disk --refs` on file with no links → violation
- ✔ `vb lint disk --used` → no rule defined, always valid
- ✔ `vb lint unknown --why` → `ErrTopicNotFound` (resolver error, not lint error)
- ✔ `lint_on_save = true` + `vb edit disk --why` → lint runs after editor exits, warnings printed
- ✔ `lint_on_save = false` + `vb edit disk --why` → no lint, editor exits cleanly
- ✔ On-save violations are `⚠` warnings — `vb edit` still exits 0
- ✔ Linter never modifies any file

---

## Test Coverage Requirements

Thresholds: **≥90% on internal packages · ≥80% on cmd**

### internal/linter/linter_test.go

| Test | Covers |
|---|---|
| `TestLint_ValidWhy` | Valid `WHY.md` → empty error slice |
| `TestLint_EmptyFile` | Empty file for any schema-checked lens → violation |
| `TestLint_NoRuleForLens` | `"used"` lens → `nil` errors, no panic |
| `TestLint_FileNotFound` | Non-existent path → error before AST parse |

### internal/linter/rules_test.go

| Test | Covers |
|---|---|
| `TestWhyRule_HasParagraph` | Valid WHY.md → no violation |
| `TestWhyRule_NoParagraph` | File with only a heading → violation |
| `TestCLIToolsRule_HasCodeBlock` | File with fenced block → no violation |
| `TestCLIToolsRule_NoCodeBlock` | Prose-only file → violation |
| `TestArchRule_HasHeading` | File with heading → no violation |
| `TestGotchasRule_HasListItem` | File with list → no violation |
| `TestRefsRule_HasLink` | File with `[text](url)` → no violation |
| `TestRefsRule_NoLink` | Prose-only file → violation |

### internal/hook/save_test.go

| Test | Covers |
|---|---|
| `TestOnSave_LintOnSaveFalse` | `lintOnSave=false` → linter not called |
| `TestOnSave_ValidFile` | Valid file + `lintOnSave=true` → `✔` printed |
| `TestOnSave_InvalidFile` | Invalid file + `lintOnSave=true` → `⚠` printed, no error returned |

### cmd/lint_test.go

| Test | Covers |
|---|---|
| `TestLintCmd_Valid` | Valid topic + lens → exits 0 |
| `TestLintCmd_Violation` | Schema-broken file → exits non-zero |
| `TestLintCmd_TopicNotFound` | Unknown topic → resolver error |
| `TestLintCmd_LensFileMissing` | Authored topic, no lens file → binder error |
| `TestLintCmd_NoLens` | No lens flag → `ErrNoLens` |

---

```
vb engineering blueprint
phase 05 · schema enforcement
```
