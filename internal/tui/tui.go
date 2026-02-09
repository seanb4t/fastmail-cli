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
	viewEmailList
	viewEmailReader
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
	emailList   *emailListModel
	emailReader *emailReaderModel
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

func (m Model) fetchEmailsCmd(mailboxID string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		emails, err := client.Mail().List(context.Background(), mailboxID, emailPageSize)
		if err != nil {
			return errMsg{err: err}
		}
		return emailsLoadedMsg{emails: emails}
	}
}

func (m Model) fetchEmailBodyCmd(emailID string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		email, err := client.Mail().GetWithBody(context.Background(), emailID)
		if err != nil {
			return errMsg{err: err}
		}
		return emailBodyLoadedMsg{email: *email}
	}
}

// isFiltering returns true if any active list is in filter mode.
func (m Model) isFiltering() bool {
	switch m.view {
	case viewMailboxList:
		return m.mailboxList.list.SettingFilter()
	case viewEmailList:
		return m.emailList != nil && m.emailList.list.SettingFilter()
	case viewEmailReader:
		return false
	}
	return false
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "q":
			// In the reader view, q means "go back" — let the reader handle it
			if m.view != viewEmailReader && !m.isFiltering() {
				m.quit = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mailboxList.setSize(msg.Width, msg.Height)
		if m.emailList != nil {
			m.emailList.setSize(msg.Width, msg.Height)
		}
		if m.emailReader != nil {
			m.emailReader.setSize(msg.Width, msg.Height)
		}

	case emailBodyLoadedMsg:
		if m.emailReader != nil {
			er := *m.emailReader
			er, _ = er.update(msg)
			m.emailReader = &er
		}
		return m, nil

	case connectedMsg:
		m.connecting = false
		return m, m.fetchMailboxesCmd()

	case mailboxSelectedMsg:
		el := newEmailListModel(msg.mailbox)
		el.setSize(m.width, m.height)
		m.emailList = &el
		m.view = viewEmailList
		return m, m.fetchEmailsCmd(msg.mailbox.ID)

	case errMsg:
		m.err = msg.err
		m.connecting = false
		return m, nil
	}

	// Delegate to current view
	return m.updateView(msg)
}

func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewMailboxList:
		return m.updateMailboxList(msg)
	case viewEmailList:
		return m.updateEmailList(msg)
	case viewEmailReader:
		return m.updateEmailReader(msg)
	}
	return m, nil
}

func (m Model) updateMailboxList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.mailboxList, cmd = m.mailboxList.update(msg)

	if m.mailboxList.selected != nil {
		mb := *m.mailboxList.selected
		m.mailboxList.selected = nil
		return m, func() tea.Msg { return mailboxSelectedMsg{mailbox: mb} }
	}

	return m, cmd
}

func (m Model) updateEmailList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.emailList == nil {
		return m, nil
	}

	var cmd tea.Cmd
	el := *m.emailList
	el, cmd = el.update(msg)
	m.emailList = &el

	if el.goBack {
		m.emailList = nil
		m.view = viewMailboxList
		return m, nil
	}

	if el.selected != nil {
		email := *el.selected
		el.selected = nil
		m.emailList = &el

		er := newEmailReaderModel(email)
		er.setSize(m.width, m.height)
		m.emailReader = &er
		m.view = viewEmailReader
		return m, m.fetchEmailBodyCmd(email.ID)
	}

	return m, cmd
}

func (m Model) updateEmailReader(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.emailReader == nil {
		return m, nil
	}

	var cmd tea.Cmd
	er := *m.emailReader
	er, cmd = er.update(msg)
	m.emailReader = &er

	if er.goBack {
		m.emailReader = nil
		m.view = viewEmailList
		return m, nil
	}

	return m, cmd
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

	switch m.view {
	case viewMailboxList:
		return m.mailboxList.view()
	case viewEmailList:
		if m.emailList != nil {
			return m.emailList.view()
		}
	case viewEmailReader:
		if m.emailReader != nil {
			return m.emailReader.view()
		}
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
