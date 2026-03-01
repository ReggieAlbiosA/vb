# vb — Phase 07 · Distribution

**Engineering Blueprint**

Ship Layer · Cross-Platform Binaries · Shell Completions · Release Automation
From Local Build to Public Release

---

## Build Summary

### What
Phase 07 is the release infrastructure for `vb`. Every previous phase has been development-facing — Phase 07 makes the tool installable, discoverable, and distributable to end users without requiring Go to be installed.

Three components:
- **GoReleaser** — automated cross-platform binary builds triggered by a Git tag push. Two binary variants: headless (`vb`) and GUI (`vb-gui`), built per-platform via CI matrix runners.
- **Shell completions** — bash, zsh, fish, and PowerShell completion scripts generated from Cobra and packaged into release archives and system directories
- **Auto docs** — man page generation from Cobra command metadata, produced at release time alongside binaries

### Why
A CLI tool that requires users to run `go install` is a tool for Go developers only. GoReleaser produces standalone binaries that work for any sysadmin or DevOps engineer regardless of what's on their machine. Shell completions turn `vb` from a memorization exercise into a discovery experience — `vb <tab>` surfaces topics, lenses, and flags (including post-blueprint additions like `vault`, `topic`, `save`, `--mermaid`). Man pages make it feel like a first-class system tool (`man vb`).

### Importance
Phase 07 does **not** change any functional behaviour of `vb`. It only changes how users obtain and interact with it. The constraints are:

- The standard `go build .` path must continue to work — GoReleaser is an additional release path, not a replacement
- GUI and non-GUI builds must both be supported in GoReleaser (two binary variants per platform)
- GUI builds require `CGO_ENABLED=1` and platform-native C libraries — they cannot be cross-compiled from a single runner
- The `.goreleaser.yaml` must produce snapshot builds locally that developers can test without a tag
- No build secrets required for a snapshot build
- The `Makefile` provides local dev build targets (`make build`, `make build-gui`, `make test`, `make test-gui`)

### Pre-Release Context

The following features were added after Phase 06 but before Phase 07. They affect distribution because they add new commands and flags that appear in completions, man pages, and binary behavior:

| Feature | Impact on Distribution |
|---|---|
| Global vault registry (`vb vault create/list/use/remove`) | New subcommands auto-included in completions + man pages via Cobra |
| `--mermaid` / `-m` modifier flag | New flag auto-included in completions |
| Native webview GUI (replaces browser) | GUI build requires `CGO_ENABLED=1` + platform C libraries |
| `go.mod` replace directive (lvlrt/webview_go fork) | GoReleaser must resolve fork correctly during build |
| `--vault` / `-V` persistent flag | New flag auto-included in completions |

---

## Branching Model — GitFlow

Phase 07 follows the GitFlow branching model. CI is split into **tests** (every branch) and **releases** (tags only).

### Branch Flow

```
feature/07-distribution
       │
       ▼  PR
      dev  ← integration branch (tests only, no releases)
       │
       ▼  create branch
  release/v0.1.0  ← version bumps, final fixes (tests only)
       │
       ├──▶ PR into main  → merge → tag v0.1.0 → triggers release workflows
       │
       └──▶ PR into dev   → merge (back-merge release fixes)
```

### CI Trigger Rules

| Event | Trigger | Workflow |
|---|---|---|
| Push/PR on `feature/*`, `NN-*` branches | Tests only | `ci.yml` |
| Push/PR on `dev` | Tests only | `ci.yml` |
| Push/PR on `release/*` | Tests only | `ci.yml` |
| Push/PR on `main` | Tests only | `ci.yml` |
| Tag push `v*` (on main) | GoReleaser + GUI builds | `release.yml` + `release-gui.yml` |

> Release workflows **never** trigger on branch pushes or PRs. Only a `v*` tag — created after `release/*` merges into `main` — triggers binary builds and GitHub Release creation.

---

## File Structure

