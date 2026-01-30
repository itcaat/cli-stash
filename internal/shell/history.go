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
		if strings.HasPrefix(cmd, "stash") {
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
		if strings.HasPrefix(line, "stash") {
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
