# Stash

A terminal UI application for saving and recalling shell commands, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Save commands** - Use `stash push` to save the last command from your shell history
- **Fuzzy search** - Quickly find saved commands with real-time filtering
- **Interactive UI** - Navigate with arrow keys, select with Enter
- **Delete commands** - Remove unwanted commands with Ctrl+D

## Installation

### Homebrew (macOS/Linux)

```bash
brew install itcaat/tap/cli-stash
```

### From Source

```bash
git clone https://github.com/itcaat/cli-stash.git
cd cli-stash
go build -o stash .
sudo mv stash /usr/local/bin/
```

## Usage

### Save a Command

After running a command you want to save:

```bash
stash push
```

This shows the last command from your shell history. Press Enter to save it, or edit it first.

### Recall a Command

```bash
stash
# or
stash pop
```

This opens an interactive UI where you can:
- Type to filter commands
- Use ↑/↓ to navigate
- Press Enter to select (prints to stdout)
- Press Ctrl+D to delete a command
- Press Esc to cancel

### List All Commands

```bash
stash list
```

### Execute a Selected Command

```bash
eval $(stash)
```

## Keybindings

| Key | Action |
|-----|--------|
| ↑ / Ctrl+P | Move up |
| ↓ / Ctrl+N | Move down |
| Enter | Select/Save |
| Ctrl+D | Delete command |
| Esc | Cancel |

## Storage

Commands are stored in `~/.stash/commands.json`.

## License

MIT
