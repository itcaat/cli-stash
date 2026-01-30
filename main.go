package main

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/itcaat/cli-stash/internal/shell"
	"github.com/itcaat/cli-stash/internal/storage"
	"github.com/itcaat/cli-stash/internal/terminal"
	"github.com/itcaat/cli-stash/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		// Default: show pop UI (list and select commands)
		runPop()
		return
	}

	switch os.Args[1] {
	case "push":
		runPush()
	case "pop":
		runPop()
	case "list":
		runList()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func runPush() {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Get last command from shell history
	lastCmd, _ := shell.GetLastCommand()

	model := ui.NewPushModel(lastCmd, store)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runPop() {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	model, err := ui.NewPopModel(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading commands: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If a command was selected, insert it into terminal
	if m, ok := finalModel.(ui.PopModel); ok {
		if selected := m.Selected(); selected != "" {
			// Try to insert into terminal input buffer
			if err := terminal.InsertInput(selected); err != nil {
				// Fallback to clipboard if TIOCSTI fails
				if clipErr := clipboard.WriteAll(selected); clipErr != nil {
					// Last resort: just print
					fmt.Println(selected)
				} else {
					fmt.Fprintf(os.Stderr, "Copied to clipboard: %s\n", selected)
				}
			}
		}
	}
}

func runList() {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	commands, err := store.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading commands: %v\n", err)
		os.Exit(1)
	}

	if len(commands) == 0 {
		fmt.Println("No saved commands. Use 'cli-stash push' to add some.")
		return
	}

	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}
}

func printHelp() {
	help := `cli-stash - Save and recall shell commands

Usage:
  cli-stash              Show saved commands with fuzzy search
  cli-stash push         Save the last command (or enter a new one)
  cli-stash pop          Same as 'cli-stash' - show and select commands
  cli-stash list         List all saved commands
  cli-stash help         Show this help message

Navigation (in interactive mode):
  ↑/↓ or Ctrl+P/N    Navigate through commands
  Enter              Select command
  Ctrl+D             Delete selected command
  Esc                Cancel

Tips:
  - Use 'cli-stash push' after running a command you want to save
  - Selected command is inserted into terminal, ready to edit/execute
  - If terminal insert fails, command is copied to clipboard
`
	fmt.Print(help)
}
