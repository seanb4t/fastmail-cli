// Package tui implements an interactive terminal UI using bubbletea.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// view represents the current screen.
type view int

const (
	viewMailboxList view = iota
)

// Model is the top-level bubbletea model.
type Model struct {
	client *fastmail.Client
	view   view
	width  int
	height int
	err    error
	quit   bool
}

// New creates a new TUI model with the given client.
func New(client *fastmail.Client) Model {
	return Model{
		client: client,
		view:   viewMailboxList,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quit {
		return ""
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.\n", m.err)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Faint(true).
		Padding(1, 2)

	title := titleStyle.Render("fastmail-cli")
	help := helpStyle.Render("q: quit")

	return fmt.Sprintf("\n%s\n\n  Welcome to fastmail-cli TUI.\n  Mailbox view coming soon.\n\n%s\n", title, help)
}

// Run starts the TUI program.
func Run(client *fastmail.Client) error {
	m := New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
