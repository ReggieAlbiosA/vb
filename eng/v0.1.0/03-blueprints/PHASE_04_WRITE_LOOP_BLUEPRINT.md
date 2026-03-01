# vb — Phase 04 · Write Loop

**Engineering Blueprint**

Authoring System · Editor Integration · Command Logging · Tag Search  
Vault-Aware Write Layer

---

## Build Summary

### What
Phase 04 is the write side of `vb`. Every previous phase has been read-only — Phase 04 adds the three authoring primitives:

- **`vb edit`** — launches `$EDITOR` (from vault config) on a resolved lens file via `os/exec`, creating the file first if it doesn't exist
- **`--used` engine** — appends a timestamped, structured entry to `USED.md` every time a command is logged, building a per-topic usage history
- **Tag search** — a cross-topic index over `#tag` occurrences in vault markdown, exposed via `vb tag <name>`

### Why
Phase 01–03 tell you what you know. Phase 04 is how you write it down and find it again. Without `vb edit`, adding content requires knowing the exact file path in the vault. Without `--used`, command history lives only in shell history, not in the vault. Without tag search, cross-topic associations are invisible.

The `vb edit` command must resolve the topic and lens through the existing Phase 02 resolver — it is **not** a raw filepath opener. This ensures topics can only be edited through valid vault paths, preventing files from being created outside `TopicRoot`.

### Importance
Phase 04 is the first phase that **writes** to the vault. Correctness constraints are stricter than the read side:

- A bug in `vb edit` could create files at the wrong path, silently fragmenting the vault
- The `--used` logger must be append-only — overwriting `USED.md` would destroy history
- Tag indexing must scan only `TopicRoot`, never the `VaultRoot` itself (to avoid indexing `.vb/` internals)

The `--used` flag also introduces the first shared "after-write" pattern that Phase 05 (lint-on-save) will call into — the hook boundary is defined here.

---

## Package Structure

```
/internal
  /editor
    open.go       ← launch $EDITOR via os/exec, create file if missing
  /logger
    used.go       ← append timestamped entry to USED.md
  /tagger
    scan.go       ← walk TopicRoot, extract #tag occurrences → tag index
    search.go     ← cross-topic tag lookup
```

> Do NOT import `internal/render` or `internal/resolver` from `internal/editor`.  
> `editor.Open()` receives an already-resolved `filePath` from the cmd layer.  
> The cmd layer (Phase 02) is responsible for resolution — the editor only opens files.  
> `internal/logger` and `internal/tagger` are independent packages with no cross-dependency.

---

## Command Wiring Overview

```
vb edit <topic> --<lens>         →  cmd/edit.go
vb <topic> --<lens> --used       →  cmd/query.go (--used flag added here)
vb tag <name>                    →  cmd/tag.go
```

---

## 00 · CLI Command Wiring

### cmd/edit.go

```go
var editCmd = &cobra.Command{
    Use:   "edit <topic>",
    Short: "Open a lens file in $EDITOR",
    Args:  cobra.ExactArgs(1),
    RunE:  runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
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

    filePath := filepath.Join(topicDir, lensFile)

    cfg, err := config.Load(ctx.VaultRoot)
    if err != nil {
        return err
    }

    return editor.Open(filePath, cfg.Editor)
}
```

### cmd/tag.go

```go
var tagCmd = &cobra.Command{
    Use:   "tag <name>",
    Short: "Search for a tag across all topics",
    Args:  cobra.ExactArgs(1),
    RunE:  runTag,
}

func runTag(cmd *cobra.Command, args []string) error {
    tagName := args[0]

    ctx, err := vault.Resolve(cwd)
    if err != nil {
        return err
    }

    results, err := tagger.Search(ctx.TopicRoot, tagName)
    if err != nil {
        return err
    }

    if len(results) == 0 {
        fmt.Printf("no topics tagged #%s\n", tagName)
        return nil
    }

    for _, r := range results {
        fmt.Printf("  %s  (%s)\n", r.Topic, r.File)
    }
    return nil
}
```

