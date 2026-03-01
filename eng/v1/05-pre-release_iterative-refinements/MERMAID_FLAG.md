# `--mermaid` / `-m` — Mermaid rendering modifier

## Context

Some lenses have two file formats: `.md` (markdown) and `.mmd` (mermaid diagram). For example, a topic may have both `ARCH.md` (prose architecture doc) and `ARCH.mmd` (architecture diagram in Mermaid syntax). Running `vb disk --arch` always resolved to `ARCH.md`. There was no way to request the mermaid version.

## Usage

```bash
# Renders ARCH.md (markdown — default)
vb disk --arch

# Renders ARCH.mmd (mermaid diagram)
vb disk --arch --mermaid
vb disk --arch -m

# Combined with GUI for native window rendering
vb disk --arch -m --gui
```

## What changed

| Before | After |
|--------|-------|
| `--arch` always resolved to `ARCH.md` | `--arch` resolves to `ARCH.md`, `--arch -m` resolves to `ARCH.mmd` |
| No way to view .mmd files via CLI flags | `-m` / `--mermaid` modifier swaps `.md` → `.mmd` |
| Mermaid detection was content-based only | File selection is now explicit via flag |

## Implementation

### `cmd/query.go`
- Added `flagMermaid` bool variable
- Registered `--mermaid` / `-m` flag: `BoolVarP(&flagMermaid, "mermaid", "m", false, ...)`
- After `ResolveLens()`, if `flagMermaid` is set: `strings.TrimSuffix(lensFile, ".md") + ".mmd"`
- Works with any lens: `--why -m` → `WHY.mmd`, `--refs -m` → `REFS.mmd`, etc.

## Tests

- `cmd/query_test.go`:
  - `TestQueryCmd_MermaidModifier`: `vb disk --arch --mermaid` with ARCH.mmd present → success
  - `TestQueryCmd_MermaidModifier_Short`: `vb disk --arch -m` → same behavior
  - `TestQueryCmd_MermaidModifier_NoFile`: `--arch --mermaid` without .mmd file → clear error
  - Updated `resetQueryFlags` to include `"mermaid"` in cleanup list