```
/                              ← project root
  .goreleaser.yaml             ← release pipeline config (headless builds only)
  Makefile                     ← local dev build targets (already exists from pre-release)
  .github/workflows/
    ci.yml                     ← tests on all branches + PRs (never releases)
    release.yml                ← headless release via GoReleaser (tag push only)
    release-gui.yml            ← per-platform GUI builds via matrix (tag push only)
  completions/
    vb.bash                    ← generated (not committed, produced at build time)
    vb.zsh                     ← generated
    vb.fish                    ← generated
    _vb (PowerShell)           ← generated
  man/
    vb.1                       ← generated man page
  cmd/
    completion.go              ← already provided by Cobra (no-op for Phase 07)
    completions.go             ← GenerateCompletions() helper
    man.go                     ← GenerateManPage() helper
    docs.go                    ← generation entrypoints (//go:build ignore)
```

> Completion scripts and man pages are generated at release time, not committed to the repo.
> `.goreleaser.yaml` IS committed — it is the release source of truth for headless builds.
> GUI builds use a separate workflow with per-platform runners.

---

## 00 · GoReleaser Config — .goreleaser.yaml

GoReleaser handles **headless builds only** (`CGO_ENABLED=0`). GUI builds require CGO and platform-native libraries, so they are built separately via per-platform CI runners (see Section 04b).

```yaml
version: 2

project_name: vb

before:
  hooks:
    - go mod tidy
    - go run cmd/generate_completions.go
    - go run cmd/generate_man.go

builds:
  # Standard (headless-safe) build — CGO_ENABLED=0, cross-compiles from any runner
  - id: vb
    main: .
    binary: vb
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: default
    builds: [vb]
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - completions/*
      - man/vb.1
      - README.md
      - LICENSE

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - Merge pull request
      - Merge branch

release:
  draft: false
  prerelease: auto
  extra_files: []

snapshot:
  name_template: "{{ .Tag }}-next"
```

> **Why no GUI build in GoReleaser?** The GUI binary (`vb-gui`) uses `webview_go` which requires `CGO_ENABLED=1` and platform-native C libraries (WebKit2GTK on Linux, WebKit on macOS, WebView2 on Windows). CGO cannot cross-compile — each target OS must be built on a runner of that OS. GoReleaser runs on a single runner, so it handles only the headless build. GUI binaries are built separately and attached to the same GitHub Release.

---

## 01 · Shell Completions — cmd/completions.go

Cobra's built-in completion engine already works (`vb completion bash`, etc.) from Phase 01. Phase 07 adds a **generation script** that writes the files to `completions/` as part of the GoReleaser `before` hooks:

```go
//go:build ignore

// generate_completions.go — run via `go run cmd/generate_completions.go`
// Called by the GoReleaser before hook to produce completion scripts.
package main

import (
	"os"

	"github.com/ReggieAlbiosA/vb/cmd"
)

func main() {
	os.MkdirAll("completions", 0o755)
	cmd.GenerateCompletions("completions")
}
```

Add `GenerateCompletions()` to the cmd package:

```go
// cmd/completions.go
func GenerateCompletions(dir string) {
	rootCmd.GenBashCompletionFile(filepath.Join(dir, "vb.bash"))          //nolint:errcheck
	rootCmd.GenZshCompletionFile(filepath.Join(dir, "vb.zsh"))            //nolint:errcheck
	rootCmd.GenFishCompletionFile(filepath.Join(dir, "vb.fish"), false)   //nolint:errcheck
	rootCmd.GenPowerShellCompletionFile(filepath.Join(dir, "_vb"))        //nolint:errcheck
}
```

> All commands registered on `rootCmd` — including post-blueprint additions (`vault`, `topic`, `save`) and flags (`--mermaid`, `--vault`) — are automatically included in generated completions via Cobra's introspection. No manual enumeration needed.

**User installation of completions after download:**

```bash
# bash
source completions/vb.bash
# or copy to: /etc/bash_completion.d/vb

# zsh
cp completions/vb.zsh ~/.zsh/completions/_vb
# or copy to: /usr/local/share/zsh/site-functions/_vb

# fish
cp completions/vb.fish ~/.config/fish/completions/vb.fish
```

