package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetLastCommand(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on the user's shell history
	cmd, err := GetLastCommand()
	
	// It's okay if there's no history or an error
	// We just want to make sure it doesn't crash
	_ = cmd
	_ = err
}

func TestGetFromZshHistory(t *testing.T) {
	// Create a temporary zsh history file
	tmpDir, err := os.MkdirTemp("", "stash-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	historyPath := filepath.Join(tmpDir, ".zsh_history")

	// Test with standard format
	t.Run("StandardFormat", func(t *testing.T) {
		content := `echo first
echo second
ls -la
`
		if err := os.WriteFile(historyPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write history file: %v", err)
		}

		cmd, err := getFromZshHistory()
		if err != nil {
			t.Errorf("getFromZshHistory() error = %v", err)
		}
		if cmd != "ls -la" {
			t.Errorf("getFromZshHistory() = %q, want %q", cmd, "ls -la")
		}
	})

	// Test with zsh extended format (: timestamp:0;command)
	t.Run("ExtendedFormat", func(t *testing.T) {
		content := `: 1234567890:0;echo first
: 1234567891:0;echo second
: 1234567892:0;git status
`
		if err := os.WriteFile(historyPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write history file: %v", err)
		}

		cmd, err := getFromZshHistory()
		if err != nil {
			t.Errorf("getFromZshHistory() error = %v", err)
		}
		if cmd != "git status" {
			t.Errorf("getFromZshHistory() = %q, want %q", cmd, "git status")
		}
	})

	// Test skipping stash commands
	t.Run("SkipStashCommands", func(t *testing.T) {
		content := `echo hello
stash push
stash
`
		if err := os.WriteFile(historyPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write history file: %v", err)
		}

		cmd, err := getFromZshHistory()
		if err != nil {
			t.Errorf("getFromZshHistory() error = %v", err)
		}
		if cmd != "echo hello" {
			t.Errorf("getFromZshHistory() = %q, want %q (should skip stash commands)", cmd, "echo hello")
		}
	})
}

func TestGetFromBashHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stash-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	historyPath := filepath.Join(tmpDir, ".bash_history")

	t.Run("BasicHistory", func(t *testing.T) {
		content := `cd /tmp
ls
pwd
`
		if err := os.WriteFile(historyPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write history file: %v", err)
		}

		cmd, err := getFromBashHistory()
		if err != nil {
			t.Errorf("getFromBashHistory() error = %v", err)
		}
		if cmd != "pwd" {
			t.Errorf("getFromBashHistory() = %q, want %q", cmd, "pwd")
		}
	})
}
