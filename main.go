package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itcaat/cli-stash/internal/shell"
	"github.com/itcaat/cli-stash/internal/storage"
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

	// If a command was selected, print it (for use with eval)
	if m, ok := finalModel.(ui.PopModel); ok {
		if selected := m.Selected(); selected != "" {
			// Print to stdout so it can be captured or executed
			fmt.Println(selected)
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
		fmt.Println("No saved commands. Use 'stash push' to add some.")
		return
	}

	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}
}

func printHelp() {
	help := `Stash - Save and recall shell commands

Usage:
  stash              Show saved commands with fuzzy search
  stash push         Save the last command (or enter a new one)
  stash pop          Same as 'stash' - show and select commands
  stash list         List all saved commands
  stash help         Show this help message

Navigation (in interactive mode):
  ↑/↓ or Ctrl+P/N    Navigate through commands
  Enter              Select command
  Ctrl+D             Delete selected command
  Esc                Cancel

Tips:
  - Use 'stash push' after running a command you want to save
  - Selected commands are printed to stdout for easy piping
  - Example: eval $(stash) to execute a selected command
`
	fmt.Print(help)
}
