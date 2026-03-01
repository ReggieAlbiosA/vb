# vb

A personal knowledge vault for the command line. Store, query, and render technical knowledge organized by topic and lens.

Unlike generic cheatsheet tools (`tldr`, `cheat`, `man`), `vb` is a personal knowledge base — you write your own notes, organized into topics with multiple lenses per topic.

## Install

### From release (recommended)

Download the latest binary from [Releases](https://github.com/ReggieAlbiosA/vb/releases) and place it in your `$PATH`.

### From source

```bash
go install github.com/ReggieAlbiosA/vb@latest
```

### GUI build (native webview)

The GUI variant requires CGO and platform libraries:

```bash
# Linux
sudo apt install libwebkit2gtk-4.1-dev
make build-gui

# macOS — no extra deps needed
make build-gui
```

## Quick start

```bash
# Initialize a vault
vb init --name myknowledge

# Create a topic
vb topic create disk

# Edit a lens
vb edit disk --arch

# Query
vb disk --arch
vb disk --why
vb disk --cli-tools

# Render as mermaid diagram
vb disk --arch --mermaid

# Open in native GUI window
vb disk --arch --gui

# Save a command
vb save disk "lsblk -f" --desc "List block devices with filesystem info"
```

## Commands

| Command | Description |
|---|---|
| `vb <topic> <--lens>` | Query a topic with a lens |
| `vb init` | Initialize a vault in the current directory |
| `vb edit <topic> <--lens>` | Open a lens file in `$EDITOR` |
| `vb lint <topic> <--lens>` | Validate a lens file against its schema |
| `vb tag <tag>` | Search for a tag across all topics |
| `vb reindex` | Rebuild the topic index |
| `vb save <topic> "<cmd>" -d "<desc>"` | Save a command to a topic |
| `vb topic create <name>` | Create a new topic |
| `vb topic list` | List all topics |
| `vb vault create <name> <path>` | Create and register a vault |
| `vb vault list` | List registered vaults |
| `vb vault use <name>` | Set default vault |
| `vb vault remove <name>` | Unregister a vault |

## Lenses

Each topic can have multiple lens files:

| Flag | File | Content |
|---|---|---|
| `--why` | `WHY.md` | Why this topic exists in your stack |
| `--arch` | `ARCH.md` | Architecture overview |
| `--cli-tools` | `CLI_TOOLS.md` | CLI tools and commands |
| `--gotchas` | `GOTCHAS.md` | Pitfalls and edge cases |
| `--refs` | `REFS.md` | Reference links |
| `--importance` | `IMPORTANCE.md` | Impact and relevance |

Custom lenses can be defined in `.vb/config.toml`.

## Modifiers

| Flag | Effect |
|---|---|
| `--gui` | Render in native desktop window |
| `--mermaid` / `-m` | Swap `.md` to `.mmd` for diagram rendering |
| `--vault` / `-V` | Target a specific vault by name |

## License

[MIT](LICENSE)
