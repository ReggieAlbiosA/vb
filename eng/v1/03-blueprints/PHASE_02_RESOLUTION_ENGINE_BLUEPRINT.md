# vb — Phase 02 · Resolution Engine

**Engineering Blueprint**

Deterministic Mapping · Topic → Lens → File Binding  
Zero Magic · Zero Ambiguity

---

## Build Summary

### What
Phase 02 is the address system of `vb`. It takes a topic name and a lens flag from the CLI and resolves them to an absolute file path — using nothing but the index and a flag-to-filename map. No filesystem scanning. No guessing. No fallback.

Three pure functions, each with a single job:
- **Topic Resolver** — looks up the topic in `index.json` → returns absolute topic directory path
- **Lens Resolver** — maps a CLI flag (`--why`) → a filename (`WHY.md`)
- **Binder** — joins the two, stats the file, returns the path or a user-readable error

### Why
`vb disk --why` means nothing to the OS. Phase 02 is the translation layer — it converts that human intent into `{TopicRoot}/hardware/disk/WHY.md`. Without it, Phase 01's index is data with no consumer and Phase 03's renderer has no file to render.

The resolution is intentionally strict and explicit: one topic, one lens, one file. No inference, no fuzzy matching, no auto-creation. If the file doesn't exist, the user gets a clear message telling them exactly what to run next.

### Importance
Phase 02 is the contract between the user's intent and the filesystem. Every query command `vb` will ever expose — in v1.0 and beyond — flows through this resolution chain. Getting it right means:

- Wrong topic → explicit error, not a wrong file
- Missing lens file → actionable error with `vb edit` hint, not a silent empty output
- `--gui` modifier → not confused with a lens, passed cleanly to the renderer

