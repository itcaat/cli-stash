package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/itcaat/cli-stash/internal/shell"
	"github.com/itcaat/cli-stash/internal/storage"
	"github.com/itcaat/cli-stash/internal/terminal"
	"github.com/itcaat/cli-stash/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "cli-stash",
	Short: "Save and recall shell commands",
	Long:  "A terminal UI application for saving and recalling shell commands with fuzzy search.",
	Run: func(cmd *cobra.Command, args []string) {
		runPop()
	},
}

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Show saved commands with fuzzy search",
	Run: func(cmd *cobra.Command, args []string) {
		runPop()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved commands",
	Run: func(cmd *cobra.Command, args []string) {
		runList()
	},
}

var addCmd = &cobra.Command{
	Use:   "add [command]",
	Short: "Add a new command (as argument or interactive)",
	Long: `Add a new command to stash.

As argument:
  cli-stash add echo "hello world"

Interactively (shows UI with last command from history):
  cli-stash add`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			saveCommand(strings.Join(args, " "))
		} else {
			runAdd()
		}
	},
}

func init() {
	rootCmd.AddCommand(popCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runAdd() {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

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

	if m, ok := finalModel.(ui.PopModel); ok {
		if selected := m.Selected(); selected != "" {
			// Increment usage counter
			store.IncrementUse(selected)

			if err := terminal.InsertInput(selected); err != nil {
				if clipErr := clipboard.WriteAll(selected); clipErr != nil {
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
		fmt.Println("No saved commands. Use 'cli-stash add' to add some.")
		return
	}

	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}
}

func saveCommand(cmd string) {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		fmt.Fprintln(os.Stderr, "Empty command")
		os.Exit(1)
	}

	if err := store.Add(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving command: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved: %s\n", cmd)
}
