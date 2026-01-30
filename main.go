package main

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

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

func init() {
	rootCmd.AddCommand(popCmd)
	rootCmd.AddCommand(listCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
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
		fmt.Println("No saved commands. Run 'cli-stash' and press Ctrl+A to add.")
		return
	}

	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}
}
