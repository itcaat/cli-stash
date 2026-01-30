package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itcaat/cli-stash/internal/storage"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
)

// PopModel represents the pop/list command UI
type PopModel struct {
	textInput    textinput.Model
	commands     []string
	filtered     []string
	cursor       int
	selected     string
	quitting     bool
	adding       bool
	storage      *storage.Storage
	executed     bool
	executeError error
}

// NewPopModel creates a new pop model
func NewPopModel(store *storage.Storage) (PopModel, error) {
	commands, err := store.List()
	if err != nil {
		return PopModel{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "Type to filter commands..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	return PopModel{
		textInput: ti,
		commands:  commands,
		filtered:  commands,
		storage:   store,
	}, nil
}

// Init initializes the model
func (m PopModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m PopModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle adding mode
		if m.adding {
			switch msg.String() {
			case "ctrl+c", "esc":
				// Cancel adding, return to list
				m.adding = false
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Type to filter commands..."
				return m, nil

			case "enter":
				// Save the new command
				text := m.textInput.Value()
				if text != "" {
					m.storage.Add(text)
					m.commands, _ = m.storage.List()
					m.filtered = m.commands
				}
				m.adding = false
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Type to filter commands..."
				m.cursor = 0
				return m, nil
			}

			// Update text input in adding mode
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// Normal mode
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
			}
			return m, tea.Quit

		case "ctrl+a":
			// Switch to adding mode
			m.adding = true
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Enter new command to save..."
			return m, nil

		case "ctrl+d", "delete":
			// Delete the selected command
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				cmdToDelete := m.filtered[m.cursor]
				if err := m.storage.Remove(cmdToDelete); err == nil {
					// Refresh the list
					m.commands, _ = m.storage.List()
					m.filtered = m.filterCommands(m.textInput.Value())
					if m.cursor >= len(m.filtered) && m.cursor > 0 {
						m.cursor--
					}
				}
			}
			return m, nil
		}
	}

	// Update text input
	prevValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)

	// Filter commands if input changed
	if m.textInput.Value() != prevValue {
		m.filtered = m.filterCommands(m.textInput.Value())
		m.cursor = 0
	}

	return m, cmd
}

// filterCommands filters commands based on input
func (m PopModel) filterCommands(query string) []string {
	if query == "" {
		return m.commands
	}

	query = strings.ToLower(query)
	var filtered []string

	for _, cmd := range m.commands {
		if strings.Contains(strings.ToLower(cmd), query) {
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}

// highlightMatch highlights the matching part of a command
func highlightMatch(cmd, query string) string {
	if query == "" {
		return cmd
	}

	lowerCmd := strings.ToLower(cmd)
	lowerQuery := strings.ToLower(query)
	
	idx := strings.Index(lowerCmd, lowerQuery)
	if idx == -1 {
		return cmd
	}

	before := cmd[:idx]
	match := cmd[idx : idx+len(query)]
	after := cmd[idx+len(query):]

	return before + matchStyle.Render(match) + after
}

// View renders the UI
func (m PopModel) View() string {
	if m.quitting {
		return dimStyle.Render("Cancelled.") + "\n"
	}

	if m.selected != "" {
		return "" // main.go handles inserting into terminal
	}

	// Adding mode
	if m.adding {
		s := titleStyle.Render("Add Command") + "\n\n"
		s += m.textInput.View() + "\n\n"
		s += dimStyle.Render("Enter to save • Esc to cancel")
		return s + "\n"
	}

	s := titleStyle.Render("Stash") + " " + dimStyle.Render("- saved commands (sorted by usage)") + "\n\n"
	s += m.textInput.View() + "\n\n"

	if len(m.commands) == 0 {
		s += dimStyle.Render("No saved commands. Press Ctrl+A or run 'cli-stash add'.") + "\n"
	} else if len(m.filtered) == 0 {
		s += dimStyle.Render("No matching commands.") + "\n"
	} else {
		// Show filtered commands
		maxShow := 10
		start := 0
		if m.cursor >= maxShow {
			start = m.cursor - maxShow + 1
		}

		end := start + maxShow
		if end > len(m.filtered) {
			end = len(m.filtered)
		}

		for i := start; i < end; i++ {
			cmd := m.filtered[i]
			displayCmd := highlightMatch(cmd, m.textInput.Value())

			if i == m.cursor {
				s += selectedStyle.Render("▸ ") + displayCmd + "\n"
			} else {
				s += normalStyle.Render("  ") + displayCmd + "\n"
			}
		}

		// Show count
		s += "\n" + dimStyle.Render(fmt.Sprintf("Showing %d of %d commands", len(m.filtered), len(m.commands)))
	}

	s += "\n\n" + dimStyle.Render("↑/↓ navigate • Enter select • Ctrl+A add • Ctrl+D delete • Esc cancel")

	return s + "\n"
}

// Selected returns the selected command
func (m PopModel) Selected() string {
	return m.selected
}
