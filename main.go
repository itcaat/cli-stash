package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

var addCmd = &cobra.Command{
	Use:   "add [command1] [command2] ... [flags]",
	Short: "Add one or more commands to the stash",
	Run:   runAdd,
}

func init() {
	rootCmd.AddCommand(popCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().String("bulk", "", "Comma-separated commands to add")
	addCmd.Flags().String("file", "", "Path to a file containing commands (one per line)")

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

func runAdd(cmd *cobra.Command, args []string) {
	bulkString, _ := cmd.Flags().GetString("bulk")
	file, _ := cmd.Flags().GetString("file")

	var commands []string
	if bulkString != "" {
		// Split the bulk string by comma and trim spaces
		for _, c := range strings.Split(bulkString, ",") {
			trimmedCmd := strings.TrimSpace(c)
			if trimmedCmd != "" {
				commands = append(commands, trimmedCmd)
			}
		}
	}

	if file != "" {
		fileCommands, err := readCommandsFromFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading commands from file: %v\n", err)
			os.Exit(1)
		}
		commands = append(commands, fileCommands...)
	}

	// Add commands from args
	commands = append(commands, args...)

	if len(commands) == 0 {
		fmt.Println("No commands to add. Use --bulk, --file flag or add commands as arguments.")
		return
	}

	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}
	if numAdded, err := store.AddBulk(commands); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding commands: %v\n", err)
		os.Exit(1)
	} else if numAdded > 0 {
		fmt.Printf("Added %d new commands.\n", numAdded)
	} else {
		fmt.Println("No new commands to add.")
	}
}

func readCommandsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commands = append(commands, line)
		}
	}

	return commands, scanner.Err()
}
