package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func makeThreadEmail(id, subject, fromName, fromEmail string) fastmail.Email {
	return fastmail.Email{
		ID:         id,
		ThreadID:   "thread1",
		Subject:    subject,
		From:       fastmail.EmailAddress{Name: fromName, Email: fromEmail},
		ReceivedAt: time.Now().Add(-1 * time.Hour),
		Preview:    "Preview of " + subject,
		Keywords:   []string{"$seen"},
	}
}

func TestNewThreadViewModel(t *testing.T) {
	email := makeThreadEmail("e1", "Test Subject", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)

	assert.True(t, tv.loading)
	assert.Contains(t, tv.list.Title, "Test Subject")
	assert.Equal(t, email, tv.email)
	assert.Nil(t, tv.viewing)
	assert.False(t, tv.ready)
	assert.False(t, tv.goBack)
}

func TestThreadItem_Title(t *testing.T) {
	email := fastmail.Email{
		From: fastmail.EmailAddress{Name: "Alice Smith", Email: "alice@example.com"},
	}
	item := threadItem{email: email}

	assert.Equal(t, "Alice Smith", item.Title())
}

func TestThreadItem_Title_NoName(t *testing.T) {
	email := fastmail.Email{
		From: fastmail.EmailAddress{Email: "bob@example.com"},
	}
	item := threadItem{email: email}

	assert.Equal(t, "bob@example.com", item.Title())
}

func TestThreadItem_Description(t *testing.T) {
	email := fastmail.Email{
		ReceivedAt: time.Now().Add(-2 * time.Hour),
		Preview:    "This is a preview of the email body",
	}
	item := threadItem{email: email}

	desc := item.Description()
	assert.Contains(t, desc, "2h")
	assert.Contains(t, desc, "This is a preview")
}

func TestThreadItem_FilterValue(t *testing.T) {
	email := fastmail.Email{
		Subject: "Important",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
	}
	item := threadItem{email: email}

	assert.Contains(t, item.FilterValue(), "Important")
	assert.Contains(t, item.FilterValue(), "Alice")
}

func TestThreadViewModel_Update_ThreadLoaded(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)

	emails := []fastmail.Email{
		makeThreadEmail("e1", "Test", "Alice", "alice@example.com"),
		makeThreadEmail("e2", "Re: Test", "Bob", "bob@example.com"),
	}

	tv, _ = tv.update(threadLoadedMsg{emails: emails})

	assert.False(t, tv.loading)
	assert.Len(t, tv.list.Items(), 2)
}

func TestThreadViewModel_HandleKey_Esc(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)
	tv.loading = false

	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.True(t, tv.goBack)
}

func TestThreadViewModel_HandleKey_Q(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)
	tv.loading = false

	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	assert.True(t, tv.goBack)
}

func TestThreadViewModel_HandleKey_Enter(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)

	// Load items first
	emails := []fastmail.Email{
		makeThreadEmail("e1", "Test", "Alice", "alice@example.com"),
		makeThreadEmail("e2", "Re: Test", "Bob", "bob@example.com"),
	}
	tv, _ = tv.update(threadLoadedMsg{emails: emails})

	// Press enter to select the first email
	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyEnter})

	require.NotNil(t, tv.viewing)
	assert.Equal(t, "e1", tv.viewing.ID)
	assert.True(t, tv.ready)
}

func TestThreadViewModel_HandleViewingKey_Esc(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)

	// Load and select
	emails := []fastmail.Email{makeThreadEmail("e1", "Test", "Alice", "alice@example.com")}
	tv, _ = tv.update(threadLoadedMsg{emails: emails})
	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, tv.viewing)

	// Press esc to go back to thread list
	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.Nil(t, tv.viewing)
	assert.False(t, tv.ready)
}

func TestThreadViewModel_HandleViewingKey_Q(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)

	emails := []fastmail.Email{makeThreadEmail("e1", "Test", "Alice", "alice@example.com")}
	tv, _ = tv.update(threadLoadedMsg{emails: emails})
	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, tv.viewing)

	tv, _ = tv.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	assert.Nil(t, tv.viewing)
}

func TestThreadViewModel_View_Loading(t *testing.T) {
	email := makeThreadEmail("e1", "Test Subject", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)

	v := tv.view()

	assert.Contains(t, v, "Loading thread")
	assert.Contains(t, v, "Test Subject")
}

func TestThreadViewModel_View_List(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)
	tv.setSize(80, 24)

	emails := []fastmail.Email{
		makeThreadEmail("e1", "Test", "Alice", "alice@example.com"),
	}
	tv, _ = tv.update(threadLoadedMsg{emails: emails})

	v := tv.view()

	// Should show the list, not loading
	assert.NotContains(t, v, "Loading thread")
}

func TestThreadViewModel_SetSize(t *testing.T) {
	email := makeThreadEmail("e1", "Test", "Alice", "alice@example.com")
	tv := newThreadViewModel(email)

	tv.setSize(100, 40)

	assert.Equal(t, 100, tv.width)
	assert.Equal(t, 40, tv.height)
}
