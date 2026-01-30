package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "stash-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create storage with custom path
	store := &Storage{
		path: filepath.Join(tmpDir, "commands.json"),
	}

	// Test Load on non-existent file
	t.Run("LoadEmpty", func(t *testing.T) {
		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if len(commands) != 0 {
			t.Errorf("Load() = %v, want empty slice", commands)
		}
	})

	// Test Add
	t.Run("Add", func(t *testing.T) {
		err := store.Add("echo hello")
		if err != nil {
			t.Errorf("Add() error = %v", err)
		}

		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("Load() len = %d, want 1", len(commands))
		}
		if commands[0].Text != "echo hello" {
			t.Errorf("Load()[0].Text = %q, want %q", commands[0].Text, "echo hello")
		}
	})

	// Test Add duplicate (should skip)
	t.Run("AddDuplicate", func(t *testing.T) {
		err := store.Add("echo hello")
		if err != nil {
			t.Errorf("Add() error = %v", err)
		}

		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("Load() len = %d, want 1 (duplicate should be skipped)", len(commands))
		}
	})

	// Test Add second command
	t.Run("AddSecond", func(t *testing.T) {
		err := store.Add("ls -la")
		if err != nil {
			t.Errorf("Add() error = %v", err)
		}

		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if len(commands) != 2 {
			t.Errorf("Load() len = %d, want 2", len(commands))
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		texts, err := store.List()
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
		if len(texts) != 2 {
			t.Errorf("List() len = %d, want 2", len(texts))
		}
	})

	// Test Remove
	t.Run("Remove", func(t *testing.T) {
		err := store.Remove("echo hello")
		if err != nil {
			t.Errorf("Remove() error = %v", err)
		}

		commands, err := store.Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("Load() len = %d, want 1", len(commands))
		}
		if commands[0].Text != "ls -la" {
			t.Errorf("Load()[0].Text = %q, want %q", commands[0].Text, "ls -la")
		}
	})
}

func TestNew(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if store == nil {
		t.Error("New() returned nil")
	}

	if store.path == "" {
		t.Error("New() path is empty")
	}

	// Verify path contains .stash
	if !filepath.IsAbs(store.path) {
		t.Errorf("New() path is not absolute: %s", store.path)
	}
}
