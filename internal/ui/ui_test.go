package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itcaat/cli-stash/internal/storage"
)

// createTestStorage creates a storage instance with a temp directory
func createTestStorage(t *testing.T) (*storage.Storage, func()) {
	tmpDir, err := os.MkdirTemp("", "stash-ui-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// We need to create the storage manually since New() uses the real home dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	store, err := storage.New()
	if err != nil {
		os.RemoveAll(tmpDir)
		os.Setenv("HOME", oldHome)
		t.Fatalf("Failed to create storage: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
		os.Setenv("HOME", oldHome)
	}

	return store, cleanup
}

func TestPushModel(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	t.Run("NewPushModel", func(t *testing.T) {
		model := NewPushModel("echo test", store)

		if model.textInput.Value() != "echo test" {
			t.Errorf("NewPushModel() textInput.Value() = %q, want %q", model.textInput.Value(), "echo test")
		}
		if model.lastCommand != "echo test" {
			t.Errorf("NewPushModel() lastCommand = %q, want %q", model.lastCommand, "echo test")
		}
	})

	t.Run("Init", func(t *testing.T) {
		model := NewPushModel("", store)
		cmd := model.Init()
		if cmd == nil {
			t.Error("Init() returned nil, expected blink command")
		}
	})

	t.Run("UpdateEsc", func(t *testing.T) {
		model := NewPushModel("test", store)
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

		pushModel := newModel.(PushModel)
		if !pushModel.quitting {
			t.Error("Update(Esc) should set quitting to true")
		}
		if cmd == nil {
			t.Error("Update(Esc) should return Quit command")
		}
	})

	t.Run("UpdateEnter", func(t *testing.T) {
		model := NewPushModel("save this", store)
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		pushModel := newModel.(PushModel)
		if !pushModel.saved {
			t.Error("Update(Enter) should set saved to true")
		}

		// Verify command was saved
		commands, err := store.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(commands) != 1 || commands[0] != "save this" {
			t.Errorf("Command was not saved correctly: %v", commands)
		}
	})

	t.Run("View", func(t *testing.T) {
		model := NewPushModel("test cmd", store)
		view := model.View()

		if !strings.Contains(view, "Stash Push") {
			t.Error("View() should contain title 'Stash Push'")
		}
		if !strings.Contains(view, "Enter to save") {
			t.Error("View() should contain help text")
		}
	})
}

func TestPopModel(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	// Add some test commands
	store.Add("echo hello")
	store.Add("ls -la")
	store.Add("git status")

	t.Run("NewPopModel", func(t *testing.T) {
		model, err := NewPopModel(store)
		if err != nil {
			t.Fatalf("NewPopModel() error = %v", err)
		}

		if len(model.commands) != 3 {
			t.Errorf("NewPopModel() commands len = %d, want 3", len(model.commands))
		}
		if len(model.filtered) != 3 {
			t.Errorf("NewPopModel() filtered len = %d, want 3", len(model.filtered))
		}
	})

	t.Run("FilterCommands", func(t *testing.T) {
		model, _ := NewPopModel(store)

		filtered := model.filterCommands("git")
		if len(filtered) != 1 {
			t.Errorf("filterCommands('git') len = %d, want 1", len(filtered))
		}
		if len(filtered) > 0 && filtered[0] != "git status" {
			t.Errorf("filterCommands('git')[0] = %q, want %q", filtered[0], "git status")
		}

		filtered = model.filterCommands("echo")
		if len(filtered) != 1 {
			t.Errorf("filterCommands('echo') len = %d, want 1", len(filtered))
		}

		filtered = model.filterCommands("")
		if len(filtered) != 3 {
			t.Errorf("filterCommands('') len = %d, want 3", len(filtered))
		}
	})

	t.Run("UpdateNavigation", func(t *testing.T) {
		model, _ := NewPopModel(store)

		// Test down navigation
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		popModel := newModel.(PopModel)
		if popModel.cursor != 1 {
			t.Errorf("Update(Down) cursor = %d, want 1", popModel.cursor)
		}

		// Test up navigation
		newModel, _ = popModel.Update(tea.KeyMsg{Type: tea.KeyUp})
		popModel = newModel.(PopModel)
		if popModel.cursor != 0 {
			t.Errorf("Update(Up) cursor = %d, want 0", popModel.cursor)
		}
	})

	t.Run("UpdateEnter", func(t *testing.T) {
		model, _ := NewPopModel(store)
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		popModel := newModel.(PopModel)
		if popModel.selected == "" {
			t.Error("Update(Enter) should set selected")
		}
	})

	t.Run("View", func(t *testing.T) {
		model, _ := NewPopModel(store)
		view := model.View()

		if !strings.Contains(view, "Stash") {
			t.Error("View() should contain title 'Stash'")
		}
		if !strings.Contains(view, "navigate") {
			t.Error("View() should contain help text")
		}
	})
}

func TestHighlightMatch(t *testing.T) {
	tests := []struct {
		cmd   string
		query string
		want  string
	}{
		{"echo hello", "", "echo hello"},
		{"git status", "git", ""}, // Contains styled text, just check it doesn't panic
		{"ls -la", "xyz", "ls -la"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd+"_"+tt.query, func(t *testing.T) {
			result := highlightMatch(tt.cmd, tt.query)
			if tt.want != "" && result != tt.want {
				// Only check exact match when no styling expected
				if tt.query == "" || !strings.Contains(strings.ToLower(tt.cmd), strings.ToLower(tt.query)) {
					if result != tt.want {
						t.Errorf("highlightMatch(%q, %q) = %q, want %q", tt.cmd, tt.query, result, tt.want)
					}
				}
			}
		})
	}
}

func TestEmptyStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stash-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create stash directory
	os.MkdirAll(filepath.Join(tmpDir, ".stash"), 0755)

	store, _ := storage.New()
	model, err := NewPopModel(store)
	if err != nil {
		t.Fatalf("NewPopModel() error = %v", err)
	}

	view := model.View()
	if !strings.Contains(view, "No saved commands") {
		t.Error("View() should show 'No saved commands' for empty storage")
	}
}
