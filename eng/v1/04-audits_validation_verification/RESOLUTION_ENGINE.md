# Phase 02 · Resolution Engine — Audit & Validation

---

## What Was Built

| File | Purpose |
|---|---|
| [internal/resolver/topic.go](file:///home/marianne/Documents/vb/internal/resolver/topic.go) | Topic lookup against `index.Schema` → absolute topic directory path |
| [internal/resolver/lens.go](file:///home/marianne/Documents/vb/internal/resolver/lens.go) | Flag → filename map, `ResolveLens()`, `ActiveLens()` with single-flag enforcement |
| [internal/resolver/binding.go](file:///home/marianne/Documents/vb/internal/resolver/binding.go) | File existence validation via `os.Stat`, user-readable error with `vb edit` hint |
| [internal/resolver/errors.go](file:///home/marianne/Documents/vb/internal/resolver/errors.go) | Sentinel errors owned in `internal/resolver/` — not a separate package |
| [cmd/query.go](file:///home/marianne/Documents/vb/cmd/query.go) | `vb <topic> --<lens>` Cobra command wiring — threads resolution chain end-to-end |

---

## Smoke Tests — Blueprint Validation Checklist

### ✅ vb disk --why with existing topic + existing lens file
Resolution chain completes, returns absolute file path ✓

### ✅ vb disk --why with existing topic + missing lens file
```
Error: "WHY.md" not authored yet for topic "disk" — run: vb edit disk --why
```
User-readable error with actionable hint ✓

### ✅ vb unknown --why → ErrTopicNotFound
```
Error: topic not found in index: "unknown"
```

### ✅ vb disk (no lens) → ErrNoLens
```
Error: no lens flag provided
```

### ✅ vb disk --why --arch (two lenses) → ErrMultipleLenses
```
Error: only one lens flag may be used at a time
```

### ✅ vb disk --why --gui → resolves cleanly
`--gui` excluded from lens resolution, passed as modifier to renderer ✓

### ✅ index.json empty {} → ErrTopicNotFound, no crash
Empty index handled gracefully — not a panic ✓

### ✅ topicRoot correctness
`ResolveTopic` joins against `TopicRoot` from `VaultContext`, not `VaultRoot` ✓  
Decoupled `knowledge_path` vaults resolve correctly ✓

---

## Unit Tests

Run with: `go test ./internal/resolver/... ./cmd/... -v -race -count=1`

### internal/resolver/topic_test.go — 4 tests
| Test | Result |
|---|---|
| `TestResolveTopic_Found` | ✅ PASS |
| `TestResolveTopic_NotFound` | ✅ PASS |
| `TestResolveTopic_EmptyIndex` | ✅ PASS |
| `TestResolveTopic_UsesTopicRoot` | ✅ PASS |

### internal/resolver/lens_test.go — 6 tests
| Test | Result |
|---|---|
| `TestResolveLens_AllValidFlags` | ✅ PASS |
| `TestResolveLens_InvalidFlag` | ✅ PASS |
| `TestActiveLens_SingleFlag` | ✅ PASS |
| `TestActiveLens_NoFlag` | ✅ PASS |
| `TestActiveLens_MultipleFlags` | ✅ PASS |
| `TestActiveLens_GUINotALens` | ✅ PASS |

### internal/resolver/binding_test.go — 3 tests
| Test | Result |
|---|---|
| `TestBind_FileExists` | ✅ PASS |
| `TestBind_FileMissing` | ✅ PASS |
| `TestBind_ErrorMessageContainsTopic` | ✅ PASS |

### cmd/query_test.go — 5 tests
| Test | Result |
|---|---|
| `TestQueryCmd_Success` | ✅ PASS |
| `TestQueryCmd_TopicNotFound` | ✅ PASS |
| `TestQueryCmd_LensFileMissing` | ✅ PASS |
| `TestQueryCmd_NoLens` | ✅ PASS |
| `TestQueryCmd_GUIModifier` | ✅ PASS |

---

## Coverage

| Package | Coverage | Threshold | Status |
|---|---|---|---|
| `internal/resolver` | **96.3%** | ≥90% | ✅ |
| `cmd/` | **81.0%** | ≥80% | ✅ |

> `internal/resolver/errors.go` has 0% directly testable statements (only `var` declarations).  
> Coverage is satisfied via error path tests in the three resolver test files.

---

## Cumulative Coverage — All Phases

| Package | Coverage |
|---|---|
| `internal/vault` | **95.0%** |
| `internal/resolver` | **96.3%** |
| `internal/index` | **93.0%** |
| `internal/config` | **88.2%** |
| `cmd/` | **81.0%** |

---

## Phase 02 Verdict

**Complete and merge-ready.**

All blueprint validation steps passed, all 18 new tests passing with race detection, coverage exceeds threshold on all packages. The `topicRoot` correctness fix verified end-to-end. Resolution chain is wired from CLI input to file path.
