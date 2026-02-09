package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestNew(t *testing.T) {
	client := fastmail.NewClient("https://api.fastmail.com", "test-token")
	m := New(client)

	assert.Equal(t, viewMailboxList, m.view)
	assert.NotNil(t, m.client)
	assert.True(t, m.connecting)
	assert.False(t, m.quit)
	assert.Nil(t, m.emailList)
}

func TestInit_ReturnsConnectCmd(t *testing.T) {
	m := New(nil)
	cmd := m.Init()

	assert.NotNil(t, cmd)
}

func TestUpdate_Quit(t *testing.T) {
	m := New(nil)
	m.connecting = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestUpdate_CtrlC(t *testing.T) {
	m := New(nil)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestUpdate_WindowSize(t *testing.T) {
	m := New(nil)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	assert.Equal(t, 120, model.width)
	assert.Equal(t, 40, model.height)
}

func TestUpdate_WindowSize_WithEmailList(t *testing.T) {
	m := New(nil)
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox"})
	m.emailList = &el

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model := updated.(Model)

	assert.Equal(t, 100, model.width)
}

func TestUpdate_Connected(t *testing.T) {
	m := New(nil)
	m.connecting = true

	updated, cmd := m.Update(connectedMsg{})
	model := updated.(Model)

	assert.False(t, model.connecting)
	assert.NotNil(t, cmd)
}

func TestUpdate_Error(t *testing.T) {
	m := New(nil)
	m.connecting = true

	updated, cmd := m.Update(errMsg{err: assert.AnError})
	model := updated.(Model)

	assert.False(t, model.connecting)
	assert.Equal(t, assert.AnError, model.err)
	assert.Nil(t, cmd)
}

func TestUpdate_MailboxesLoaded(t *testing.T) {
	m := New(nil)
	m.connecting = false

	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, TotalEmails: 42, UnreadEmails: 3},
		{ID: "2", Name: "Sent", Role: fastmail.RoleSent, TotalEmails: 10},
	}

	updated, _ := m.Update(mailboxesLoadedMsg{mailboxes: mailboxes})
	model := updated.(Model)

	assert.False(t, model.mailboxList.loading)
}

func TestUpdate_MailboxesLoaded_AutoSelectsInbox(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	mailboxes := []fastmail.Mailbox{
		{ID: "mb-sent", Name: "Sent", Role: fastmail.RoleSent},
		{ID: "mb-inbox", Name: "Inbox", Role: fastmail.RoleInbox},
	}

	updated, cmd := m.Update(mailboxesLoadedMsg{mailboxes: mailboxes})
	model := updated.(Model)

	// Inbox selection happens synchronously — no second Update needed
	assert.Equal(t, viewEmailList, model.view)
	require.NotNil(t, model.emailList)
	assert.Equal(t, "Inbox", model.emailList.list.Title)
	// cmd should include fetchEmailsCmd
	require.NotNil(t, cmd)
}

func TestUpdate_MailboxesLoaded_NoInbox_NoAutoSelect(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	mailboxes := []fastmail.Mailbox{
		{ID: "mb-sent", Name: "Sent", Role: fastmail.RoleSent},
		{ID: "mb-drafts", Name: "Drafts", Role: fastmail.RoleDrafts},
	}

	updated, _ := m.Update(mailboxesLoadedMsg{mailboxes: mailboxes})
	model := updated.(Model)

	assert.Nil(t, model.emailList)
	assert.Equal(t, viewMailboxList, model.view)
}

func TestUpdate_MailboxesLoaded_ExistingEmailList_NoAutoSelect(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	el := newEmailListModel(fastmail.Mailbox{Name: "Sent", ID: "mb-sent"})
	m.emailList = &el
	m.view = viewEmailList

	mailboxes := []fastmail.Mailbox{
		{ID: "mb-inbox", Name: "Inbox", Role: fastmail.RoleInbox},
	}

	updated, _ := m.Update(mailboxesLoadedMsg{mailboxes: mailboxes})
	model := updated.(Model)

	// Should keep the existing email list, not replace with Inbox
	assert.Equal(t, "Sent", model.emailList.list.Title)
}