> `--used` is added as a persistent flag on `queryCmd` in Phase 02's `cmd/query.go`.  
> When set, `runQuery` calls `logger.Append(ctx.VaultRoot, topicDir, topic, lens)` after a successful render.

---

## 01 · Editor Integration — internal/editor/open.go

```go
// Open launches the configured editor on filePath.
// If filePath does not exist, it is created (with an empty scaffold) before opening.
// editor is read from vault config (cfg.Editor); falls back to $EDITOR env var if blank.
func Open(filePath string, editor string) error {
    if err := ensureFile(filePath); err != nil {
        return fmt.Errorf("creating %s: %w", filePath, err)
    }

    ed := resolveEditor(editor)
    if ed == "" {
        return errors.New("no editor configured — set 'editor' in .vb/config.toml or $EDITOR")
    }

    cmd := exec.Command(ed, filePath)
    cmd.Stdin  = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

// ensureFile creates filePath (and parent dirs) if it does not already exist.
func ensureFile(filePath string) error {
    if _, err := os.Stat(filePath); err == nil {
        return nil // already exists
    }
    if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
        return err
    }
    return os.WriteFile(filePath, []byte{}, 0o644)
}

// resolveEditor returns the editor to use: config value first, then $EDITOR env var.
func resolveEditor(cfgEditor string) string {
    if cfgEditor != "" {
        return cfgEditor
    }
    return os.Getenv("EDITOR")
}
```

**Error Cases**
- `editor` blank and `$EDITOR` unset → explicit error, no silent fallback to `vi`
- `filePath` parent dirs missing → `os.MkdirAll` creates them (supports new topic creation)
- Editor exits non-zero → error propagated as-is to caller

---

## 02 · `--used` Logger — internal/logger/used.go

```go
// Entry is a single timestamped usage log line.
type Entry struct {
    Timestamp time.Time
    Topic     string
    Lens      string
}

// Append writes one Entry to USED.md inside the topic directory.
// File is created if it does not exist. Existing content is never overwritten.
func Append(topicDir, topic, lens string) error {
    usedPath := filepath.Join(topicDir, "USED.md")

    entry := formatEntry(Entry{
        Timestamp: time.Now().UTC(),
        Topic:     topic,
        Lens:      lens,
    })

    f, err := os.OpenFile(usedPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
    if err != nil {
        return fmt.Errorf("opening USED.md: %w", err)
    }
    defer f.Close()

    _, err = fmt.Fprintln(f, entry)
    return err
}

// formatEntry produces a single, human-readable log line.
func formatEntry(e Entry) string {
    return fmt.Sprintf("- %s  vb %s --%s",
        e.Timestamp.Format("2006-01-02 15:04 UTC"),
        e.Topic,
        e.Lens,
    )
}
```

**Constraints**
- `os.O_APPEND` is non-negotiable — never truncate `USED.md`
- Timestamp format is fixed: `2006-01-02 15:04 UTC` (Go reference time)
- Log lines are plain Markdown list items so `USED.md` is valid vault content

**Example `USED.md` after three queries:**

```markdown
- 2026-02-27 14:01 UTC  vb disk --why
- 2026-02-27 15:22 UTC  vb disk --arch
- 2026-02-28 09:04 UTC  vb ssh --cli-tools
```

---

## 03 · Tag Search — internal/tagger/scan.go

```go
// TagResult is a single tag hit in the vault.
type TagResult struct {
    Topic string
    File  string
    Tag   string
}

// Search walks topicRoot and returns all TagResults matching tagName.
// Tag format: #tagname (word-boundary match, case-insensitive).
func Search(topicRoot, tagName string) ([]TagResult, error) {
    pattern := regexp.MustCompile(`(?i)#` + regexp.QuoteMeta(tagName) + `\b`)

    var results []TagResult

    err := filepath.WalkDir(topicRoot, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() || filepath.Ext(path) != ".md" {
            return nil
        }

        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        if pattern.Match(content) {
            rel, _ := filepath.Rel(topicRoot, path)
            parts  := strings.SplitN(rel, string(os.PathSeparator), 2)
            topic  := parts[0]
            results = append(results, TagResult{
                Topic: topic,
                File:  filepath.Base(path),
                Tag:   tagName,
            })
        }
        return nil
    })

    return results, err
}
```

**Constraints**
- Walk `TopicRoot` only — never the vault root (avoids scanning `.vb/` internals)
- Tag pattern uses `\b` word boundary — `#disk` does not match `#disks`
- Case-insensitive match — `#SSH` matches `#ssh`
- Only `.md` files are scanned — binary files and config files are skipped

