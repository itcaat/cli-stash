package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itcaat/cli-stash/internal/storage"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// PushModel represents the push command UI
type PushModel struct {
	textInput   textinput.Model
	lastCommand string
	saved       bool
	err         error
	quitting    bool
	storage     *storage.Storage
}

// NewPushModel creates a new push model
func NewPushModel(lastCmd string, store *storage.Storage) PushModel {
	ti := textinput.New()
	ti.Placeholder = "Enter command to save..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60

	if lastCmd != "" {
		ti.SetValue(lastCmd)
	}

	return PushModel{
		textInput:   ti,
		lastCommand: lastCmd,
		storage:     store,
	}
}

// Init initializes the model
func (m PushModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m PushModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			text := m.textInput.Value()
			if text != "" {
				if err := m.storage.Add(text); err != nil {
					m.err = err
				} else {
					m.saved = true
				}
			}
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the UI
func (m PushModel) View() string {
	if m.quitting {
		return subtitleStyle.Render("Cancelled.") + "\n"
	}

	if m.saved {
		return successStyle.Render("✓ Command saved to stash!") + "\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	s := titleStyle.Render("Stash Push") + "\n\n"

	if m.lastCommand != "" {
		s += subtitleStyle.Render("Last command detected:") + "\n"
	}

	s += m.textInput.View() + "\n\n"
	s += subtitleStyle.Render("Press Enter to save • Esc to cancel")

	return s + "\n"
}

// Saved returns whether the command was saved
func (m PushModel) Saved() bool {
	return m.saved
}
