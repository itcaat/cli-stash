package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itcaat/cli-stash/internal/shell"
	"github.com/itcaat/cli-stash/internal/storage"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	selectedStyle = lipgloss.NewStyle().
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
	textInput     textinput.Model
	commands      []string // saved commands
	filtered      []string
	history       []string // shell history
	historyFilter []string
	cursor        int
	selected      string
	quitting      bool
	historyMode   bool   // true = browsing history, false = browsing saved
	editMode      bool   // true = editing a command
	editOriginal  string // original command being edited
	storage       *storage.Storage
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
	ti.CharLimit = 1000
	ti.Width = 120

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
		// Edit mode
		if m.editMode {
			switch msg.String() {
			case "ctrl+c", "esc":
				// Cancel editing
				m.editMode = false
				m.editOriginal = ""
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Type to filter commands..."
				return m, nil

			case "enter":
				// Save edited command
				newText := m.textInput.Value()
				if newText != "" && newText != m.editOriginal {
					m.storage.Update(m.editOriginal, newText)
					m.commands, _ = m.storage.List()
					m.filtered = m.commands
				}
				m.editMode = false
				m.editOriginal = ""
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Type to filter commands..."
				m.cursor = 0
				return m, nil
			}

			// Update text input in edit mode
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// History browsing mode
		if m.historyMode {
			switch msg.String() {
			case "ctrl+c", "esc":
				// Return to saved commands
				m.historyMode = false
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Type to filter commands..."
				m.filtered = m.commands
				m.cursor = 0
				return m, nil

			case "up", "ctrl+p":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil

			case "down", "ctrl+n":
				if m.cursor < len(m.historyFilter)-1 {
					m.cursor++
				}
				return m, nil

			case "enter":
				// Save selected history command
				if len(m.historyFilter) > 0 && m.cursor < len(m.historyFilter) {
					selectedCmd := m.historyFilter[m.cursor]
					m.storage.Add(selectedCmd)
					m.commands, _ = m.storage.List()
					m.filtered = m.commands
				}
				// Return to saved commands view
				m.historyMode = false
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Type to filter commands..."
				m.cursor = 0
				return m, nil
			}

			// Update text input in history mode
			prevValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)

			if m.textInput.Value() != prevValue {
				m.historyFilter = m.filterHistory(m.textInput.Value())
				m.cursor = 0
			}
			return m, cmd
		}

		// Normal mode (saved commands)
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
			// Switch to history mode
			m.historyMode = true
			m.history = shell.GetHistory(500)
			m.historyFilter = m.history
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Type to filter history..."
			m.cursor = 0
			return m, nil

		case "ctrl+e":
			// Edit the selected command
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.editMode = true
				m.editOriginal = m.filtered[m.cursor]
				m.textInput.SetValue(m.editOriginal)
				m.textInput.Placeholder = ""
				m.textInput.CursorEnd()
			}
			return m, nil

		case "ctrl+d", "delete":
			// Delete the selected command
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				cmdToDelete := m.filtered[m.cursor]
				if err := m.storage.Remove(cmdToDelete); err == nil {
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

	if m.textInput.Value() != prevValue {
		m.filtered = m.filterCommands(m.textInput.Value())
		m.cursor = 0
	}

	return m, cmd
}

// filterCommands filters saved commands based on input
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

// filterHistory filters shell history based on input
func (m PopModel) filterHistory(query string) []string {
	if query == "" {
		return m.history
	}

	query = strings.ToLower(query)
	var filtered []string

	for _, cmd := range m.history {
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

	// Edit mode
	if m.editMode {
		s := titleStyle.Render("Edit Command") + "\n\n"
		s += m.textInput.View() + "\n\n"
		s += dimStyle.Render("Enter to save • Esc to cancel")
		return s + "\n"
	}

	// History mode
	if m.historyMode {
		s := titleStyle.Render("Shell History") + " " + dimStyle.Render("- select to save") + "\n\n"
		s += m.textInput.View() + "\n\n"

		if len(m.history) == 0 {
			s += dimStyle.Render("No shell history found.") + "\n"
		} else if len(m.historyFilter) == 0 {
			s += dimStyle.Render("No matching commands.") + "\n"
		} else {
			s += m.renderList(m.historyFilter)
			s += "\n" + dimStyle.Render(fmt.Sprintf("Showing %d of %d history items", len(m.historyFilter), len(m.history)))
		}

		s += "\n\n" + dimStyle.Render("↑/↓ navigate • Enter save • Esc back")
		return s + "\n"
	}

	// Normal mode (saved commands)
	s := titleStyle.Render("Stash") + " " + dimStyle.Render("- saved commands (sorted by usage)") + "\n\n"
	s += m.textInput.View() + "\n\n"

	if len(m.commands) == 0 {
		s += dimStyle.Render("No saved commands. Press Ctrl+A to add from history.") + "\n"
	} else if len(m.filtered) == 0 {
		s += dimStyle.Render("No matching commands.") + "\n"
	} else {
		s += m.renderList(m.filtered)
		s += "\n" + dimStyle.Render(fmt.Sprintf("Showing %d of %d commands", len(m.filtered), len(m.commands)))
	}

	s += "\n\n" + dimStyle.Render("↑/↓ navigate • Enter select • Ctrl+A add • Ctrl+E edit • Ctrl+D delete • Esc cancel")

	return s + "\n"
}

// renderList renders a list of commands with cursor
func (m PopModel) renderList(items []string) string {
	maxShow := 10
	start := 0
	if m.cursor >= maxShow {
		start = m.cursor - maxShow + 1
	}

	end := start + maxShow
	if end > len(items) {
		end = len(items)
	}

	var s string
	for i := start; i < end; i++ {
		cmd := items[i]

		if i == m.cursor {
			// Selected: bright cyan with underline
			s += selectedStyle.Render("▸ " + cmd) + "\n"
		} else {
			displayCmd := highlightMatch(cmd, m.textInput.Value())
			s += normalStyle.Render("  ") + displayCmd + "\n"
		}
	}

	return s
}

// Selected returns the selected command
func (m PopModel) Selected() string {
	return m.selected
}
