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
	viewMovePicker
)

// errMsg wraps errors from async commands.
type errMsg struct{ err error }

// connectedMsg signals successful client connection.
type connectedMsg struct{}

// Model is the top-level bubbletea model.
type Model struct {
	client       *fastmail.Client
	view         view
	mailboxList  mailboxListModel
	emailList    *emailListModel
	emailReader  *emailReaderModel
	movePicker   *movePickerModel
	actionSource view // which view initiated the current action
	width        int
	height       int
	err          error
	quit         bool
	connecting   bool
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
	case viewMovePicker:
		return m.movePicker != nil && m.movePicker.list.SettingFilter()
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
			// In reader/move picker views, q means "go back" — let the view handle it
			if m.view != viewEmailReader && m.view != viewMovePicker && !m.isFiltering() {
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
		if m.movePicker != nil {
			m.movePicker.setSize(msg.Width, msg.Height)
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

	case emailActionDoneMsg:
		return m.handleActionDone(msg)

	case emailActionErrMsg:
		return m.handleActionErr(msg)

	case mailboxesForMoveMsg:
		mp := newMovePickerModel(msg.emailID, msg.mailboxes)
		mp.setSize(m.width, m.height)
		m.movePicker = &mp
		m.view = viewMovePicker
		return m, nil

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
	case viewMovePicker:
		return m.updateMovePicker(msg)
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

	if el.action != nil {
		act := *el.action
		el.action = nil
		m.emailList = &el
		return m.dispatchAction(act, viewEmailList)
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

	if er.action != nil {
		act := *er.action
		er.action = nil
		m.emailReader = &er
		return m.dispatchAction(act, viewEmailReader)
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
	case viewMovePicker:
		if m.movePicker != nil {
			return m.movePicker.view()
		}
	}

	return ""
}

// dispatchAction fires the appropriate tea.Cmd for an email action.
func (m Model) dispatchAction(act emailAction, source view) (tea.Model, tea.Cmd) {
	m.actionSource = source
	switch act.kind {
	case "archive":
		return m, archiveEmailCmd(m.client, act.email.ID)
	case "delete":
		return m, deleteEmailCmd(m.client, act.email.ID)
	case "toggleRead":
		return m, toggleReadCmd(m.client, act.email.ID, act.email.IsRead())
	case "toggleFlag":
		return m, toggleFlagCmd(m.client, act.email.ID, act.email.IsFlagged())
	case "move":
		return m, m.fetchMailboxesForMoveCmd(act.email.ID)
	}
	return m, nil
}

// mailboxesForMoveMsg carries mailbox data and email ID for the move picker.
type mailboxesForMoveMsg struct {
	mailboxes []fastmail.Mailbox
	emailID   string
}

func (m Model) fetchMailboxesForMoveCmd(emailID string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		mailboxes, err := client.Mailbox().List(context.Background())
		if err != nil {
			return emailActionErrMsg{err: err, action: "Move"}
		}
		return mailboxesForMoveMsg{mailboxes: mailboxes, emailID: emailID}
	}
}

// handleActionDone processes a successful email action.
func (m Model) handleActionDone(msg emailActionDoneMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Remove/update item in the email list
	if m.emailList != nil {
		switch msg.action {
		case "Archived", "Deleted", "Moved":
			m.removeEmailFromList(msg.emailID)
		default:
			m.updateEmailInList(msg.emailID, msg.action)
		}
	}

	// If action was from reader and removed the email, go back to list
	if m.actionSource == viewEmailReader {
		switch msg.action {
		case "Archived", "Deleted", "Moved":
			m.emailReader = nil
			m.view = viewEmailList
		}
	}

	// Show status on the active view
	switch m.view {
	case viewEmailList:
		if m.emailList != nil {
			cmd = m.emailList.status.setStatus(msg.action, false)
		}
	case viewEmailReader:
		if m.emailReader != nil {
			cmd = m.emailReader.status.setStatus(msg.action, false)
		}
	case viewMailboxList, viewMovePicker:
		// No status to show on these views
	}

	return m, cmd
}

// handleActionErr processes a failed email action.
func (m Model) handleActionErr(msg emailActionErrMsg) (tea.Model, tea.Cmd) {
	errText := fmt.Sprintf("%s failed: %v", msg.action, msg.err)
	var cmd tea.Cmd

	switch m.actionSource {
	case viewEmailList:
		if m.emailList != nil {
			cmd = m.emailList.status.setStatus(errText, true)
		}
	case viewEmailReader:
		if m.emailReader != nil {
			cmd = m.emailReader.status.setStatus(errText, true)
		}
	case viewMailboxList, viewMovePicker:
		// Actions are not initiated from these views
	}

	return m, cmd
}

// removeEmailFromList removes an email item from the email list by ID.
func (m Model) removeEmailFromList(emailID string) {
	if m.emailList == nil {
		return
	}
	items := m.emailList.list.Items()
	for i, item := range items {
		if ei, ok := item.(emailItem); ok && ei.email.ID == emailID {
			m.emailList.list.RemoveItem(i)
			return
		}
	}
}

// updateEmailInList updates the keywords on an email item in the list.
func (m Model) updateEmailInList(emailID, action string) {
	if m.emailList == nil {
		return
	}
	items := m.emailList.list.Items()
	for i, item := range items {
		if ei, ok := item.(emailItem); ok && ei.email.ID == emailID {
			ei.email.Keywords = applyKeywordAction(ei.email, action)
			m.emailList.list.SetItem(i, ei)
			return
		}
	}
}

// applyKeywordAction returns updated keywords for an email based on the action.
func applyKeywordAction(email fastmail.Email, action string) []string {
	switch action {
	case "Marked read":
		if !email.IsRead() {
			return append(email.Keywords, fastmail.KeywordSeen)
		}
	case "Marked unread":
		return removeKeyword(email.Keywords, fastmail.KeywordSeen)
	case "Flagged":
		if !email.IsFlagged() {
			return append(email.Keywords, fastmail.KeywordFlagged)
		}
	case "Unflagged":
		return removeKeyword(email.Keywords, fastmail.KeywordFlagged)
	}
	return email.Keywords
}

// removeKeyword returns a copy of keywords with the specified keyword removed.
func removeKeyword(keywords []string, keyword string) []string {
	result := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if kw != keyword {
			result = append(result, kw)
		}
	}
	return result
}

func (m Model) updateMovePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.movePicker == nil {
		return m, nil
	}

	mp := *m.movePicker
	var cmd tea.Cmd
	mp, cmd = mp.update(msg)
	m.movePicker = &mp

	if mp.canceled {
		m.movePicker = nil
		m.view = m.actionSource
		return m, nil
	}

	if mp.selected != nil {
		mailbox := *mp.selected
		emailID := mp.emailID
		m.movePicker = nil
		m.view = m.actionSource
		return m, moveEmailCmd(m.client, emailID, mailbox.ID)
	}

	return m, cmd
}

// Run starts the TUI program.
func Run(client *fastmail.Client) error {
	m := New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
