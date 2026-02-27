# vb — Phase 01 · Core Infrastructure

**Engineering Blueprint**

CLI Skeleton · Vault Discovery · Index Engine
Deterministic Infrastructure Layer

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

```
vb engineering blueprint
phase 01 · deterministic infrastructure
```