---

## Validation Checklist

- ✔ `vb edit disk --why` with existing topic + existing lens → opens editor on `WHY.md`
- ✔ `vb edit disk --why` with existing topic + missing `WHY.md` → creates file, then opens editor
- ✔ `vb edit disk --why` on topic not in index → `ErrTopicNotFound` (must `vb reindex` first)
- ✔ Editor exits 0 → `vb edit` exits 0 with no output
- ✔ `cfg.Editor` blank + `$EDITOR` unset → explicit error message
- ✔ `vb disk --why --used` → renders output + appends entry to `disk/USED.md`
- ✔ `vb disk --why --used` run twice → two entries in `USED.md`, original entry preserved
- ✔ `vb tag ssh` → lists all topics containing `#ssh`
- ✔ `vb tag unknown` → prints "no topics tagged #unknown", exits 0
- ✔ Tag scan does not read `.vb/` directory contents

---

## Test Coverage Requirements

Thresholds: **≥90% on internal packages · ≥80% on cmd**

### internal/editor/open_test.go

| Test | Covers |
|---|---|
| `TestOpen_ExistingFile` | Editor invoked on existing file without re-creating it |
| `TestOpen_CreatesFile` | Missing lens file is created before editor launch |
| `TestOpen_CreatesMissingDirs` | Parent directories created for new topic |
| `TestOpen_NoEditor` | Blank config + blank `$EDITOR` → explicit error |
| `TestResolveEditor_ConfigFirst` | Config value takes priority over `$EDITOR` |
| `TestResolveEditor_EnvFallback` | Falls back to `$EDITOR` when config is blank |

### internal/logger/used_test.go

| Test | Covers |
|---|---|
| `TestAppend_CreatesFile` | `USED.md` created on first log call |
| `TestAppend_AppendsNotOverwrites` | Second call appends; first entry preserved |
| `TestAppend_EntryFormat` | Output line matches `- YYYY-MM-DD HH:MM UTC  vb topic --lens` |
| `TestFormatEntry_Timestamp` | Timestamp format is fixed to UTC reference layout |
| `TestAppend_WriteError` | Propagates error if `USED.md` cannot be opened |

### internal/tagger/scan_test.go

| Test | Covers |
|---|---|
| `TestSearch_Found` | Topic with matching tag returned |
| `TestSearch_NotFound` | No match → empty slice, no error |
| `TestSearch_CaseInsensitive` | `#SSH` matches `#ssh` |
| `TestSearch_WordBoundary` | `#disk` does not match `#disks` |
| `TestSearch_MultipleTopics` | Tag in two topics → two results |
| `TestSearch_SkipsNonMD` | Binary or config files not scanned |
| `TestSearch_SkipsVbDir` | `.vb/` contents never appear in results |

### cmd/edit_test.go

| Test | Covers |
|---|---|
| `TestEditCmd_Success` | `vb edit disk --why` → no error, editor called |
| `TestEditCmd_TopicNotFound` | Unknown topic → `ErrTopicNotFound` from resolver |
| `TestEditCmd_NoLens` | No lens flag → `ErrNoLens` |
| `TestEditCmd_CreatesLensFile` | Missing lens file created before editor launch |

### cmd/tag_test.go

| Test | Covers |
|---|---|
| `TestTagCmd_Found` | Known tag → results printed |
| `TestTagCmd_NotFound` | Unknown tag → "no topics tagged" message, exits 0 |
| `TestTagCmd_NoArg` | No argument → cobra argument validation error |

> `--used` flag coverage is included in `cmd/query_test.go` — add:  
> `TestQueryCmd_UsedFlag`: `vb disk --why --used` → entry appended to `USED.md`.

---

```
vb engineering blueprint
phase 04 · write loop
```
