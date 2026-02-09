// Package tui implements an interactive terminal UI using bubbletea.
package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	viewThreadView
	viewAttachmentPicker
	viewCompose
)

// errMsg wraps errors from async commands.
type errMsg struct{ err error }

// connectedMsg signals successful client connection.
type connectedMsg struct{}

// Model is the top-level bubbletea model.
type Model struct {
	client           *fastmail.Client
	view             view
	mailboxList      mailboxListModel
	emailList        *emailListModel
	emailReader      *emailReaderModel
	movePicker       *movePickerModel
	threadView       *threadViewModel
	attachmentPicker *attachmentPickerModel
	composeView      *composeModel
	actionSource     view // which view initiated the current action
	width            int
	height           int
	err              error
	quit             bool
	connecting       bool
	helpOverlay      bool
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

func (m Model) fetchThreadCmd(threadID string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		emails, err := client.Mail().GetThread(context.Background(), threadID)
		if err != nil {
			return errMsg{err: err}
		}
		return threadLoadedMsg{emails: emails}
	}
}

func (m Model) searchEmailsCmd(query string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		emails, err := client.Mail().Search(context.Background(), query, emailPageSize)
		if err != nil {
			return errMsg{err: err}
		}
		return searchResultsMsg{emails: emails}
	}
}

// isFiltering returns true if any active list is in filter mode.
func (m Model) isFiltering() bool {
	switch m.view {
	case viewMailboxList:
		return m.mailboxList.list.SettingFilter()
	case viewEmailList:
		return m.emailList != nil && (m.emailList.list.SettingFilter() || m.emailList.searchMode)
	case viewEmailReader:
		return false
	case viewMovePicker:
		return m.movePicker != nil && m.movePicker.list.SettingFilter()
	case viewThreadView:
		return false
	case viewAttachmentPicker:
		return false
	case viewCompose:
		return true
	}
	return false
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if result, cmd, handled := m.handleGlobalKeys(msg); handled {
			return result, cmd
		}

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)

	case threadLoadedMsg:
		if m.threadView != nil {
			tv := *m.threadView
			tv, _ = tv.update(msg)
			m.threadView = &tv
		}
		return m, nil

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

	case attachmentDownloadedMsg:
		if m.emailReader != nil {
			return m, m.emailReader.status.setStatus(fmt.Sprintf("Downloaded %s", msg.name), false)
		}
		return m, nil

	case attachmentDownloadErrMsg:
		if m.emailReader != nil {
			return m, m.emailReader.status.setStatus(fmt.Sprintf("Download failed: %v", msg.err), true)
		}
		return m, nil

	case emailSentMsg:
		return m.handleEmailSent(msg)

	case emailSendErrMsg:
		return m.handleEmailSendErr(msg)

	case errMsg:
		m.err = msg.err
		m.connecting = false
		return m, nil
	}

	// Delegate to current view
	return m.updateView(msg)
}

// handleWindowSize propagates a resize to all sub-models.
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) {
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
	if m.threadView != nil {
		m.threadView.setSize(msg.Width, msg.Height)
	}
	if m.attachmentPicker != nil {
		m.attachmentPicker.setSize(msg.Width, msg.Height)
	}
	if m.composeView != nil {
		m.composeView.setSize(msg.Width, msg.Height)
	}
}

