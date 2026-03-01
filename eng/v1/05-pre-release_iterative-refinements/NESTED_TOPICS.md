# Nested Topics — `vb topic` command group

## Context

Currently all topics in `vb` are flat — `hardware/disk`, `networking/ssh` — where the first path segment is a category and the second is the topic. There's no support for topics inside topics. Users need sub-topics to model hierarchical knowledge (e.g. `partition` has sub-topics `fs`, `mnt`, `swap`; `fs` itself could have `ext4`, `btrfs`).

This enhancement adds:
- **`vb topic create <name> [--in <parent>]`** — create a topic with 6 empty lens files, optionally nested under a parent
- **`vb topic list [--tree]`** — list all topics, optionally as a tree
- **`..` separator** for querying nested topics: `vb partition..fs --why`
- **Dual-key indexing** in `vb reindex` so both `fs` (leaf shorthand) and `partition..fs` (explicit path) resolve

## Commands

```
vb topic create disk                     # create TopicRoot/disk/ with 6 empty .md files
vb topic create fs --in partition        # create TopicRoot/partition/fs/ with 6 .md files
vb topic create mnt --in partition..fs   # create TopicRoot/partition/fs/mnt/ with 6 .md files
vb topic list                            # flat sorted list of all topics
vb topic list --tree                     # indented tree view
```

Querying nested topics:
```
vb partition..fs --why                   # resolves to TopicRoot/partition/fs/WHY.md
vb partition..fs..mnt --arch             # resolves to TopicRoot/partition/fs/mnt/ARCH.md
vb fs --why                              # leaf shorthand (works if "fs" is unambiguous)
```

## Design Rules

- A directory is a **topic** only if it contains at least one `.md` file
- Only the **leaf** (the created topic) gets scaffolded with 6 empty lens files
- Parent directories in `--in` are just `mkdir -p`'d path segments (no .md files created for them)
- Parents that already have .md files remain valid topics — nesting doesn't break them
- `..` is the separator for explicit nested topic addressing (not `/` or `:`)

---

## Implementation

### Dual-key indexing — `internal/index/index.go`

`Build()` now emits two keys for any topic whose relative path has depth > 1:

```go
topicName := d.Name()
schema.Topics[topicName] = rel  // "fs" → "partition/fs"

dotDotKey := strings.ReplaceAll(rel, "/", "..")
if dotDotKey != topicName {
    schema.Topics[dotDotKey] = rel  // "partition..fs" → "partition/fs"
}
```

Leaf-name keys use last-write-wins (consistent with prior behavior). The `..`-joined key provides unambiguous disambiguation.

### `..` resolution — `internal/resolver/topic.go`

No changes needed — `ResolveTopic()` is a map lookup. The `..`-joined key from indexing means `schema.Topics["partition..fs"]` already maps to `"partition/fs"`.

### Topic commands — `cmd/topic.go`

| Command | Flags | Behavior |
|---------|-------|----------|
| `vb topic create <name>` | `--in / -i` | Scaffold 6 empty lens .md files; auto-reindex |
| `vb topic list` | `--tree / -t` | Flat sorted list or indented tree view |

`runTopicCreate`:
1. `resolveVault()` → ctx
2. If `--in`: resolve parent via index, nest under it
3. Guard against overwriting existing topic (has .md files)
4. `os.MkdirAll` + create 6 empty lens files
5. `index.Build()` — auto-reindex

`runTopicList`:
- Flat mode: deduplicate paths, prefer `..`-joined display key, sorted
- Tree mode: deduplicate paths, print with 2-space indentation per depth level

---

## Files modified

| File | Change |
|------|--------|
| `internal/index/index.go` | Add `..`-joined key in `Build()` |
| `internal/index/index_test.go` | Updated count assertions + 3 new dual-key tests |
| `cmd/root.go` | Register `topicCmd` |

## Files created

| File | Purpose |
|------|---------|
| `cmd/topic.go` | `vb topic create` and `vb topic list` |
| `cmd/topic_test.go` | 8 tests for topic create/list |

## Test coverage

- `cmd/topic_test.go`: 8 tests (flat create, nested create, deep nested, already exists, parent not found, list empty, list flat, list tree)
- `internal/index/index_test.go`: 3 new tests (dual key, dual key top-level, dual key deep)
- Full suite: `go test ./...` — all packages pass
