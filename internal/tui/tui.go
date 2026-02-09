// Package tui implements an interactive terminal UI using bubbletea.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// view represents the current screen.
type view int

const (
	viewMailboxList view = iota
)

// errMsg wraps errors from async commands.
type errMsg struct{ err error }

// connectedMsg signals successful client connection.
type connectedMsg struct{}

// Model is the top-level bubbletea model.
type Model struct {
	client      *fastmail.Client
	view        view
	mailboxList mailboxListModel
	width       int
	height      int
	err         error
	quit        bool
	connecting  bool
}

// New creates a new TUI model with the given client.
func New(client *fastmail.Client) Model {
	return Model{
		client:      client,
		view:        viewMailboxList,
		mailboxList: newMailboxListModel(),
		connecting:  true,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.connectCmd()
}

func (m Model) connectCmd() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		if err := client.Connect(context.Background()); err != nil {
			return errMsg{err: err}
		}
		return connectedMsg{}
	}
}

func (m Model) fetchMailboxesCmd() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		mailboxes, err := client.Mailbox().List(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		return mailboxesLoadedMsg{mailboxes: mailboxes}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit — but only when the list isn't filtering
		if !m.mailboxList.list.SettingFilter() {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quit = true
				return m, tea.Quit
			}
		} else if msg.String() == "ctrl+c" {
			m.quit = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mailboxList.setSize(msg.Width, msg.Height)

	case connectedMsg:
		m.connecting = false
		return m, m.fetchMailboxesCmd()

	case errMsg:
		m.err = msg.err
		m.connecting = false
		return m, nil
	}

	// Delegate to current view
	if m.view == viewMailboxList {
		var cmd tea.Cmd
		m.mailboxList, cmd = m.mailboxList.update(msg)
		return m, cmd
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

	if m.connecting {
		return "\n  Connecting to Fastmail..."
	}

	if m.view == viewMailboxList {
		return m.mailboxList.view()
	}

	return ""
}

// Run starts the TUI program.
func Run(client *fastmail.Client) error {
	m := New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