func TestUpdate_MailboxSelected(t *testing.T) {
	m := New(nil)
	m.connecting = false

	mb := fastmail.Mailbox{ID: "mb1", Name: "Inbox"}
	updated, cmd := m.Update(mailboxSelectedMsg{mailbox: mb})
	model := updated.(Model)

	assert.Equal(t, viewEmailList, model.view)
	require.NotNil(t, model.emailList)
	assert.Equal(t, "Inbox", model.emailList.list.Title)
	assert.NotNil(t, cmd) // fetchEmailsCmd
}

func TestUpdate_EmailList_GoBack(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.loading = false
	m.emailList = &el

	// Esc should go back
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)

	assert.Equal(t, viewMailboxList, model.view)
	assert.Nil(t, model.emailList)
}

func TestView_Connecting(t *testing.T) {
	m := New(nil)
	m.connecting = true

	assert.Contains(t, m.View(), "Connecting")
}

func TestView_Quit(t *testing.T) {
	m := New(nil)
	m.quit = true

	assert.Empty(t, m.View())
}

func TestView_Error(t *testing.T) {
	m := New(nil)
	m.err = assert.AnError

	assert.Contains(t, m.View(), "Error:")
}

func TestView_MailboxList(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.mailboxList.loading = false
	m.mailboxList.setSize(20, 35)

	v := m.View()
	assert.Contains(t, v, "Mailboxes")
}

func TestView_EmailList(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.loading = false
	el.setSize(80, 24)
	m.emailList = &el

	v := m.View()
	assert.Contains(t, v, "Inbox")
}

func TestDispatchAction_ToggleFlag(t *testing.T) {
	m := New(nil)
	email := fastmail.Email{ID: "e1", Keywords: []string{}}
	act := emailAction{kind: "toggleFlag", email: email}

	_, cmd := m.dispatchAction(act, viewEmailList)

	assert.NotNil(t, cmd)
}

func TestHandleActionDone_Flagged(t *testing.T) {
	m := New(nil)
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.list.SetSize(80, 24)
	emails := []fastmail.Email{
		{ID: "e1", Subject: "Test", Keywords: []string{}},
	}
	el, _ = el.update(emailsLoadedMsg{emails: emails})
	m.emailList = &el
	m.view = viewEmailList

	result, _ := m.handleActionDone(emailActionDoneMsg{action: "Flagged", emailID: "e1"})
	updated := result.(Model)

	// Verify the email now has the $flagged keyword
	items := updated.emailList.list.Items()
	require.Len(t, items, 1)
	ei := items[0].(emailItem)
	assert.True(t, ei.email.IsFlagged())
}

func TestHandleActionDone_Unflagged(t *testing.T) {
	m := New(nil)
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.list.SetSize(80, 24)
	emails := []fastmail.Email{
		{ID: "e1", Subject: "Test", Keywords: []string{"$flagged", "$seen"}},
	}
	el, _ = el.update(emailsLoadedMsg{emails: emails})
	m.emailList = &el
	m.view = viewEmailList

	result, _ := m.handleActionDone(emailActionDoneMsg{action: "Unflagged", emailID: "e1"})
	updated := result.(Model)

	// Verify the email no longer has the $flagged keyword
	items := updated.emailList.list.Items()
	require.Len(t, items, 1)
	ei := items[0].(emailItem)
	assert.False(t, ei.email.IsFlagged())
	// $seen should still be present
	assert.True(t, ei.email.IsRead())
}

func TestUpdate_QuestionMark_ShowsHelp(t *testing.T) {
	m := New(nil)
	m.connecting = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model := updated.(Model)

	assert.True(t, model.helpOverlay)
	assert.Nil(t, cmd)
}

func TestUpdate_HelpOverlay_DismissOnKeypress(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.helpOverlay = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model := updated.(Model)

	assert.False(t, model.helpOverlay)
	assert.Nil(t, cmd)
}

func TestUpdate_HelpOverlay_CtrlCStillQuits(t *testing.T) {
	m := New(nil)
	m.helpOverlay = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestView_HelpOverlay(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.helpOverlay = true
	m.view = viewMailboxList

	v := m.View()

	assert.Contains(t, v, "Mailbox List")
	assert.Contains(t, v, "Open mailbox")
}

func TestUpdate_ThreadView_GoBack(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewThreadView

	email := fastmail.Email{ID: "e1", Subject: "Test", ThreadID: "t1"}
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)
	tv.loading = false
	m.threadView = &tv

	// Press esc to go back to email reader
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)

	assert.Equal(t, viewEmailReader, model.view)
	assert.Nil(t, model.threadView)
}

