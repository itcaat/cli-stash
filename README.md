# cli-stash

A terminal UI application for saving and recalling shell commands, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Save commands** - Browse shell history and save with Ctrl+A
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

### Recall a Command

```bash
cli-stash
```

This opens an interactive UI where you can:
- Type to filter commands
- Use ↑/↓ to navigate
- Press Enter to select (command is inserted into terminal)
- Press Ctrl+A to add from shell history
- Press Ctrl+D to delete a command
- Press Esc to cancel

### Add a Command

Press **Ctrl+A** in the main view to browse your shell history. Type to filter, then press Enter to save the selected command.

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
| Ctrl+A | Browse shell history |
| Ctrl+D | Delete command |
| Esc | Cancel / Back |

## Storage

Commands are stored in `~/.stash/commands.json`.

## License

MIT
