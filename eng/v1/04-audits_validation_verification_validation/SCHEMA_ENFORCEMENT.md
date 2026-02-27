# Phase 05 · Schema Enforcement — Audit & Validation

---

## What Was Built

| File | Purpose |
|---|---|
| [internal/linter/linter.go](file:///home/marianne/Documents/vb/internal/linter/linter.go) | `Lint()` entry point — reads file, parses AST via Goldmark, dispatches to lens rule |
| [internal/linter/rules.go](file:///home/marianne/Documents/vb/internal/linter/rules.go) | `LensRule` interface + 6 per-lens rule implementations + AST walk helpers |
| [internal/linter/errors.go](file:///home/marianne/Documents/vb/internal/linter/errors.go) | `LintError` struct with `Error()` method |
| [internal/hook/save.go](file:///home/marianne/Documents/vb/internal/hook/save.go) | `OnSave()` — post-edit hook, runs lint as `⚠` warnings when `lint_on_save = true` |
| [cmd/lint.go](file:///home/marianne/Documents/vb/cmd/lint.go) | `vb lint <topic> --<lens>` — full resolution chain → lint → structured output |

---

## Phase Completion

| Component | Blueprint requirement | Status |
|---|---|---|
| `Lint()` entry point | Goldmark AST parse, rule dispatch, schema-free fallback | ✅ Complete |
| `LensRules` map | 6 lens rules (`why`, `importance`, `cli-tools`, `arch`, `gotchas`, `refs`) | ✅ Complete |
| `USED.md` excluded | No rule for `"used"` — append-only log, always valid | ✅ Complete |
| `LintError` type | Struct + `Error()` method | ✅ Complete |
| `OnSave()` hook | `lintOnSave=false` skips; violations are `⚠` warnings, not errors | ✅ Complete |
| `cmd/lint.go` | Resolution chain → `Bind()` → `Lint()` → structured stdout/stderr | ✅ Complete |
| Non-zero exit on violation | `vb lint` exits non-zero; `OnSave` always exits 0 | ✅ Complete |
| No file modification | Pure read-only — `os.ReadFile` + AST walk only | ✅ Complete |
| Goldmark reuse | Direct import, no re-adding dependency | ✅ Complete |
| `internal/linter` import isolation | No import from `internal/render` | ✅ Complete |

**Phase 05 completion: 10/10 blueprint requirements — 100%**

---

## Smoke Tests — Blueprint Validation Checklist

### ✅ vb lint disk --why on a valid WHY.md → exits 0
```
✔ why: schema valid
```

### ✅ vb lint disk --why on an empty WHY.md → exits non-zero
```
✘ [why] WHY.md must contain at least one paragraph
Error: 1 schema violation(s) found
```

### ✅ vb lint disk --cli-tools on file with no code blocks → violation
```
✘ [cli-tools] CLI_TOOLS.md must contain at least one fenced code block
```

### ✅ vb lint disk --refs on file with no links → violation
```
✘ [refs] REFS.md must contain at least one link
```

### ✅ vb lint disk --used → no rule defined, exits 0
`"used"` has no entry in `LensRules` — returns `nil`, treated as always valid ✓

### ✅ vb lint unknown --why → ErrTopicNotFound
Resolver error returned before linter is reached — linter never called ✓

### ✅ lint_on_save = true + vb edit disk --why → warnings printed after editor exits
`OnSave()` called with `lintOnSave=true` — schema violations print as `⚠ [why] ...` to stderr ✓

### ✅ lint_on_save = false → no lint, editor exits cleanly
`OnSave()` returns immediately when `lintOnSave=false` — no stderr output ✓

### ✅ On-save violations are ⚠ warnings — vb edit still exits 0
Hook does not return an error — exit code 0 confirmed ✓

### ✅ Linter never modifies any file
Read-only path confirmed: `os.ReadFile` → AST parse → rule check — no write calls ✓

---

## Unit Tests

Run with: `go test ./internal/linter/... ./internal/hook/... ./cmd/... -v -race -count=1`

### internal/linter/linter_test.go — 5 tests

| Test | Result |
|---|---|
| `TestLint_ValidWhy` | ✅ PASS |
| `TestLint_EmptyFile` | ✅ PASS |
| `TestLint_NoRuleForLens` | ✅ PASS |
| `TestLint_FileNotFound` | ✅ PASS |
| `TestLintError_ErrorMethod` | ✅ PASS |

### internal/linter/rules_test.go — 12 tests

| Test | Result |
|---|---|
| `TestWhyRule_HasParagraph` | ✅ PASS |
| `TestWhyRule_NoParagraph` | ✅ PASS |
| `TestImportanceRule_HasParagraph` | ✅ PASS |
| `TestImportanceRule_NoParagraph` | ✅ PASS |
| `TestCLIToolsRule_HasCodeBlock` | ✅ PASS |
| `TestCLIToolsRule_NoCodeBlock` | ✅ PASS |
| `TestArchRule_HasHeading` | ✅ PASS |
| `TestArchRule_NoHeading` | ✅ PASS |
| `TestGotchasRule_HasListItem` | ✅ PASS |
| `TestGotchasRule_NoListItem` | ✅ PASS |
| `TestRefsRule_HasLink` | ✅ PASS |
| `TestRefsRule_NoLink` | ✅ PASS |

### internal/hook/save_test.go — 4 tests

| Test | Result |
|---|---|
| `TestOnSave_LintOnSaveFalse` | ✅ PASS |
| `TestOnSave_ValidFile` | ✅ PASS |
| `TestOnSave_LintReadError` | ✅ PASS |
| `TestOnSave_InvalidFile` | ✅ PASS |

### cmd/lint_test.go — 5 tests

| Test | Result |
|---|---|
| `TestLintCmd_Valid` | ✅ PASS |
| `TestLintCmd_Violation` | ✅ PASS |
| `TestLintCmd_TopicNotFound` | ✅ PASS |
| `TestLintCmd_LensFileMissing` | ✅ PASS |
| `TestLintCmd_NoLens` | ✅ PASS |

---

## Coverage

| Package | Coverage | Threshold | Status |
|---|---|---|---|
| `internal/linter` | **100.0%** | ≥90% | ✅ |
| `internal/hook` | **100.0%** | ≥90% | ✅ |
| `cmd/` | **82.6%** | ≥80% | ✅ |

---

## Cumulative Coverage — All Phases

| Package | Coverage |
|---|---|
| `internal/vault` | **95.0%** |
| `internal/resolver` | **96.3%** |
| `internal/index` | **93.0%** |
| `internal/render` | **92.9%** |
| `internal/editor` | **94.4%** |
| `internal/linter` | **100.0%** |
| `internal/hook` | **100.0%** |
| `internal/logger` | **100.0%** |
| `internal/tagger` | **94.1%** |
| `internal/config` | **88.2%** |
| `cmd/` | **82.6%** |

---

## Phase 05 Verdict

**Complete and merge-ready.**

All 10 blueprint requirements implemented. 26 new tests passing with race detection across linter, hook, and cmd packages. Both `internal/linter` and `internal/hook` hit 100% coverage. Schema enforcement is purely read-only — no file modifications. On-save hook correctly separates advisory warnings from hard failures. `vb lint` exits non-zero for CI gating; `OnSave` never blocks `vb edit`.
