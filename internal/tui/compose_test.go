package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestNewComposeModel(t *testing.T) {
	m := newComposeModel()

	assert.Equal(t, fieldTo, m.activeField)
	assert.True(t, m.toInput.Focused())
	assert.False(t, m.subjectInput.Focused())
	assert.False(t, m.bodyInput.Focused())
	assert.False(t, m.canceled)
	assert.False(t, m.send)
	assert.Empty(t, m.err)
}

func TestComposeModel_Tab_Navigation(t *testing.T) {
	m := newComposeModel()

	// To -> Subject
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, fieldSubject, m.activeField)
	assert.True(t, m.subjectInput.Focused())
	assert.False(t, m.toInput.Focused())

	// Subject -> Body
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, fieldBody, m.activeField)
	assert.True(t, m.bodyInput.Focused())
	assert.False(t, m.subjectInput.Focused())

	// Body -> To (wrap around)
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, fieldTo, m.activeField)
	assert.True(t, m.toInput.Focused())
	assert.False(t, m.bodyInput.Focused())
}

func TestComposeModel_ShiftTab_Navigation(t *testing.T) {
	m := newComposeModel()

	// To -> Body (wrap backwards)
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, fieldBody, m.activeField)
	assert.True(t, m.bodyInput.Focused())
	assert.False(t, m.toInput.Focused())

	// Body -> Subject
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, fieldSubject, m.activeField)
	assert.True(t, m.subjectInput.Focused())
	assert.False(t, m.bodyInput.Focused())

	// Subject -> To
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, fieldTo, m.activeField)
	assert.True(t, m.toInput.Focused())
	assert.False(t, m.subjectInput.Focused())
}

func TestComposeModel_Esc_Cancels(t *testing.T) {
	m := newComposeModel()

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.True(t, m.canceled)
}

func TestComposeModel_CtrlS_Send(t *testing.T) {
	m := newComposeModel()
	m.toInput.SetValue("alice@example.com")

	// Tab to subject, set value
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyTab})
	m.subjectInput.SetValue("Hello")

	// Ctrl+S to send
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyCtrlS})

	assert.True(t, m.send)
	assert.Empty(t, m.err)
}

func TestComposeModel_CtrlS_NoTo(t *testing.T) {
	m := newComposeModel()
	// Leave To empty, set subject
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyTab})
	m.subjectInput.SetValue("Hello")

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyCtrlS})

	assert.False(t, m.send)
	assert.Contains(t, m.err, "To")
}

func TestComposeModel_CtrlS_NoSubject(t *testing.T) {
	m := newComposeModel()
	m.toInput.SetValue("alice@example.com")
	// Leave Subject empty

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyCtrlS})

	assert.False(t, m.send)
	assert.Contains(t, m.err, "Subject")
}

func TestComposeModel_View(t *testing.T) {
	m := newComposeModel()

	v := m.view()

	assert.Contains(t, v, "New Email")
	assert.Contains(t, v, "ctrl+s send")
	assert.Contains(t, v, "esc cancel")
	assert.Contains(t, v, "tab")
}

func TestComposeModel_View_WithError(t *testing.T) {
	m := newComposeModel()
	m.err = "To field is required"

	v := m.view()

	assert.Contains(t, v, "To field is required")
}

func TestComposeModel_Accessors(t *testing.T) {
	m := newComposeModel()
	m.toInput.SetValue("  alice@example.com  ")
	m.subjectInput.SetValue("  Hello World  ")
	m.bodyInput.SetValue("Body text here")

	assert.Equal(t, "alice@example.com", m.toAddress())
	assert.Equal(t, "Hello World", m.subject())
	assert.Equal(t, "Body text here", m.body())
}

func TestComposeModel_SetSize(t *testing.T) {
	m := newComposeModel()
	m.setSize(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestEmailListModel_HandleKey_CSetsCompose(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)
	m.loading = false

	emails := makeTestEmails()
	m, _ = m.update(emailsLoadedMsg{emails: emails})

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	assert.True(t, m.compose)
}

func TestUpdate_Compose_Cancel(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewCompose

	cm := newComposeModel()
	cm.setSize(80, 24)
	m.composeView = &cm

	// Press esc to cancel
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)

	assert.Equal(t, viewEmailList, model.view)
	assert.Nil(t, model.composeView)
}

func TestUpdate_EmailSent(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewCompose

	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.list.SetSize(80, 24)
	el.loading = false
	m.emailList = &el

	cm := newComposeModel()
	m.composeView = &cm

	updated, _ := m.Update(emailSentMsg{emailID: "new-email-123"})
	model := updated.(Model)

	assert.Equal(t, viewEmailList, model.view)
	assert.Nil(t, model.composeView)
	require.NotNil(t, model.emailList)
	assert.True(t, model.emailList.status.visible)
	assert.Contains(t, model.emailList.status.message, "Email sent")
}

func TestUpdate_EmailSendErr(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewCompose

	cm := newComposeModel()
	m.composeView = &cm

	updated, _ := m.Update(emailSendErrMsg{err: assert.AnError})
	model := updated.(Model)

	require.NotNil(t, model.composeView)
	assert.Contains(t, model.composeView.err, "Send failed")
}

func TestUpdate_Compose_QuestionMark_PassesThrough(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewCompose

	cm := newComposeModel()
	cm.setSize(80, 24)
	m.composeView = &cm

	// '?' should NOT trigger help overlay in compose mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model := updated.(Model)

	assert.False(t, model.helpOverlay)
}

func TestUpdate_Compose_Q_PassesThrough(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewCompose

	cm := newComposeModel()
	cm.setSize(80, 24)
	m.composeView = &cm

	// 'q' should NOT quit in compose mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model := updated.(Model)

	assert.False(t, model.quit)
}
