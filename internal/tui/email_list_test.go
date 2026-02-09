package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func makeTestEmails() []fastmail.Email {
	return []fastmail.Email{
		{
			ID:         "e1",
			Subject:    "Hello World",
			From:       fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
			ReceivedAt: time.Now().Add(-2 * time.Hour),
			Preview:    "This is a test email",
			Keywords:   []string{},
		},
		{
			ID:         "e2",
			Subject:    "Meeting Tomorrow",
			From:       fastmail.EmailAddress{Email: "bob@example.com"},
			ReceivedAt: time.Now().Add(-3 * 24 * time.Hour),
			Preview:    "Don't forget the meeting",
			Keywords:   []string{"$seen"},
		},
		{
			ID:         "e3",
			Subject:    "Important Update",
			From:       fastmail.EmailAddress{Name: "Carol", Email: "carol@example.com"},
			ReceivedAt: time.Now().Add(-30 * time.Minute),
			Preview:    "Please review ASAP",
			Keywords:   []string{"$flagged"},
		},
	}
}

func TestEmailItem_Title_Unread(t *testing.T) {
	email := fastmail.Email{
		Subject: "Test",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
	}
	item := emailItem{email: email}

	title := item.Title()
	assert.Contains(t, title, "*")
	assert.Contains(t, title, "Alice")
	assert.Contains(t, title, "Test")
}

func TestEmailItem_Title_Read(t *testing.T) {
	email := fastmail.Email{
		Subject:  "Test",
		From:     fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
		Keywords: []string{"$seen"},
	}
	item := emailItem{email: email}

	title := item.Title()
	assert.NotContains(t, title, "*")
	assert.Contains(t, title, "Alice")
}

func TestEmailItem_Title_Flagged(t *testing.T) {
	email := fastmail.Email{
		Subject:  "Important",
		From:     fastmail.EmailAddress{Name: "Bob", Email: "bob@example.com"},
		Keywords: []string{"$flagged"},
	}
	item := emailItem{email: email}

	assert.Contains(t, item.Title(), "!")
}

func TestEmailItem_Title_NoName(t *testing.T) {
	email := fastmail.Email{
		Subject: "Test",
		From:    fastmail.EmailAddress{Email: "alice@example.com"},
	}
	item := emailItem{email: email}

	assert.Contains(t, item.Title(), "alice@example.com")
}

func TestEmailItem_Description(t *testing.T) {
	email := fastmail.Email{
		ReceivedAt: time.Now().Add(-2 * time.Hour),
		Preview:    "Hello there",
	}
	item := emailItem{email: email}

	desc := item.Description()
	assert.Contains(t, desc, "2h ago")
	assert.Contains(t, desc, "Hello there")
}

func TestEmailItem_FilterValue(t *testing.T) {
	email := fastmail.Email{
		Subject: "Meeting",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
	}
	item := emailItem{email: email}

	assert.Contains(t, item.FilterValue(), "Meeting")
	assert.Contains(t, item.FilterValue(), "Alice")
}

func TestNewEmailListModel(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)

	assert.True(t, m.loading)
	assert.Nil(t, m.selected)
	assert.False(t, m.goBack)
	assert.Equal(t, "Inbox", m.list.Title)
}

func TestEmailListModel_Update_EmailsLoaded(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)

	emails := makeTestEmails()
	m, _ = m.update(emailsLoadedMsg{emails: emails})

	assert.False(t, m.loading)
	assert.Equal(t, 3, len(m.list.Items()))
}

func TestEmailListModel_Update_Esc(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.loading = false

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.True(t, m.goBack)
}

func TestEmailListModel_Update_EnterSelectsEmail(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)

	emails := makeTestEmails()
	m, _ = m.update(emailsLoadedMsg{emails: emails})
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEnter})

	require.NotNil(t, m.selected)
	assert.Equal(t, "e1", m.selected.ID)
}

func TestEmailListModel_View_Loading(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)

	assert.Contains(t, m.view(), "Loading")
	assert.Contains(t, m.view(), "Inbox")
}

func TestEmailsToItems(t *testing.T) {
	emails := makeTestEmails()
	items := emailsToItems(emails)

	assert.Len(t, items, 3)
	assert.Equal(t, "e1", items[0].(emailItem).email.ID)
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		t        time.Time
		contains string
	}{
		{"zero", time.Time{}, ""},
		{"just now", time.Now().Add(-30 * time.Second), "just now"},
		{"minutes", time.Now().Add(-5 * time.Minute), "5m ago"},
		{"hours", time.Now().Add(-3 * time.Hour), "3h ago"},
		{"days", time.Now().Add(-2 * 24 * time.Hour), "2d ago"},
		{"older", time.Now().Add(-30 * 24 * time.Hour), "Jan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.t)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hel…", truncate("hello world", 4))
	assert.Equal(t, "", truncate("", 5))
}

func TestEmailListModel_HandleKey_FlagSetsAction(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)

	emails := makeTestEmails()
	m, _ = m.update(emailsLoadedMsg{emails: emails})

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})

	require.NotNil(t, m.action)
	assert.Equal(t, "toggleFlag", m.action.kind)
	assert.Equal(t, "e1", m.action.email.ID)
}
