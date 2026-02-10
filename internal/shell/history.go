package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetLastCommand attempts to get the last command from shell history
func GetLastCommand() (string, error) {
	// Check HISTFILE first
	if histFile := os.Getenv("HISTFILE"); histFile != "" {
		if cmd, err := getLastFromFile(histFile); err == nil && cmd != "" {
			return cmd, nil
		}
	}

	// Detect current shell
	shell := detectCurrentShell()

	switch shell {
	case "fish":
		if cmd, err := getLastFromFishHistory(); err == nil && cmd != "" {
			return cmd, nil
		}
	case "zsh":
		if cmd, err := getFromZshHistory(); err == nil && cmd != "" {
			return cmd, nil
		}
	case "bash":
		if cmd, err := getFromBashHistory(); err == nil && cmd != "" {
			return cmd, nil
		}
	}

	// Try all shells as fallback
	if cmd, err := getFromZshHistory(); err == nil && cmd != "" {
		return cmd, nil
	}
	if cmd, err := getFromBashHistory(); err == nil && cmd != "" {
		return cmd, nil
	}
	if cmd, err := getLastFromFishHistory(); err == nil && cmd != "" {
		return cmd, nil
	}

	return "", nil
}

// GetHistory returns recent shell history commands (newest first)
// Merges history from all available shells
func GetHistory(limit int) []string {
	seen := make(map[string]bool)
	var allCommands []string

	// Helper to add commands without duplicates
	addCommands := func(cmds []string) {
		for _, cmd := range cmds {
			if !seen[cmd] {
				seen[cmd] = true
				allCommands = append(allCommands, cmd)
			}
		}
	}

	// Check HISTFILE first
	if histFile := os.Getenv("HISTFILE"); histFile != "" {
		addCommands(getHistoryFromFile(histFile, limit))
	}

	// Try all shells and merge (each already returns newest first)
	addCommands(getFishHistoryAll(limit))
	addCommands(getZshHistoryAll(limit))
	addCommands(getBashHistoryAll(limit))

	// Limit total results
	if len(allCommands) > limit {
		allCommands = allCommands[:limit]
	}

	return allCommands
}

// GetFishHistory returns history from fish shell only
func GetFishHistory(limit int) []string {
	return getFishHistoryAll(limit)
}

// GetZshHistory returns history from zsh only
func GetZshHistory(limit int) []string {
	return getZshHistoryAll(limit)
}

// GetBashHistory returns history from bash only
func GetBashHistory(limit int) []string {
	return getBashHistoryAll(limit)
}

// detectCurrentShell tries to detect the actual running shell
func detectCurrentShell() string {
	// Fish sets FISH_VERSION
	if os.Getenv("FISH_VERSION") != "" {
		return "fish"
	}

	// Zsh sets ZSH_VERSION
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}

	// Bash sets BASH_VERSION
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}

	// Fall back to $SHELL
	return filepath.Base(os.Getenv("SHELL"))
}

// getHistoryFromFile reads history from a custom HISTFILE (bash/zsh format)
func getHistoryFromFile(path string, limit int) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	seen := make(map[string]bool)
	var commands []string

	for i := len(lines) - 1; i >= 0 && len(commands) < limit; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Handle zsh extended history format
		cmd := line
		if strings.HasPrefix(line, ":") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				cmd = parts[1]
			}
		}

		if shouldSkipCommand(cmd) || seen[cmd] {
			continue
		}

		seen[cmd] = true
		commands = append(commands, cmd)
	}

	return commands
}

// getLastFromFile reads the last command from a custom HISTFILE
func getLastFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		cmd := line
		if strings.HasPrefix(line, ":") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				cmd = parts[1]
			}
		}

		if !shouldSkipCommand(cmd) {
			return cmd, nil
		}
	}

	return "", nil
}

// getFishHistoryAll reads history from fish shell
// Fish history is at ~/.local/share/fish/fish_history
// Format: "- cmd: command\n  when: timestamp\n"
func getFishHistoryAll(limit int) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	// Fish can also use XDG_DATA_HOME
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	historyPath := filepath.Join(dataHome, "fish", "fish_history")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	seen := make(map[string]bool)
	var commands []string

	// Parse fish history format (read from end)
	for i := len(lines) - 1; i >= 0 && len(commands) < limit; i-- {
		line := lines[i]

		// Look for "- cmd: " prefix
		if strings.HasPrefix(line, "- cmd: ") {
			cmd := strings.TrimPrefix(line, "- cmd: ")
			cmd = unescapeFishCommand(cmd)

			if shouldSkipCommand(cmd) || seen[cmd] {
				continue
			}

			seen[cmd] = true
			commands = append(commands, cmd)
		}
	}

	return commands
}

// getLastFromFishHistory gets the last command from fish history
func getLastFromFishHistory() (string, error) {
	cmds := getFishHistoryAll(1)
	if len(cmds) > 0 {
		return cmds[0], nil
	}
	return "", nil
}

