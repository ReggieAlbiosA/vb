# vb — Roadmap

**Architecture First**

Layered Execution Plan · Deterministic CLI Core

---

# Phase 01 · Core Infrastructure

**Foundation Layer**
CLI Skeleton · Vault Discovery · Index Engine

## Steps

01. **Project Bootstrap**
Setup Go module, Cobra root command, Viper config wiring.

02. **`vb init`**
Create `.vb/` directory and default `config.toml` with:

```toml
knowledge_path = "."
```

By default, topics live alongside `.vb/`,
but `knowledge_path` can point anywhere (decouples vault marker from topic storage).

03. **Upward Vault Resolver**
Two-stage resolution:

- Walk parent directories to find `.vb/` marker.
- Read `knowledge_path` from `config.toml` to locate actual topic folders
(may differ from vault root).

04. **Index Engine**
Generate and maintain `index.json` for deterministic topic resolution.

## Validation

✔ init vault → reindex → confirm deterministic topic lookup.

---

# Phase 02 · Resolution Engine

**Deterministic Mapping**
Topic → Lens → File Binding

## Steps

01. **Topic Resolver**
Use `index.json` for fast folder mapping.
Index is scoped against `knowledge_path`, not vault root.

02. **Flag Mapping**

- `--why` → `WHY.md`
- `--arch` → `ARCH.md`
- `--cli-tools` → `CLI_TOOLS.md`
- etc.

03. **Strict File Validation**
Clear errors if lens file missing.

## Validation

✔ Query existing + non-existing lens files.

---

# Phase 03 · Rendering Layer

**Output Abstraction**
Markdown Parsing & Styled Terminal Output

## Steps

01. **Goldmark Parser**
Parse vault markdown into AST.

02. **Glamour Render**
Styled terminal markdown output.

03. **Lip Gloss Themes**
Lens-specific visual identity.

## Validation

✔ Styled output differs per lens while resolution logic remains unchanged.

---

# Phase 04 · Write Loop

**Authoring System**
Editor Integration & Command Logging

## Steps

01. **`vb edit`**
Launch `$EDITOR` via `os/exec`.

02. **`--used` Engine**
Append timestamped command entries to `USED.md`.

03. **Tag Search**
Cross-topic tag lookup.

## Validation

✔ Log commands → re-query → search by tag.

---

# Phase 05 · Schema Enforcement

**Quality Control**
Lens-Aware Markdown Linter

## Steps

01. **`vb lint`**
AST inspection via Goldmark.

02. **Lens Rules**
Enforce schema per file type (`WHY.md`, `CLI_TOOLS.md`, etc.).

03. **Auto-Lint Hook**
Optional lint on save.

## Validation

✔ Break file intentionally → structured error shown.

---

# Phase 06 · GUI Modifier

**Optional Visual Layer**
Webview & Mermaid Rendering

## Steps

01. **Wails/Lorca Bridge**
Isolated GUI adapter.

02. **`--gui` Router**
Divert output from terminal to webview.

03. **Mermaid Loader**
Render `ARCH.mmd` dynamically.

## Validation

✔ GUI renders correctly.
✔ No-display environments gracefully fall back to terminal.

---

# Phase 07 · Distribution

**Ship Layer**
Completions · Binaries · Release Automation

## Steps

01. **Shell Completion**
Generate bash/zsh/fish/PowerShell scripts.

02. **GoReleaser**
Cross-platform builds.

03. **Auto Docs**
Generate usage documentation from Cobra.

## Validation

✔ Snapshot build produces working binaries.

---

vb · layered architecture roadmap
v1.1 · deterministic core