---

## 02 · Man Page Generation — cmd/man.go

Using `github.com/spf13/cobra/doc` (already a Cobra sibling package):

```go
//go:build ignore

// generate_man.go — run via `go run cmd/generate_man.go`
package main

import (
	"os"

	"github.com/ReggieAlbiosA/vb/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	os.MkdirAll("man", 0o755)
	cmd.GenerateManPage("man")
}
```

```go
// cmd/man.go
func GenerateManPage(dir string) {
	header := &doc.GenManHeader{
		Title:   "VB",
		Section: "1",
		Source:  "vb",
		Manual:  "vb Manual",
	}
	doc.GenManTree(rootCmd, header, dir) //nolint:errcheck
}
```

> `GenManTree` walks the full command tree. Post-blueprint subcommands (`vb vault create`, `vb vault list`, `vb topic create`, etc.) generate their own man pages automatically (e.g., `vb-vault-create.1`).

**System installation:**

```bash
sudo cp man/vb*.1 /usr/local/share/man/man1/
sudo mandb
man vb
man vb-vault-create
```

---

## 03 · GoReleaser Before Hooks — Updated Config

Generation hooks are included in the `before` section of `.goreleaser.yaml` (shown in Section 00):

```yaml
before:
  hooks:
    - go mod tidy
    - go run cmd/generate_completions.go
    - go run cmd/generate_man.go
```

---

## 04a · GitHub Actions — CI Test Workflow (All Branches)

`.github/workflows/ci.yml` — runs tests on every push and PR. Covers all branches: feature, dev, release, main. **Never triggers release builds.**

```yaml
name: CI

on:
  push:
    branches:
      - dev
      - main
      - "release/**"
      - "feature/**"
      - "[0-9][0-9]-*"
      - "hotfix/**"
    tags-ignore:
      - "v*"
  pull_request:
    branches:
      - dev
      - main

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: go test ./... -v -race -count=1

      - name: Vet
        run: go vet ./...

      - name: Build (headless)
        run: go build -o vb .
```

> `tags-ignore: v*` ensures that when a tag is pushed, this workflow does NOT also run — only the release workflows fire on tags.
> The branch list covers all GitFlow branch patterns. PRs into `dev` (from feature branches) and into `main` (from release branches) both trigger tests.

---

## 04b · GitHub Actions — Headless Release Workflow (Tag Only)

`.github/workflows/release.yml` — triggered **only** by `v*` tag push. Creates the GitHub Release and uploads headless binaries via GoReleaser.

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

> Only the `GITHUB_TOKEN` is needed — automatically provided by GitHub Actions.
> No additional secrets required for standard release to GitHub Releases.
> This workflow creates the GitHub Release object. The GUI workflow (04c) attaches additional binaries to it.

---

## 04c · GitHub Actions — GUI Release Workflow (Tag Only, Per-Platform)

`.github/workflows/release-gui.yml` — triggered **only** by `v*` tag push. Builds GUI binaries on native per-platform runners and attaches them to the GitHub Release created by 04b.