The `topicRoot` fix (using `TopicRoot` not `VaultRoot` from Phase 01's `VaultContext`) is the most critical correctness constraint in this phase. It is what makes decoupled `knowledge_path` vaults work end-to-end.

---

## Package Structure


```
/internal
  /resolver
    topic.go      ← topic lookup against index
    lens.go       ← flag → filename mapping + --gui exclusion
    binding.go    ← file existence validation
    errors.go     ← sentinel errors (owned here, not in a separate package)
```

> `/internal/index/` already exists from Phase 01 — reuse `index.Load()` directly.  
> Do NOT create a new `/index/loader.go`. Import the existing package.  
> Resolver layer must not depend on renderer or linter. Pure resolution logic only.

---

## Resolution Flow

```
CLI Input:
  vb disk --why

Execution Order:
  1. Vault resolution (Phase 01) → VaultContext{VaultRoot, TopicRoot}
  2. Load index.json via existing index.Load(VaultRoot)
  3. Resolve topic via index map → absolute topic directory path
  4. Resolve lens flag → filename (--gui excluded here, passed to renderer)
  5. Construct absolute file path = topicDir + lensFile
  6. Validate file existence (stat only — no scan)
  7. Return resolved file path to cmd layer
```

---

## 00 · CLI Command Wiring — cmd/query.go

Phase 02 must expose the resolution as a Cobra command so `vb disk --why` works.

```go
// cmd/query.go
var (
    flagWhy       bool
    flagImportance bool
    flagCLITools  bool
    flagArch      bool
    flagUsed      bool
    flagGotchas   bool
    flagRefs      bool
    flagGUI       bool  // modifier — not a lens, passed to renderer in Phase 03
)

var queryCmd = &cobra.Command{
    Use:   "<topic>",
    Short: "Query a topic by lens",
    Args:  cobra.ExactArgs(1),
    RunE:  runQuery,
}

func runQuery(cmd *cobra.Command, args []string) error {
    topic := args[0]

    // Determine which lens flag was set.
    lens, err := resolver.ActiveLens(cmd.Flags())
    if err != nil {
        return err
    }

    // Two-stage vault resolution (Phase 01).
    ctx, err := vault.Resolve(cwd)

    // Load index from existing package.
    schema, err := index.Load(ctx.VaultRoot)

    // Resolve topic → topicDir (using TopicRoot, not VaultRoot).
    topicDir, err := resolver.ResolveTopic(topic, schema, ctx.TopicRoot)

    // Resolve lens flag → filename.
    lensFile, err := resolver.ResolveLens(lens)

    // Validate file exists.
    filePath, err := resolver.Bind(topicDir, lensFile)

    // Hand off to renderer (Phase 03), passing flagGUI modifier.
    return render.File(filePath, flagGUI)
}
```

---

## 01 · Topic Resolver — internal/resolver/topic.go

```go
// ResolveTopic looks up topic in the loaded index and returns
// the absolute path to the topic directory.
//
// Note: paths in index.json are relative to TopicRoot (not VaultRoot).
// TopicRoot comes from VaultContext resolved in Phase 01.
func ResolveTopic(topic string, schema index.Schema, topicRoot string) (string, error) {
    relPath, exists := schema.Topics[topic]
    if !exists {
        return "", fmt.Errorf("%w: %q", ErrTopicNotFound, topic)
    }
    return filepath.Join(topicRoot, relPath), nil  // ← TopicRoot, NOT VaultRoot
}
```

**Error Cases**
- Topic not found in index
- `index.json` missing (handled by `index.Load()`)
- `index.json` empty/corrupted (handled by `index.Load()`)

---

## 02 · Lens Mapping — internal/resolver/lens.go

```go
// LensToFile maps CLI flag names to their corresponding vault filenames.
// --gui is intentionally absent — it is a rendering modifier, not a lens.
// It is parsed in the cmd layer and passed to the renderer in Phase 03/06.
var LensToFile = map[string]string{
    "why":        "WHY.md",
    "importance": "IMPORTANCE.md",
    "cli-tools":  "CLI_TOOLS.md",
    "arch":       "ARCH.md",
    "used":       "USED.md",
    "gotchas":    "GOTCHAS.md",
    "refs":       "REFS.md",
}

// ResolveLens converts a CLI flag name to its vault filename.
func ResolveLens(flag string) (string, error) {
    file, exists := LensToFile[flag]
    if !exists {
        return "", fmt.Errorf("%w: --%s", ErrInvalidLens, flag)
    }
    return file, nil
}

// ActiveLens inspects the flagset and returns the single active lens name.
// Errors if zero or multiple lenses are set.
func ActiveLens(flags *pflag.FlagSet) (string, error) {
    var active []string
    for name := range LensToFile {
        f := flags.Lookup(name)
        if f != nil && f.Changed {
            active = append(active, name)
        }
    }
    if len(active) == 0 {
        return "", ErrNoLens
    }
    if len(active) > 1 {
        return "", fmt.Errorf("%w: %v", ErrMultipleLenses, active)
    }
    return active[0], nil
}
```

**Error Cases**
- No lens flag passed → `ErrNoLens`
- Multiple lens flags passed → `ErrMultipleLenses`
- Unknown lens flag → `ErrInvalidLens`

---

## 03 · Strict File Validation — internal/resolver/binding.go

```go
// Bind constructs the absolute file path and verifies it exists.
// Returns a user-readable error if the lens file hasn't been authored yet.
func Bind(topicDir string, lensFile string) (string, error) {
    fullPath := filepath.Join(topicDir, lensFile)

    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        topic := filepath.Base(topicDir)
        return "", fmt.Errorf("%w: %q not authored yet for topic %q — run: vb edit %s --%s",
            ErrLensFileMissing, lensFile, topic, topic, lensNameFor(lensFile))
    }
    return fullPath, nil
}
```

**Error Case — user-readable message:**
```
Error: "WHY.md" not authored yet for topic "disk" — run: vb edit disk --why
```

---

## Error Model — internal/resolver/errors.go

Errors live in `internal/resolver/` — not in a separate top-level `/errors` package.

```go
var (
    ErrTopicNotFound  = errors.New("topic not found in index")
    ErrInvalidLens    = errors.New("invalid lens flag")
    ErrNoLens         = errors.New("no lens flag provided")
    ErrMultipleLenses = errors.New("only one lens flag may be used at a time")
    ErrLensFileMissing = errors.New("lens file not authored yet")
)
```

> Errors must be explicit and user-readable.  
> No silent fallback. No auto-creation.

---

## Validation Checklist

- ✔ `vb disk --why` with existing topic + existing lens file → returns file path
- ✔ `vb disk --why` with existing topic + missing lens file → user-readable error with `vb edit` hint
- ✔ `vb unknown --why` → `ErrTopicNotFound`
- ✔ `vb disk --invalid-flag` → `ErrInvalidLens`
- ✔ `vb disk` (no lens) → `ErrNoLens`
- ✔ `vb disk --why --arch` (two lenses) → `ErrMultipleLenses`
- ✔ `vb disk --why --gui` → resolves file path, passes `gui=true` to renderer
- ✔ `index.json` empty `{}` → `ErrTopicNotFound` (not a crash)
- ✔ Confirm no filesystem scan occurs — resolution uses index only

---

## Test Coverage Requirements

Phase 02 introduces three new testable files in `internal/resolver/` and one new cmd file.  
Coverage thresholds follow Phase 01 standards: **≥90% on internal packages, ≥80% on cmd**.

### internal/resolver/topic_test.go

| Test | Covers |
|---|---|
| `TestResolveTopic_Found` | Known topic → returns correct absolute path using `topicRoot` |
| `TestResolveTopic_NotFound` | Unknown topic → `ErrTopicNotFound` |
| `TestResolveTopic_EmptyIndex` | Empty `{}` index → `ErrTopicNotFound`, no crash |
| `TestResolveTopic_UsesTopicRoot` | Path joins against `topicRoot`, not `vaultRoot` |

### internal/resolver/lens_test.go

| Test | Covers |
|---|---|
| `TestResolveLens_AllValidFlags` | All 7 flags map to correct filenames |
| `TestResolveLens_InvalidFlag` | Unknown flag → `ErrInvalidLens` |
| `TestActiveLens_SingleFlag` | Exactly one lens set → returns its name |
| `TestActiveLens_NoFlag` | No lens set → `ErrNoLens` |
| `TestActiveLens_MultipleFlags` | Two lenses set → `ErrMultipleLenses` |
| `TestActiveLens_GUINotALens` | `--gui` cannot be returned as a lens |

### internal/resolver/binding_test.go

| Test | Covers |
|---|---|
| `TestBind_FileExists` | Valid topicDir + lensFile → returns full path |
| `TestBind_FileMissing` | Missing lens file → `ErrLensFileMissing` with user-readable message |
| `TestBind_ErrorMessageContainsTopic` | Error message includes topic name and `vb edit` hint |

### cmd/query_test.go

| Test | Covers |
|---|---|
| `TestQueryCmd_Success` | `vb disk --why` with authored file → no error |
| `TestQueryCmd_TopicNotFound` | `vb unknown --why` → error |
| `TestQueryCmd_LensFileMissing` | `vb disk --why` with no `WHY.md` → user-readable error |
| `TestQueryCmd_NoLens` | `vb disk` (no flag) → error |
| `TestQueryCmd_GUIModifier` | `vb disk --why --gui` → resolves without error |

> **Note:** `internal/resolver/errors.go` has 0% directly testable statements  
> (only `var` declarations). Coverage is satisfied through the above tests exercising the error paths.

---

```
vb engineering blueprint
phase 02 · deterministic resolution
```