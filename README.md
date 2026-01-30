# cli-stash

A terminal UI application for saving and recalling shell commands, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Save commands** - Use `cli-stash add` or pipe commands directly
- **Fuzzy search** - Quickly find saved commands with real-time filtering
- **Smart sorting** - Commands sorted by usage frequency
- **Interactive UI** - Navigate with arrow keys, select with Enter

## Installation

### Homebrew (macOS/Linux)

```bash
brew install itcaat/tap/cli-stash
```

### From Source

```bash
git clone https://github.com/itcaat/cli-stash.git
cd cli-stash
go build -o cli-stash .
sudo mv cli-stash /usr/local/bin/
```

## Usage

### Add a Command

As argument:
```bash
cli-stash add echo "hello world"
```

Interactive mode (shows last command from history):
```bash
cli-stash add
```

### Recall a Command

```bash
cli-stash
# or
cli-stash pop
```

This opens an interactive UI where you can:
- Type to filter commands
- Use ↑/↓ to navigate
- Press Enter to select (command is inserted into terminal)
- Press Ctrl+A to add a new command
- Press Ctrl+D to delete a command
- Press Esc to cancel

### List All Commands

```bash
cli-stash list
```

## How It Works

When you select a command, it's automatically inserted into your terminal prompt. Just press Enter to execute it, or edit it first.

Commands are sorted by usage frequency - most used commands appear first.

## Keybindings

| Key | Action |
|-----|--------|
| ↑ / ↓ | Navigate |
| Enter | Select/Save |
| Ctrl+A | Add new command |
| Ctrl+D | Delete command |
| Esc | Cancel |

## Storage

Commands are stored in `~/.stash/commands.json`.

## License

MIT