// unescapeFishCommand handles fish's escape sequences
func unescapeFishCommand(cmd string) string {
	// Fish escapes newlines as \n and backslashes as \\
	cmd = strings.ReplaceAll(cmd, "\\n", "\n")
	cmd = strings.ReplaceAll(cmd, "\\\\", "\\")
	return cmd
}

// shouldSkipCommand returns true if command should be filtered out
func shouldSkipCommand(cmd string) bool {
	return strings.HasPrefix(cmd, "stash") || strings.HasPrefix(cmd, "cli-stash")
}

// getZshHistoryAll reads recent commands from ~/.zsh_history
func getZshHistoryAll(limit int) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, ".zsh_history")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil
	}

	// Parse zsh history handling multi-line commands
	// Multi-line commands end with \ and continue on next line
	commands := parseZshHistory(string(data), limit)
	return commands
}

// parseZshHistory parses zsh history format including multi-line commands
func parseZshHistory(data string, limit int) []string {
	lines := strings.Split(data, "\n")
	seen := make(map[string]bool)
	var commands []string
	var allCommands []string
	var currentCmd strings.Builder

	for _, line := range lines {
		// Check if this is a new command entry (starts with : for extended history)
		if strings.HasPrefix(line, ": ") {
			// Save previous command if any
			if currentCmd.Len() > 0 {
				allCommands = append(allCommands, currentCmd.String())
				currentCmd.Reset()
			}

			// Extract command after timestamp (format: ": timestamp:0;command")
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				currentCmd.WriteString(parts[1])
			}
		} else if line != "" {
			// Continuation line
			if currentCmd.Len() > 0 {
				currentCmd.WriteString("\n")
			}
			currentCmd.WriteString(line)
		}
	}

	// Don't forget last command
	if currentCmd.Len() > 0 {
		allCommands = append(allCommands, currentCmd.String())
	}

	// Now filter and dedupe, reading from newest (end) first
	for i := len(allCommands) - 1; i >= 0 && len(commands) < limit; i-- {
		cmd := strings.TrimSpace(allCommands[i])
		// Unescape double backslashes (zsh stores \ as \\)
		cmd = strings.ReplaceAll(cmd, "\\\\", "\\")
		if cmd == "" || shouldSkipCommand(cmd) || seen[cmd] {
			continue
		}
		seen[cmd] = true
		commands = append(commands, cmd)
	}

	return commands
}

// getBashHistoryAll reads recent commands from ~/.bash_history
func getBashHistoryAll(limit int) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, ".bash_history")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil
	}

	// Parse bash history handling multi-line commands
	commands := parseBashHistory(string(data), limit)
	return commands
}

// parseBashHistory parses bash history format including multi-line commands
func parseBashHistory(data string, limit int) []string {
	lines := strings.Split(data, "\n")
	seen := make(map[string]bool)
	var commands []string

	// Build list of complete commands (handling continuations)
	var allCommands []string
	var currentCmd strings.Builder

	for _, line := range lines {
		// Check for line continuation (ends with \)
		if strings.HasSuffix(line, "\\") {
			if currentCmd.Len() > 0 {
				currentCmd.WriteString("\n")
			}
			currentCmd.WriteString(line)
		} else {
			if currentCmd.Len() > 0 {
				currentCmd.WriteString("\n")
				currentCmd.WriteString(line)
				allCommands = append(allCommands, currentCmd.String())
				currentCmd.Reset()
			} else if line != "" {
				allCommands = append(allCommands, line)
			}
		}
	}

	// Don't forget last command if incomplete
	if currentCmd.Len() > 0 {
		allCommands = append(allCommands, currentCmd.String())
	}

	// Now filter and dedupe, reading from newest (end) first
	for i := len(allCommands) - 1; i >= 0 && len(commands) < limit; i-- {
		cmd := strings.TrimSpace(allCommands[i])
		if cmd == "" || shouldSkipCommand(cmd) || seen[cmd] {
			continue
		}
		seen[cmd] = true
		commands = append(commands, cmd)
	}

	return commands
}

// getFromZshHistory reads the last command from ~/.zsh_history
func getFromZshHistory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	historyPath := filepath.Join(homeDir, ".zsh_history")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		cmd := line
		if strings.HasPrefix(line, ":") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				cmd = parts[1]
			}
		}

		if !shouldSkipCommand(cmd) {
			return cmd, nil
		}
	}

	return "", nil
}

// getFromBashHistory reads the last command from ~/.bash_history
func getFromBashHistory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	historyPath := filepath.Join(homeDir, ".bash_history")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if !shouldSkipCommand(line) {
			return line, nil
		}
	}

	return "", nil
}

// ExecuteCommand runs a command in a shell
func ExecuteCommand(cmd string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	command := exec.Command(shell, "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}
