# cli-stash

A terminal UI application for saving and recalling shell commands, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Save commands** - Use `cli-stash push` to save the last command from your shell history
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
go build -o cli-stash .
sudo mv cli-stash /usr/local/bin/
```

## Usage

### Save a Command

After running a command you want to save:

```bash
cli-stash push
```

This shows the last command from your shell history. Press Enter to save it, or edit it first.

### Recall a Command

```bash
cli-stash
# or
cli-stash pop
```

This opens an interactive UI where you can:
- Type to filter commands
- Use ↑/↓ to navigate
- Press Enter to select (command is inserted into terminal, ready to execute)
- Press Ctrl+D to delete a command
- Press Esc to cancel

### List All Commands

```bash
cli-stash list
```

### How It Works

When you select a command, it's automatically inserted into your terminal prompt. Just press Enter to execute it, or edit it first.

If terminal insert is not supported (e.g., on newer Linux kernels), the command is copied to clipboard instead.

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