func TestUpdate_QuestionMark_IgnoredInEmailReader(t *testing.T) {
	// In the email reader, '?' should still show help (not filtering)
	m := New(nil)
	m.connecting = false
	m.view = viewEmailReader
	email := fastmail.Email{ID: "e1", Subject: "Test"}
	er := newEmailReaderModel(email)
	er.setSize(80, 24)
	m.emailReader = &er

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model := updated.(Model)

	assert.True(t, model.helpOverlay)
	assert.Nil(t, cmd)
}

func TestUpdate_SearchEmailsCmd(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.list.SetSize(80, 24)

	emails := makeTestEmails()
	el, _ = el.update(emailsLoadedMsg{emails: emails})

	// Set up search signal
	query := "test query"
	el.search = &query
	el.loading = true
	m.emailList = &el

	// Call updateEmailList — it should pick up the search signal
	result, cmd := m.updateEmailList(tea.Msg(nil))
	model := result.(Model)

	// search signal should be consumed
	assert.Nil(t, model.emailList.search)
	// a command should be returned (the searchEmailsCmd)
	assert.NotNil(t, cmd)
}

func TestUpdate_AttachmentPicker_Cancel(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewAttachmentPicker

	// Set up an email reader to return to
	email := fastmail.Email{ID: "e1", Subject: "Test"}
	er := newEmailReaderModel(email)
	er.setSize(80, 24)
	m.emailReader = &er

	// Set up attachment picker
	attachments := []fastmail.Attachment{
		{BlobID: "b1", Name: "file.pdf", Type: "application/pdf", Size: 1024},
	}
	ap := newAttachmentPickerModel("e1", attachments)
	ap.setSize(80, 24)
	m.attachmentPicker = &ap

	// Press esc to cancel
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)

	assert.Equal(t, viewEmailReader, model.view)
	assert.Nil(t, model.attachmentPicker)
}

func TestUpdate_AttachmentDownloaded(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailReader

	email := fastmail.Email{ID: "e1", Subject: "Test"}
	er := newEmailReaderModel(email)
	er.setSize(80, 24)
	m.emailReader = &er

	updated, cmd := m.Update(attachmentDownloadedMsg{name: "report.pdf", path: "/tmp/report.pdf"})
	model := updated.(Model)

	// Status message should be set on the reader
	require.NotNil(t, model.emailReader)
	assert.True(t, model.emailReader.status.visible)
	assert.Contains(t, model.emailReader.status.message, "Downloaded report.pdf")
	assert.NotNil(t, cmd) // tick cmd for status clear
}

func TestUpdate_AttachmentDownloadErr(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailReader

	email := fastmail.Email{ID: "e1", Subject: "Test"}
	er := newEmailReaderModel(email)
	er.setSize(80, 24)
	m.emailReader = &er

	updated, cmd := m.Update(attachmentDownloadErrMsg{err: assert.AnError, name: "report.pdf"})
	model := updated.(Model)

	require.NotNil(t, model.emailReader)
	assert.True(t, model.emailReader.status.visible)
	assert.Contains(t, model.emailReader.status.message, "Download failed")
	assert.NotNil(t, cmd)
}

func TestUpdate_EmailReader_FlagAction(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailReader
	email := fastmail.Email{ID: "e1", Subject: "Test", Keywords: []string{"$flagged"}}
	er := newEmailReaderModel(email)
	er.setSize(80, 24)
	m.emailReader = &er

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	model := updated.(Model)

	require.NotNil(t, model.emailReader)
	// After dispatching, the action should have been consumed and dispatched
	// The reader's action field is cleared after dispatch in updateEmailReader
	// We can verify the action source was set correctly
	assert.Equal(t, viewEmailReader, model.actionSource)
}

