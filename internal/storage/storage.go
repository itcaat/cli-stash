package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Command represents a saved command with metadata
type Command struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UseCount  int       `json:"use_count"`
}

// Storage handles saving and loading commands
type Storage struct {
	path string
}

// New creates a new Storage instance
func New() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	stashDir := filepath.Join(homeDir, ".stash")
	if err := os.MkdirAll(stashDir, 0755); err != nil {
		return nil, err
	}

	return &Storage{
		path: filepath.Join(stashDir, "commands.json"),
	}, nil
}

// Load reads all commands from storage
func (s *Storage) Load() ([]Command, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Command{}, nil
		}
		return nil, err
	}

	var commands []Command
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, err
	}

	return commands, nil
}

// Save writes all commands to storage
func (s *Storage) Save(commands []Command) error {
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// Add appends a new command to storage
func (s *Storage) Add(text string) error {
	commands, err := s.Load()
	if err != nil {
		return err
	}

	// Check if command already exists
	for _, cmd := range commands {
		if cmd.Text == text {
			return nil // Already exists, skip
		}
	}

	commands = append(commands, Command{
		Text:      text,
		CreatedAt: time.Now(),
		UseCount:  0,
	})

	return s.Save(commands)
}

// Remove deletes a command from storage
func (s *Storage) Remove(text string) error {
	commands, err := s.Load()
	if err != nil {
		return err
	}

	filtered := make([]Command, 0, len(commands))
	for _, cmd := range commands {
		if cmd.Text != text {
			filtered = append(filtered, cmd)
		}
	}

	return s.Save(filtered)
}

// IncrementUse increases the use count for a command
func (s *Storage) IncrementUse(text string) error {
	commands, err := s.Load()
	if err != nil {
		return err
	}

	for i, cmd := range commands {
		if cmd.Text == text {
			commands[i].UseCount++
			break
		}
	}

	return s.Save(commands)
}

// List returns all command texts sorted by usage (most used first)
func (s *Storage) List() ([]string, error) {
	commands, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Sort by use count (descending), then by date (newest first)
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].UseCount != commands[j].UseCount {
			return commands[i].UseCount > commands[j].UseCount
		}
		return commands[i].CreatedAt.After(commands[j].CreatedAt)
	})

	texts := make([]string, len(commands))
	for i, cmd := range commands {
		texts[i] = cmd.Text
	}

	return texts, nil
}
