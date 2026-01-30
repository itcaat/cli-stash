package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetLastCommand attempts to get the last command from shell history
func GetLastCommand() (string, error) {
	// Try reading from zsh history first (most common on macOS)
	if cmd, err := getFromZshHistory(); err == nil && cmd != "" {
		return cmd, nil
	}

	// Fall back to bash history
	if cmd, err := getFromBashHistory(); err == nil && cmd != "" {
		return cmd, nil
	}

	return "", nil
}

// GetHistory returns recent shell history commands (newest first)
func GetHistory(limit int) []string {
	// Try zsh first
	if cmds := getZshHistoryAll(limit); len(cmds) > 0 {
		return cmds
	}

	// Fall back to bash
	return getBashHistoryAll(limit)
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

	lines := strings.Split(string(data), "\n")
	seen := make(map[string]bool)
	var commands []string

	// Read from end (newest first)
	for i := len(lines) - 1; i >= 0 && len(commands) < limit; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Zsh history format: ": timestamp:0;command" or just "command"
		cmd := line
		if strings.HasPrefix(line, ":") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				cmd = parts[1]
			}
		}

		// Skip stash commands and duplicates
		if strings.HasPrefix(cmd, "stash") || strings.HasPrefix(cmd, "cli-stash") {
			continue
		}
		if seen[cmd] {
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

	lines := strings.Split(string(data), "\n")
	seen := make(map[string]bool)
	var commands []string

	for i := len(lines) - 1; i >= 0 && len(commands) < limit; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "stash") || strings.HasPrefix(line, "cli-stash") {
			continue
		}
		if seen[line] {
			continue
		}

		seen[line] = true
		commands = append(commands, line)
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
	
	// Find the last non-empty line that's not a stash command
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Zsh history format can be: ": timestamp:0;command" or just "command"
		cmd := line
		if strings.HasPrefix(line, ":") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				cmd = parts[1]
			}
		}

		// Skip stash commands themselves
		if strings.HasPrefix(cmd, "stash") || strings.HasPrefix(cmd, "cli-stash") {
			continue
		}

		return cmd, nil
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
	
	// Find the last non-empty line that's not a stash command
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Skip stash commands themselves
		if strings.HasPrefix(line, "stash") || strings.HasPrefix(line, "cli-stash") {
			continue
		}

		return line, nil
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
