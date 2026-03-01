# Global Vault Registry — Use vb from anywhere

## Context

Previously `vb` only worked inside a vault directory tree — it walked upward from cwd looking for `.vb/`. If you were in `~/Downloads` or any directory outside a vault, every command failed. This enhancement adds a global vault registry so `vb` can resolve a default vault from any directory, plus an explicit `--vault` / `-V` flag to target any registered vault by name.

## Usage

```bash
# Create and register a new vault (auto-sets as default if first)
vb vault create sysknow ~/knowledge

# Or register during init
vb init --name sysknow

# List registered vaults (* marks default)
vb vault list
# * sysknow             /home/user/knowledge

# Switch default vault
vb vault use sysknow

# Unregister a vault (files are not deleted)
vb vault remove sysknow

# Use vb from anywhere — default vault is resolved automatically
cd ~/Downloads
vb disk --why          # resolves via default vault

# Target a specific vault explicitly
vb disk --why -V work-vault
```

## What changed

| Before | After |
|--------|-------|
| `vb` only worked inside a vault directory tree | `vb` works from any directory via default vault |
| No way to manage multiple vaults | `vb vault create/list/use/remove` |
| `vb init` was local-only | `vb init --name` also registers in global registry |
| Single resolution strategy (cwd walk) | 3-priority: `--vault` flag → cwd walk → default registry vault |

## Vault Resolution Strategy

`resolveVault()` in `cmd/resolve.go` applies 3 priorities:

1. **`--vault` / `-V` flag**: look up the name in `~/.config/vb/vaults.json`
2. **cwd walk**: walk upward from current directory looking for `.vb/` (backward-compatible)
3. **Default vault**: fall back to the default vault in the global registry

## Implementation

### `internal/registry/registry.go` (new)
- `Registry` struct: `Default` string + `Vaults` map (name → abs path)
- `RegistryPath()` → `~/.config/vb/vaults.json` (respects `$XDG_CONFIG_HOME`)
- `Load()` / `Save()` — JSON file read/write, creates config dir if needed
- `Add()`, `Remove()`, `SetDefault()`, `Lookup()` — registry mutations with error handling

### `internal/registry/guard.go` (new)
- `CheckNesting(path)` — prevents creating vaults inside or around existing vaults
- Ancestor check: walks parent directories upward looking for `.vb/`
- Descendant check: walks subtree looking for `.vb/` in subdirectories

### `internal/vault/init.go` (new)
- `Init(path)` — creates `.vb/` scaffold (config.toml + index.json)
- Guards against re-initialization (errors if `.vb/` already exists)

### `cmd/vault.go` (new)
- `vb vault create <name> <path>` — nesting guard, mkdir, init scaffold, register, auto-set default if first
- `vb vault list` — sorted list with `*` marker for default
- `vb vault use <name>` — set default vault
- `vb vault remove <name>` — unregister (files unchanged), clears default if removed

### `cmd/resolve.go` (new)
- `flagVault` string — persistent `--vault` / `-V` flag on rootCmd
- `resolveVault()` — 3-priority resolution strategy

### `cmd/init.go`
- Added `--name` / `-n` flag
- After `vault.Init()`, optionally registers in global registry

### `cmd/root.go`
- Registered `vaultCmd` subcommand
- Added `--vault` / `-V` persistent flag

### `cmd/custom_flags.go`
- Updated `registerCustomLenses` to look up the default vault in the registry
- Enables custom lenses to work from any directory when a default vault is set

### All command files (`query.go`, `edit.go`, `lint.go`, `reindex.go`, `tag.go`, `save.go`)
- Refactored to use centralized `resolveVault()` helper

## Tests

- `internal/registry/registry_test.go`: 10 tests (load empty, add/remove/lookup, set default, save/load round-trip, duplicate name, not found)
- `internal/registry/guard_test.go`: 5 tests (no nesting, ancestor vault, descendant vault, self .vb, non-existent path)
- `internal/vault/init_test.go`: 3 tests (success, already initialized, scaffold contents)
- `cmd/vault_test.go`: 9 tests (create, list, use, remove, create duplicate, remove unknown, list empty, create nesting, create existing vault)