```yaml
name: Release GUI

on:
  push:
    tags:
      - "v*"

jobs:
  gui-binaries:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
            artifact: vb-gui-linux-amd64
            setup: sudo apt-get update && sudo apt-get install -y libwebkit2gtk-4.1-dev
          - os: ubuntu-24.04-arm
            goos: linux
            goarch: arm64
            artifact: vb-gui-linux-arm64
            setup: sudo apt-get update && sudo apt-get install -y libwebkit2gtk-4.1-dev
          - os: macos-latest
            goos: darwin
            goarch: arm64
            artifact: vb-gui-darwin-arm64
            setup: ""
          - os: macos-13
            goos: darwin
            goarch: amd64
            artifact: vb-gui-darwin-amd64
            setup: ""
          - os: windows-latest
            goos: windows
            goarch: amd64
            artifact: vb-gui-windows-amd64.exe
            setup: ""

    runs-on: ${{ matrix.os }}
    permissions:
      contents: write

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Install platform dependencies
        if: matrix.setup != ''
        run: ${{ matrix.setup }}

      - name: Build GUI binary
        env:
          CGO_ENABLED: 1
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: go build -tags gui -o ${{ matrix.artifact }} .

      - name: Archive (tar.gz)
        if: matrix.goos != 'windows'
        run: tar -czf ${{ matrix.artifact }}.tar.gz ${{ matrix.artifact }}

      - name: Archive (zip)
        if: matrix.goos == 'windows'
        run: Compress-Archive -Path ${{ matrix.artifact }} -DestinationPath ${{ matrix.artifact }}.zip
        shell: pwsh

      - name: Upload to GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            ${{ matrix.artifact }}.tar.gz
            ${{ matrix.artifact }}.zip
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Platform Build Requirements

| Platform | Runner | CGO | Native Library | Notes |
|---|---|---|---|---|
| Linux amd64 | `ubuntu-latest` | Required | `libwebkit2gtk-4.1-dev` | Installed via apt |
| Linux arm64 | `ubuntu-24.04-arm` | Required | `libwebkit2gtk-4.1-dev` | ARM runner |
| macOS arm64 | `macos-latest` | Required | WebKit (built-in) | Zero external deps |
| macOS amd64 | `macos-13` | Required | WebKit (built-in) | Intel runner |
| Windows amd64 | `windows-latest` | Required | WebView2/Edge (built-in) | Ships with Win 10/11 |

### go.mod Replace Directive

The GUI build depends on `webview/webview_go`, which currently requires a fork for webkit2gtk-4.1 support:

```
replace github.com/webview/webview_go => github.com/lvlrt/webview_go v0.0.0-20250119213827-fc6fe8152db0
```

This replace directive is in `go.mod` and is resolved transparently by `go build`. When upstream PR #62 is merged, the replace can be removed. GoReleaser and CI both honor `go.mod` replace directives — no special handling needed.

---

## 05 · .gitignore Updates

Add generated artifacts to `.gitignore` — they are produced at release time, not committed:

```gitignore
# GoReleaser output
dist/

# Generated at release time
completions/
man/

# Local binary artifacts
vb
vb-gui
```

---

## 06 · Local Development Builds

The `Makefile` (created during pre-release refinements) provides local build targets:

```makefile
.PHONY: build build-gui test test-gui

build:
	go build -o vb .

build-gui:
	CGO_ENABLED=1 go build -tags gui -o vb .

test:
	go test ./...

test-gui:
	CGO_ENABLED=1 go test -tags gui ./...
