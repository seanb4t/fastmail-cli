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

func TestEmailListModel_HandleKey_SlashEntersSearchMode(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.loading = false

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	assert.True(t, m.searchMode)
}

func TestEmailListModel_SearchInput_EscCancels(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.loading = false
	m.searchMode = true
	m.searchInput.Focus()

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.False(t, m.searchMode)
	assert.Empty(t, m.searchInput.Value())
}

func TestEmailListModel_SearchInput_EnterWithQuery(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)

	emails := makeTestEmails()
	m, _ = m.update(emailsLoadedMsg{emails: emails})

	// Enter search mode
	m.searchMode = true
	m.searchInput.Focus()
	m.searchInput.SetValue("test query")

	// Press enter
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, m.searchMode)
	assert.True(t, m.loading)
	assert.Equal(t, "test query", m.searchQuery)
	require.NotNil(t, m.search)
	assert.Equal(t, "test query", *m.search)
	// Original items should be saved
	assert.NotNil(t, m.savedItems)
}

func TestEmailListModel_SearchInput_EnterEmpty(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.loading = false
	m.searchMode = true
	m.searchInput.Focus()
	m.searchInput.SetValue("")

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, m.searchMode)
	assert.Nil(t, m.search)
}

func TestEmailListModel_EscClearsSearchResults(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)

	// Load original emails
	emails := makeTestEmails()
	m, _ = m.update(emailsLoadedMsg{emails: emails})
	originalCount := len(m.list.Items())

	// Simulate search results being active
	m.searchActive = true
	m.searchQuery = "test"
	m.savedItems = m.list.Items()
	m.list.Title = "Search: test"

	// Set search results (fewer items)
	searchResults := []fastmail.Email{emails[0]}
	searchItems := emailsToItems(searchResults)
	m.list.SetItems(searchItems)

	// Press esc to clear search
	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.False(t, m.searchActive)
	assert.Empty(t, m.searchQuery)
	assert.Equal(t, mb.Name, m.list.Title)
	assert.Equal(t, originalCount, len(m.list.Items()))
	assert.Nil(t, m.savedItems)
}

func TestEmailListModel_View_SearchMode(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)
	m.loading = false
	m.searchMode = true
	m.searchInput.Focus()

	v := m.view()
	assert.Contains(t, v, "Search emails")
}

func TestEmailListModel_View_SearchLoading(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.loading = true
	m.searchQuery = "important"

	v := m.view()
	assert.Contains(t, v, "Searching for")
	assert.Contains(t, v, `"important"`)
}

func TestEmailListModel_Update_SearchResults(t *testing.T) {
	mb := fastmail.Mailbox{Name: "Inbox", ID: "mb1"}
	m := newEmailListModel(mb)
	m.list.SetSize(80, 24)
	m.loading = true
	m.searchQuery = "test"

	results := []fastmail.Email{
		{ID: "s1", Subject: "Search Result"},
	}
	m, _ = m.update(searchResultsMsg{emails: results})

	assert.False(t, m.loading)
	assert.True(t, m.searchActive)
	assert.Contains(t, m.list.Title, "Search:")
	assert.Equal(t, 1, len(m.list.Items()))
}
