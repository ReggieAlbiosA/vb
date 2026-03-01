# vb — Phase 01 · Core Infrastructure

**Engineering Blueprint**

CLI Skeleton · Vault Discovery · Index Engine
Deterministic Infrastructure Layer

---

## Build Summary

### What
Phase 01 establishes the deterministic core that every subsequent phase depends on. It has no query commands, no rendering, no linting — just the three primitives the entire tool is built on:

- **`vb init`** — drops `.vb/` into a user-chosen directory, marking it as a vault and writing the default config
- **Vault Resolver** — a two-stage algorithm that walks parent directories to find `.vb/`, then reads `knowledge_path` from config to locate topic folders (which may live elsewhere entirely)
- **Index Engine** — walks `TopicRoot`, detects topic folders by `.md` file presence, and writes `index.json` — the lookup table that replaces all filesystem scanning in future phases

### Why
The vault can live anywhere on the machine. The topics can live somewhere else entirely. Nothing can work correctly — not Phase 02's resolution, not Phase 03's rendering, not `vb disk --why` — without a mechanism that knows both where `.vb/` is **and** where the topics actually are. Phase 01 is that mechanism.

The index exists so every future query is O(1) — a map lookup — not a directory scan. Topic resolution never touches the filesystem after Phase 01 runs `vb reindex`.

### Importance
Without Phase 01, every other phase is unrunnable. `vb disk --why` has no way to find `hardware/disk/WHY.md` without the resolver returning `TopicRoot` and the index returning `hardware/disk`. This phase is not a foundation in the abstract sense — it is the literal dependency of every command `vb` will ever expose.

If the vault resolver returns the wrong root, or the index walks the wrong directory, everything built on top silently produces wrong results. Phase 01 is the only phase where a bug propagates to 100% of future user interactions.

---

## Execution Order (Strict Sequence)

### 01 · Project Bootstrap

1. **Create Go module**

```bash
go mod init github.com/yourname/vb
```

2. **Install Cobra + Viper**

```bash
go get github.com/spf13/cobra
go get github.com/spf13/viper
```

3. **Create directory structure**

```
/cmd
root.go
init.go

/internal
config/
vault/
index/
```

---

## 02 · `vb init` Command

> **Context:** `.vb/` is a user-land runtime artifact — created by the user in their chosen knowledge base directory.
> It has no relation to the vb source code tree above.
>
> The compiled `vb` binary lives in the user’s PATH (e.g. `/usr/local/bin/vb`).
> The vault lives wherever the user runs `vb init`.

### Responsibilities

- Check if `.vb/` already exists
- If not:
- Create `.vb/`
- Create `.vb/config.toml`
- Create `.vb/index.json` (empty scaffold)
- Print confirmation

---

### Default `config.toml`

```toml
knowledge_path = "."
editor = "nano"
theme = "dark"
lint_on_save = false

# knowledge_path = "." means topic folders live alongside .vb/
# can be set to any absolute or relative path to decouple storage
```

---

## 03 · Upward Vault Resolver

### Stage 1 — Find Vault Marker

```text
start_dir := current working directory

loop:
if ".vb/" exists in start_dir: break
if start_dir == filesystem root:
error: vault not initialized
start_dir = parent(start_dir)
```

---

### Stage 2 — Resolve Knowledge Path

```text
cfg := parse(start_dir + "/.vb/config.toml")
knowledge_path := resolve(start_dir, cfg.knowledge_path)

# this may differ from start_dir if user set
# an absolute or relative path

return:
vault_root: start_dir
topic_root: knowledge_path
```

**Constraints**

- Stage 1 must not depend on index layer
- Stage 2 reads config only — no filesystem walk yet

---

## 04 · Index Engine

### Purpose

Generate deterministic topic lookup.

### `index.json` Schema

```json
{
"topics": {
"disk": "hardware/disk",
"ssh": "networking/ssh"
}
}
```

---

### Execution Order

1. Resolve `knowledge_path` via Vault Resolver (Stage 2)
2. Walk `knowledge_path` directory ← **NOT** the vault root
3. Detect topic folders (leaf nodes with any `.md` files)
4. Map topic name → path relative to `knowledge_path`
5. Serialize to `.vb/index.json`

```
# The index path values are relative to knowledge_path
# so the index remains valid even if the vault root
# directory is moved.
```

---

## Validation Checklist

- ✔ Run `vb init`
- ✔ Confirm `.vb/config.toml` exists with `knowledge_path = "."`
- ✔ Confirm `.vb/index.json` exists (empty scaffold)
- ✔ Create `hardware/disk/` manually
- ✔ Run `vb reindex`
- ✔ Confirm `disk` appears in `index.json`
- ✔ Confirm upward resolution works from nested directories
- ✔ Set `knowledge_path` to an external directory → reindex → confirm topics resolve from new path

---

## Test Coverage Requirements

Thresholds: **≥90% on internal packages · ≥80% on cmd**.

### internal/config — target ≥90%

| Test | Covers |
|---|---|
| `TestDefault` | Default config values are correct |
| `TestDefaultTOML_ContainsKeys` | Generated TOML contains all required keys |
| `TestLoad_ReadsFromFile` | Custom TOML values load correctly |
| `TestLoad_MissingFile_ReturnsDefaults` | Missing config.toml silently returns defaults |

### internal/vault — target ≥90%

| Test | Covers |
|---|---|
| `TestResolve_FromVaultRoot` | Resolves when CWD is vault root |
| `TestResolve_FromNestedDir` | Stage 1 walks upward from nested subdirectory |
| `TestResolve_NoVault` | Returns `ErrNoVault` when no `.vb/` exists |
| `TestResolve_AbsoluteKnowledgePath` | Absolute `knowledge_path` respected |
| `TestResolve_RelativeKnowledgePath` | Relative `knowledge_path` resolved against vault root |

### internal/index — target ≥90%

| Test | Covers |
|---|---|
| `TestBuild_Empty` | No topics → empty index |
| `TestBuild_WithTopics` | Directories with `.md` files are indexed |
| `TestBuild_SkipsVbDir` | `.vb/` is never indexed |
| `TestBuild_IgnoresDirsWithoutMD` | Category dirs without direct `.md` files skipped |
| `TestBuild_DecoupledTopicRoot` | Index built from separate `topicRoot` |
| `TestBuild_WritesIndexJSON` | `index.json` written to `vaultRoot/.vb/` |
| `TestBuild_NonExistentTopicRoot` | Error on non-existent `topicRoot` |
| `TestBuild_WriteError` | Error when `index.json` cannot be written |
| `TestLoad_RoundTrip` | Build → Load returns same schema |
| `TestLoad_FileNotFound` | Error when `index.json` missing |
| `TestLoad_MalformedJSON` | Error on invalid JSON with correct message |

### cmd — target ≥80%

| Test | Covers |
|---|---|
| `TestInitCmd_Success` | Creates `.vb/`, `config.toml`, `index.json` |
| `TestInitCmd_AlreadyInitialized` | Errors on double-init |
| `TestReindexCmd_Success` | Indexes topics, writes `index.json` |
| `TestReindexCmd_NoVault` | Errors when run outside any vault |

---

```
vb engineering blueprint
phase 01 · deterministic infrastructure
```