func TestUpdate_MailboxSelected_UsesLayoutConstrainedSize(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	mb := fastmail.Mailbox{ID: "mb1", Name: "Inbox"}
	updated, _ := m.Update(mailboxSelectedMsg{mailbox: mb})
	model := updated.(Model)

	require.NotNil(t, model.emailList)
	layout := model.panes.computeLayout(120, 40)
	// Email list should use layout-constrained width, not full terminal width.
	// The bubbles list Width reflects the constrained size.
	assert.Equal(t, layout.mainWidth-2, model.emailList.list.Width())
}

func TestNew_HasPaneManager(t *testing.T) {
	m := New(nil)
	assert.Equal(t, PaneEmailList, m.panes.focus)
	assert.True(t, m.panes.sidebar)
}

func TestNew_HasTheme(t *testing.T) {
	m := New(nil)
	assert.NotEqual(t, lipgloss.Color(""), m.theme.FocusBorder)
}

func TestUpdate_Tab_CyclesFocus(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.panes.focus = PaneMailbox

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := updated.(Model)

	assert.Equal(t, PaneEmailList, model.panes.focus)
}

func TestUpdate_B_TogglesSidebar(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	m.emailList = &el

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	model := updated.(Model)

	assert.False(t, model.panes.sidebar)
}

func TestUpdate_Plus_AdjustsSplit(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	m.emailList = &el
	m.panes.splitPct = 50

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	model := updated.(Model)

	assert.Equal(t, 60, model.panes.splitPct)
}

func TestUpdate_Minus_AdjustsSplit(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	m.emailList = &el
	m.panes.splitPct = 50

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	model := updated.(Model)

	assert.Equal(t, 40, model.panes.splitPct)
}

func TestUpdate_EmailSelected_OpensPreview(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.loading = false
	m.emailList = &el

	email := fastmail.Email{ID: "e1", Subject: "Test"}
	el.selected = &email
	m.emailList = &el

	// When email is selected, preview pane should be populated
	result, cmd := m.updateEmailList(tea.Msg(nil))
	model := result.(Model)

	// Preview should be initialized
	require.NotNil(t, model.emailReader)
	// Should still be in the dashboard layout (not fullscreen reader)
	assert.Equal(t, viewEmailList, model.view)
	// Should fetch the body
	assert.NotNil(t, cmd)
}

func TestUpdate_Enter_InPreview_OpensFullscreen(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	m.panes.focus = PanePreview
	email := fastmail.Email{ID: "e1", Subject: "Test"}
	er := newEmailReaderModel(email)
	er.setSize(80, 20)
	m.emailReader = &er

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	assert.Equal(t, viewEmailReader, model.view)
}

func TestView_StatsBar_Visible(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.statsBar.unreadCount = 42
	m.statsBar.flaggedCount = 5

	v := m.View()
	assert.Contains(t, v, "42")
	assert.Contains(t, v, "5")
}

func TestUpdate_MailboxesLoaded_UpdatesStats(t *testing.T) {
	m := New(nil)
	m.connecting = false
	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", UnreadEmails: 10},
		{ID: "2", Name: "Work", UnreadEmails: 5},
	}

	updated, _ := m.Update(mailboxesLoadedMsg{mailboxes: mailboxes})
	model := updated.(Model)

	assert.Equal(t, uint64(15), model.statsBar.unreadCount)
}

func TestView_ComposeOverlay_RendersOnTopOfLayout(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.view = viewCompose
	cm := newComposeModel()
	cm.setSize(120, 40)
	m.composeView = &cm

	v := m.View()
	assert.Contains(t, v, "To:")
}

func TestView_HelpOverlay_RendersOnTopOfLayout(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.helpOverlay = true

	v := m.View()
	assert.Contains(t, v, "help")
}

func TestUpdate_ThreadView_FromPreview_StaysInLayout(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.view = viewEmailList
	m.panes.focus = PanePreview

	email := fastmail.Email{ID: "e1", ThreadID: "t1"}
	er := newEmailReaderModel(email)
	er.setSize(80, 20)
	er.showThread = true
	m.emailReader = &er

	// updateEmailReader should detect showThread and create threadView
	result, cmd := m.updateEmailReader(tea.Msg(nil))
	model := result.(Model)

	require.NotNil(t, model.threadView)
	assert.NotNil(t, cmd) // fetchThreadCmd
}
