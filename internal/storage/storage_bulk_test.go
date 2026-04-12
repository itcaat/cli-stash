package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddBulk(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "stash-test-bulk")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create storage with custom path
	store := &Storage{
		path: filepath.Join(tmpDir, "commands.json"),
	}

	// Initial commands
	initialCmds := []string{"echo 'first'", "ls -l"}
	numAdded, err := store.AddBulk(initialCmds)
	if err != nil {
		t.Fatalf("Initial AddBulk() error = %v", err)
	}
	if numAdded != 2 {
		t.Fatalf("Initial AddBulk() numAdded = %d, want 2", numAdded)
	}

	// Test AddBulk with new and duplicate commands
	t.Run("AddBulk_NewAndDuplicates", func(t *testing.T) {
		bulkCmds := []string{"git status", "ls -l", "echo 'another'", "git status"}
		numAdded, err := store.AddBulk(bulkCmds)
		if err != nil {
			t.Errorf("AddBulk() error = %v", err)
		}
		if numAdded != 2 {
			t.Errorf("AddBulk() numAdded = %d, want 2", numAdded)
		}

		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}

		// Expect 4 unique commands
		if len(commands) != 4 {
			t.Errorf("Load() len = %d, want 4", len(commands))
		}

		// Check for specific commands
		expectedTexts := map[string]bool{
			"echo 'first'":   true,
			"ls -l":          true,
			"git status":     true,
			"echo 'another'": true,
		}

		for _, cmd := range commands {
			if !expectedTexts[cmd.Text] {
				t.Errorf("Unexpected command found: %q", cmd.Text)
			}
		}
	})

	// Test AddBulk with only existing commands
	t.Run("AddBulk_OnlyExisting", func(t *testing.T) {
		bulkCmds := []string{"ls -l", "echo 'first'"}
		numAdded, err := store.AddBulk(bulkCmds)
		if err != nil {
			t.Errorf("AddBulk() error = %v", err)
		}
		if numAdded != 0 {
			t.Errorf("AddBulk() numAdded = %d, want 0", numAdded)
		}

		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if len(commands) != 4 {
			t.Errorf("Load() len after adding only existing = %d, want 4", len(commands))
		}
	})
}