// handleGlobalKeys processes top-level keybindings before view delegation.
// Returns (model, cmd, true) if the key was handled.
func (m Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if msg.String() == "ctrl+c" {
		m.quit = true
		return m, tea.Quit, true
	}
	// Dismiss help overlay on any other keypress
	if m.helpOverlay {
		m.helpOverlay = false
		return m, nil, true
	}
	switch msg.String() {
	case "q":
		// In reader/move picker/thread/attachment/compose views, q means "go back" or is a character — let the view handle it
		if m.view != viewEmailReader && m.view != viewMovePicker && m.view != viewThreadView && m.view != viewAttachmentPicker && m.view != viewCompose && !m.isFiltering() {
			m.quit = true
			return m, tea.Quit, true
		}
	case "?":
		if !m.isFiltering() && m.view != viewCompose {
			m.helpOverlay = true
			return m, nil, true
		}
	}
	return m, nil, false
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
	case viewThreadView:
		return m.updateThreadView(msg)
	case viewAttachmentPicker:
		return m.updateAttachmentPicker(msg)
	case viewCompose:
		return m.updateCompose(msg)
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

	if el.compose {
		el.compose = false
		m.emailList = &el
		cm := newComposeModel()
		cm.setSize(m.width, m.height)
		m.composeView = &cm
		m.view = viewCompose
		return m, nil
	}

	if el.search != nil {
		query := *el.search
		el.search = nil
		m.emailList = &el
		return m, m.searchEmailsCmd(query)
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

	if er.showThread {
		er.showThread = false
		m.emailReader = &er
		tv := newThreadViewModel(er.email)
		tv.setSize(m.width, m.height)
		m.threadView = &tv
		m.view = viewThreadView
		return m, m.fetchThreadCmd(er.email.ThreadID)
	}

	if er.showAttachments {
		er.showAttachments = false
		m.emailReader = &er
		ap := newAttachmentPickerModel(er.email.ID, er.email.Attachments)
		ap.setSize(m.width, m.height)
		m.attachmentPicker = &ap
		m.view = viewAttachmentPicker
		return m, nil
	}

	if er.reply || er.replyAll {
		replyAll := er.replyAll
		er.reply = false
		er.replyAll = false
		m.emailReader = &er
		cm := newReplyComposeModel(er.email, replyAll)
		cm.setSize(m.width, m.height)
		m.composeView = &cm
		m.view = viewCompose
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

func (m Model) updateThreadView(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.threadView == nil {
		return m, nil
	}

	tv := *m.threadView
	var cmd tea.Cmd
	tv, cmd = tv.update(msg)
	m.threadView = &tv

	if tv.goBack {
		m.threadView = nil
		m.view = viewEmailReader
		return m, nil
	}

	return m, cmd
}

func (m Model) updateAttachmentPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.attachmentPicker == nil {
		return m, nil
	}

	ap := *m.attachmentPicker
	var cmd tea.Cmd
	ap, cmd = ap.update(msg)
	m.attachmentPicker = &ap

	if ap.canceled {
		m.attachmentPicker = nil
		m.view = viewEmailReader
		return m, nil
	}

	if ap.selected != nil {
		att := *ap.selected
		m.attachmentPicker = nil
		m.view = viewEmailReader
		return m, m.downloadAttachmentCmd(att)
	}

	return m, cmd
}

func (m Model) downloadAttachmentCmd(att fastmail.Attachment) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		reader, err := client.Mail().DownloadAttachment(context.Background(), att.BlobID, att.Name)
		if err != nil {
			return attachmentDownloadErrMsg{err: err, name: att.Name}
		}
		defer func() { _ = reader.Close() }()

		home, err := os.UserHomeDir()
		if err != nil {
			return attachmentDownloadErrMsg{err: err, name: att.Name}
		}

		path := filepath.Join(home, "Downloads", att.Name)

		f, err := os.Create(path) //nolint:gosec // User-initiated download to known directory
		if err != nil {
			return attachmentDownloadErrMsg{err: err, name: att.Name}
		}

		if _, copyErr := io.Copy(f, reader); copyErr != nil {
			_ = f.Close()
			return attachmentDownloadErrMsg{err: copyErr, name: att.Name}
		}

		if closeErr := f.Close(); closeErr != nil {
			return attachmentDownloadErrMsg{err: closeErr, name: att.Name}
		}

		return attachmentDownloadedMsg{name: att.Name, path: path}
	}
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

	if m.helpOverlay {
		content := helpForView(m.view)
		return "\n  " + strings.ReplaceAll(content, "\n", "\n  ")
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
	case viewThreadView:
		if m.threadView != nil {
			return m.threadView.view()
		}
	case viewAttachmentPicker:
		if m.attachmentPicker != nil {
			return m.attachmentPicker.view()
		}
	case viewCompose:
		if m.composeView != nil {
			return m.composeView.view()
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
	case viewMailboxList, viewMovePicker, viewThreadView, viewAttachmentPicker, viewCompose:
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
	case viewMailboxList, viewMovePicker, viewThreadView, viewAttachmentPicker, viewCompose:
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

func (m Model) updateCompose(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.composeView == nil {
		return m, nil
	}

	cm := *m.composeView
	var cmd tea.Cmd
	cm, cmd = cm.update(msg)
	m.composeView = &cm

	if cm.canceled {
		m.composeView = nil
		if cm.replyEmailID != "" && m.emailReader != nil {
			m.view = viewEmailReader
		} else {
			m.view = viewEmailList
		}
		return m, nil
	}

	if cm.send {
		cm.send = false
		m.composeView = &cm
		if cm.replyEmailID != "" {
			return m, m.replyEmailCmd(cm)
		}
		return m, m.sendEmailCmd(cm)
	}

	return m, cmd
}

func (m Model) handleEmailSent(_ emailSentMsg) (tea.Model, tea.Cmd) {
	wasReply := m.composeView != nil && m.composeView.replyEmailID != ""
	m.composeView = nil
	if wasReply && m.emailReader != nil {
		m.view = viewEmailReader
		return m, m.emailReader.status.setStatus("Reply sent", false)
	}
	m.view = viewEmailList
	if m.emailList != nil {
		return m, m.emailList.status.setStatus("Email sent", false)
	}
	return m, nil
}

func (m Model) handleEmailSendErr(msg emailSendErrMsg) (tea.Model, tea.Cmd) {
	if m.composeView != nil {
		m.composeView.err = fmt.Sprintf("Send failed: %v", msg.err)
	}
	return m, nil
}

func (m Model) sendEmailCmd(cm composeModel) tea.Cmd {
	client := m.client
	to := cm.toAddress()
	subject := cm.subject()
	body := cm.body()
	return func() tea.Msg {
		opts := fastmail.SendOptions{
			To:      []fastmail.EmailAddress{{Email: to}},
			Subject: subject,
			Body:    body,
		}
		emailID, err := client.Mail().Send(context.Background(), opts)
		if err != nil {
			return emailSendErrMsg{err: err}
		}
		return emailSentMsg{emailID: emailID}
	}
}

func (m Model) replyEmailCmd(cm composeModel) tea.Cmd {
	client := m.client
	emailID := cm.replyEmailID
	body := cm.body()
	replyAll := cm.replyAll
	return func() tea.Msg {
		opts := fastmail.ReplyOptions{
			EmailID:  emailID,
			Body:     body,
			ReplyAll: replyAll,
		}
		replyID, err := client.Mail().Reply(context.Background(), opts)
		if err != nil {
			return emailSendErrMsg{err: err}
		}
		return emailSentMsg{emailID: replyID}
	}
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