```

> `make build` produces the headless binary. `make build-gui` produces the GUI binary (requires platform C libraries). Both paths remain available regardless of GoReleaser.

---

## Validation Checklist

### GitFlow CI Triggers

- ✔ Push to `07-distribution` branch → `ci.yml` runs tests, **no release**
- ✔ PR from `07-distribution` into `dev` → `ci.yml` runs tests, **no release**
- ✔ Push to `dev` (after merge) → `ci.yml` runs tests, **no release**
- ✔ Push to `release/v0.1.0` branch → `ci.yml` runs tests, **no release**
- ✔ PR from `release/v0.1.0` into `main` → `ci.yml` runs tests, **no release**
- ✔ Push to `main` (after merge) → `ci.yml` runs tests, **no release**
- ✔ Tag push `v0.1.0` on `main` → `release.yml` + `release-gui.yml` fire, `ci.yml` does **not**

### GoReleaser (Headless)

- ✔ `goreleaser check` passes on `.goreleaser.yaml` — no config errors
- ✔ `goreleaser build --snapshot --clean` produces headless binaries in `dist/` for all targets
- ✔ `dist/vb_linux_amd64_v1/vb --help` works on the host machine
- ✔ Completions directory populated by snapshot build
- ✔ `source completions/vb.bash && vb <tab>` shows all subcommands (including `vault`, `topic`, `save`)
- ✔ `man/vb.1` generated and valid: `man ./man/vb.1`
- ✔ Tag push `v0.1.0` triggers GitHub Actions release workflow
- ✔ GitHub Release contains: headless binaries, checksums, completions, man pages, README
- ✔ `go build .` still works without GoReleaser (local dev path unchanged)

### GUI Builds (Per-Platform)

- ✔ `make build-gui` on Linux produces working GUI binary (requires `libwebkit2gtk-4.1-dev`)
- ✔ `./vb disk --arch --gui` opens native webview window (not browser)
- ✔ `./vb disk --arch --mermaid --gui` renders mermaid diagram in native window
- ✔ Second `--gui` invocation adds tab to existing window via IPC
- ✔ Headless fallback: `--gui` without DISPLAY/WAYLAND_DISPLAY prints to terminal
- ✔ GitHub Actions GUI workflow produces binaries for all 5 platform targets
- ✔ GUI binaries attached to the same GitHub Release as headless binaries

### Release Artifact Inventory

A complete `v*` release should contain:

| Artifact | Source |
|---|---|
| `vb_linux_amd64.tar.gz` | GoReleaser |
| `vb_linux_arm64.tar.gz` | GoReleaser |
| `vb_darwin_amd64.tar.gz` | GoReleaser |
| `vb_darwin_arm64.tar.gz` | GoReleaser |
| `vb_windows_amd64.zip` | GoReleaser |
| `vb-gui-linux-amd64.tar.gz` | GUI workflow (ubuntu-latest) |
| `vb-gui-linux-arm64.tar.gz` | GUI workflow (ubuntu-24.04-arm) |
| `vb-gui-darwin-arm64.tar.gz` | GUI workflow (macos-latest) |
| `vb-gui-darwin-amd64.tar.gz` | GUI workflow (macos-13) |
| `vb-gui-windows-amd64.exe.zip` | GUI workflow (windows-latest) |
| `checksums.txt` | GoReleaser |

---

## Test Coverage Requirements

Thresholds: **≥80% on cmd**

> Phase 07 adds no new runtime packages — only build tooling and generation scripts.
> The generation scripts use `//go:build ignore` and are excluded from `go test ./...`.

### cmd/completions_test.go

| Test | Covers |
|---|---|
| `TestGenerateCompletions_CreatesFiles` | All 4 completion files written to temp dir |
| `TestGenerateManPage_CreatesFile` | `vb.1` written to temp dir, non-empty |

Run:
```bash
go test ./cmd/... -v -race -count=1
```

---

## Release Procedure (GitFlow)

Step-by-step process to ship a release from feature branch to tagged GitHub Release.

### 1. Feature → Dev

```bash
# Work on feature branch
git checkout -b 07-distribution dev
# ... implement, commit, push
git push -u origin 07-distribution

# Create PR: 07-distribution → dev
# CI runs tests (ci.yml) — must pass before merge
# Merge PR into dev
```

### 2. Dev → Release Branch

```bash
# Create release branch from dev
git checkout -b release/v0.1.0 dev
# Version bumps, changelog edits, final fixes
git push -u origin release/v0.1.0

# CI runs tests (ci.yml) on the release branch
```

### 3. Release → Main + Dev

```bash
# PR: release/v0.1.0 → main
# CI runs tests — must pass before merge
# Merge PR into main

# PR: release/v0.1.0 → dev (back-merge)
# Merge PR into dev
```

### 4. Tag → Release

```bash
# Tag the merge commit on main
git checkout main
git pull
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Tag push triggers:
#   release.yml   → GoReleaser builds headless binaries, creates GitHub Release
#   release-gui.yml → per-platform runners build GUI binaries, attach to same Release
```

### 5. Verify

- GitHub Release page shows all 11 artifacts (5 headless + 5 GUI + checksums)
- Download and test headless binary on target platform
- Download and test GUI binary on target platform (requires native libraries)

---

```
vb engineering blueprint
phase 07 · distribution
```
