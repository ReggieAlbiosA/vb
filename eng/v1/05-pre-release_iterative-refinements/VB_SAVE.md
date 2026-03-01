# `vb save` — Command Cookbook (replaces access log)

## Context

The old `--used` modifier appended timestamped access log entries to USED.md (`- 2026-02-28 14:30 UTC  vb disk --why`). Nobody reads access logs. This enhancement replaces that with a **command cookbook** — save shell commands with descriptions, recall them instantly.

## Usage

```bash
# Save commands you just used
vb save partition "lsblk" -d "show all block devices in tree view"
vb save partition "sudo parted -l" -d "list all partition tables GPT-friendly"
vb save disk "sudo smartctl -H /dev/sda" -d "quick SMART health check"

# Read saved commands back
vb partition --used

# Manually curate saved commands
vb edit partition --used
```

## What changed

| Before | After |
|--------|-------|
| `--used` was a query modifier (append access log) | `--used` is a proper lens (renders USED.md) |
| `vb disk --why --used` logged the query | `vb disk --used` renders saved commands |
| USED.md format: `- TIMESTAMP  vb topic --lens` | USED.md format: `- command — description` |
| No `vb save` command | `vb save <topic> <command> -d <desc>` |

## Implementation

### `internal/resolver/lens.go`
- Added `"used": "USED.md"` to LensToFile map
- `--used` is now a first-class lens like `--why`, `--arch`, etc.

### `internal/logger/used.go`
- Replaced `Append(topicDir, topic, lens)` with `Save(topicDir, command, description)`
- New format: `- <command> — <description>`
- Same append-only file mechanics

### `cmd/save.go` (new)
- `vb save <topic> <command> -d <description>`
- ExactArgs(2), required `-d` flag
- Resolves topic via index, calls logger.Save

### `cmd/query.go`
- Removed logger import and `if flagUsed { logger.Append(...) }` block
- `--used` flag description changed to "saved commands for this topic"
- Works as a lens automatically via LensToFile pipeline

### `cmd/edit.go`
- Added `editFlagUsed` — `vb edit <topic> --used` opens USED.md in editor

## Tests

- `internal/logger/used_test.go`: 4 tests (create, append, format, error)
- `cmd/save_test.go`: 4 tests (success, not found, missing desc, multiple entries)
- `internal/resolver/lens_test.go`: Updated `TestActiveLens_UsedIsALens`
- `cmd/query_test.go`: Updated `TestQueryCmd_UsedLens`
