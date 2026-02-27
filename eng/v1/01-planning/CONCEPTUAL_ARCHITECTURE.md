# vb — verbose

**CLI Tool Blueprint**

Multi-Lens Terminal Knowledge Base · Go · Cross-Platform · Single Binary · Vault-Based

---

# Identity

## CLI Tool Name

**vb** (verbose)

---

## What It Is

**Multi-Lens Terminal Knowledge Base** — a vault-based personal reference system queryable entirely from the terminal.

---

## What It Does

**vb** is a terminal-native knowledge base where you store and query technical information through **topic + lens**
combinations.

You give it a noun like **disk** or **ssh** and a flag like:

- `--why`
- `--arch`
- `--cli-tools`
- `--importance`

And it returns exactly that dimension of knowledge — nothing more, nothing tangential.

Knowledge lives as structured Markdown files inside a **vault** — a folder you initialize with:

```bash
vb init
```

Every topic has its own folder containing flag-mapped files like:

- `WHY.md`
- `ARCH.md`
- `CLI_TOOLS.md`
- `USED.md`

You author directly in your editor of choice.

No GUI required.
No special interface.

The more you feed it — especially via `--used` — the more it becomes your single terminal command for answering:

- What is this?
- Why does it exist?
- How does it work?
- What do I use for it?
- What exact command did I use last time?

No browser.
No context switch.
Just ask.

---

# Vault Structure

```
<any-dir>/ # vault root (any name, any location)
  │
  ├── .vb/ # vault marker (created via vb init)
  │ ├── config.toml # editor, theme, lint_on_save, knowledge_path
  │ └── index.json # auto-generated topic index
  │
  ├── hardware/ # category
  │ └── disk/ # topic
  │ ├── WHY.md # vb disk --why
  │ ├── IMPORTANCE.md # vb disk --importance
  │ ├── CLI_TOOLS.md # vb disk --cli-tools
  │ ├── ARCH.md # vb disk --arch
  │ ├── ARCH.mmd # vb disk --arch --gui
  │ ├── USED.md # vb disk --used
  │ ├── GOTCHAS.md # vb disk --gotchas
  │ └── REFS.md # vb disk --refs
  │
  ├── networking/
  ├── security/
  └── devops/
  ```

  ---

  # Flag System — Lenses

  Each flag maps directly to a file.

  ## `--why`
  → `WHY.md`
  Foundational reason the topic exists.

  ---

  ## `--importance`
  → `IMPORTANCE.md`
  Why it matters. What depends on it.

  ---

  ## `--cli-tools`
  → `CLI_TOOLS.md`
  Scoped command reference for this topic only.

  ---

  ## `--arch`
  → `ARCH.md`
  Terminal tree format architecture.

  Add `--gui` to render Mermaid from `ARCH.mmd`.

  ---

  ## `--used`
  → `USED.md`

  Your personal command log.

  Example:

  ```bash
  vb disk --used 'lsblk -f' -d "show all filesystems"
  ```

  Solves the permanent:

  > "Wait… what was that command again?"

  ---

  ## `--gotchas`
  → `GOTCHAS.md`
  Common mistakes and footguns.

  ---

  ## `--refs`
  → `REFS.md`
  Source documentation and references.

  ---

  ## `--gui` (Modifier)

  Composable with any flag.

  Examples:

  ```bash
  vb disk --arch --gui
  vb disk --why --gui
  ```

  Terminal is default.
  GUI is optional.

  ---

  # Tech Stack

  ## Language — Foundation

  ### Go

  - Single static binary
  - No runtime dependencies
  - Cross-platform
  - Fast startup
  - Low memory

  Ship:

  - `vb`
  - `vb.exe`
  - `vb-macos`

  ---

  ## Core CLI Structure

  ### Cobra
  - Command structure
  - Flags
  - Subcommands
  - Auto help
  - Shell completion

  ### Viper
  - Reads `.vb/config.toml`
  - Manages persistent config

  ---

  ## Terminal Rendering

  ### Glamour
  Markdown → styled terminal output

  ### Lip Gloss
  Terminal styling (color, borders, padding)

  ### Goldmark
  Markdown parsing + AST access (for linting)

  ---

  ## Linting & Validation

  ### Custom Schema Linter

  Rules per flag file:

  - `CLI_TOOLS.md` → command blocks required
  - `WHY.md` → prose
  - `ARCH.md` → tree nodes
  - `USED.md` → command + description pairs

  Runs via:

  ```bash
  vb lint <topic>
    ```

    Or automatically if:

    ```toml
    lint_on_save = true
    ```

    ---

    ## GUI Render (`--gui`)

    ### Wails or lorca

    - Lightweight Go webview
    - Not Electron
    - Spawned only when `--gui` is passed
    - Closes after use

    ---

    ## Editor Integration

    ### os/exec (stdlib)

    Launch:

    - micro
    - nano
    - vim
    - $EDITOR

    No dependency required.

    ---

    ## Distribution

    ### goreleaser

    Builds:

    - vb-linux-amd64
    - vb-darwin-arm64
    - vb-windows-amd64.exe

    ---

    ## Shell Completion

    Built-in via Cobra:

    ```bash
    vb <TAB>
      vb disk <TAB>
        ```

        Autocomplete topics from index.

        ---

        # Project Specifications

        ## Vault Initialization

        ```bash
        vb init
        ```

        Creates `.vb/` in current directory.

        Without it, vb does nothing.

        Multiple vaults supported.

        ---

        ## Exact File-to-Flag Mapping

        No magic.

        - `--why` → `WHY.md`
        - `--arch` → `ARCH.md`
        - `--cli-tools` → `CLI_TOOLS.md`
        - `--used` → `USED.md`

        If file doesn't exist → explicit error.

        ---

        ## `--used` Is the Daily Driver

        Append command:

        ```bash
        vb disk --used 'lsblk -f' -d "show filesystems"
        ```

        Query:

        ```bash
        vb disk --used
        ```

        Tag:

        ```bash
        vb --used --tag mount
        ```

        ---

        ## `--gui` Is Always Optional

        Terminal-first design.

        GUI never forced.

        Graceful fallback if display unavailable.

        ---

        ## `.vb/config.toml`

        Persistent config:

        ```toml
        editor = "vim"
        knowledge_path = "."
        theme = "dark"
        tree_style = "unicode"
        lint_on_save = true
        ```

        Separates tool defaults from user ownership.

        ---

        ## Scoped Querying

        `vb disk --cli-tools` returns disk tools only.

        No tangents.

        Scope enforced by authoring + linter.

        ---

        ## Cross-Platform by Design

        Single Go binary.

        Works in:

        - Linux
        - macOS
        - Windows
        - SSH
        - Containers
        - VMs

        ---

        ## Value Compounds Over Time

        Day 1: sparse.
        Month 6: irreplaceable.

        Every saved `--used` command is friction permanently removed.

        The vault becomes:

        Your experience.
        Your patterns.
        Your memory.

        No one else has that.

        ---

        vb · verbose · multi-lens terminal knowledge base
        Go · Cobra · Glamour · Lip Gloss · Wails