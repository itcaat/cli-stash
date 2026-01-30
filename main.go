package main

import (
	"bufio"
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

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Save the last command or enter a new one",
	Run: func(cmd *cobra.Command, args []string) {
		runPush()
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
	Use:     "add",
	Aliases: []string{"-"},
	Short:   "Save command from pipe",
	Long:    "Read commands from stdin and save them. Usage: echo 'command' | cli-stash add",
	Run: func(cmd *cobra.Command, args []string) {
		if isPiped() {
			saveFromPipe()
		} else {
			fmt.Fprintln(os.Stderr, "Usage: echo 'command' | cli-stash add")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
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

func runPush() {
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
		fmt.Println("No saved commands. Use 'cli-stash push' to add some.")
		return
	}

	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}
}

func isPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func saveFromPipe() {
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if err := store.Add(line); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving command: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Saved: %s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
