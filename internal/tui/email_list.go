package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

const (
	emailPageSize = 50
	keyEsc        = "esc"
	keyEnter      = "enter"
)

// emailItem implements list.DefaultItem for the bubbles list component.
type emailItem struct {
	email fastmail.Email
}

func (i emailItem) Title() string {
	prefix := " "
	if !i.email.IsRead() {
		prefix = "*"
	}
	if i.email.IsFlagged() {
		prefix = "!"
	}

	from := i.email.From.Email
	if i.email.From.Name != "" {
		from = i.email.From.Name
	}

	return fmt.Sprintf("%s %s — %s", prefix, from, i.email.Subject)
}

func (i emailItem) Description() string {
	age := formatAge(i.email.ReceivedAt)
	return fmt.Sprintf("%s  %s", age, truncate(i.email.Preview, 80))
}

func (i emailItem) FilterValue() string {
	return i.email.Subject + " " + i.email.From.String()
}

// searchResultsMsg carries emails returned from a search query.
type searchResultsMsg struct {
	emails []fastmail.Email
}

// emailListModel manages the email list view for a selected mailbox.
type emailListModel struct {
	list          list.Model
	mailbox       fastmail.Mailbox
	loading       bool
	selected      *fastmail.Email
	goBack        bool
	status        statusModel
	pendingDelete bool
	action        *emailAction
	searchInput   textinput.Model
	searchMode    bool   // actively typing in search input
	searchActive  bool   // showing search results
	searchQuery   string // current search query
	savedItems    []list.Item
	search        *string // signals parent to run search
	compose       bool    // signals parent to open compose view
}

func newEmailListModel(mailbox fastmail.Mailbox) emailListModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).
		BorderForeground(lipgloss.Color("12"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("8")).
		BorderForeground(lipgloss.Color("12"))

	l := list.New(nil, delegate, 0, 0)
	l.Title = mailbox.Name
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	ti := textinput.New()
	ti.Placeholder = "Search emails..."
	ti.CharLimit = 256

	return emailListModel{
		list:        l,
		mailbox:     mailbox,
		loading:     true,
		searchInput: ti,
	}
}

func (m *emailListModel) setSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m emailListModel) update(msg tea.Msg) (emailListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case emailsLoadedMsg:
		items := emailsToItems(msg.emails)
		m.loading = false
		cmd := m.list.SetItems(items)
		return m, cmd

	case searchResultsMsg:
		items := emailsToItems(msg.emails)
		m.loading = false
		m.searchActive = true
		m.list.Title = fmt.Sprintf("Search: %s", m.searchQuery)
		cmd := m.list.SetItems(items)
		return m, cmd

	case statusClearMsg:
		m.status.update(msg)
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchInput(msg)
		}
		if !m.list.SettingFilter() {
			if handled, cmd := m.handleKey(msg.String()); handled {
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// handleKey processes keybindings for navigation and actions.
// Returns true if the key was handled.
func (m *emailListModel) handleKey(key string) (bool, tea.Cmd) {
	// Cancel pending delete on any key that isn't "x"
	if m.pendingDelete && key != "x" {
		m.pendingDelete = false
		cmd := m.status.setStatus("", false)
		m.status.visible = false
		return true, cmd
	}

	switch key {
	case keyEsc:
		if m.searchActive {
			m.searchActive = false
			m.searchQuery = ""
			m.list.Title = m.mailbox.Name
			cmd := m.list.SetItems(m.savedItems)
			m.savedItems = nil
			return true, cmd
		}
		m.goBack = true
		return true, nil
	case keyEnter:
		if item, ok := m.list.SelectedItem().(emailItem); ok {
			m.selected = &item.email
		}
		return true, nil
	case "/":
		m.searchMode = true
		m.searchInput.Focus()
		return true, nil
	case "c":
		m.compose = true
		return true, nil
	case "a", ".", "m", "f":
		return m.handleActionKey(key)
	case "x":
		return m.handleDeleteKey()
	}
	return false, nil
}

func (m *emailListModel) handleActionKey(key string) (bool, tea.Cmd) {
	item, ok := m.list.SelectedItem().(emailItem)
	if !ok {
		return true, nil
	}
	kindMap := map[string]string{"a": "archive", ".": "toggleRead", "m": "move", "f": "toggleFlag"}
	m.action = &emailAction{kind: kindMap[key], email: item.email}
	return true, nil
}

func (m *emailListModel) handleDeleteKey() (bool, tea.Cmd) {
	item, ok := m.list.SelectedItem().(emailItem)
	if !ok {
		return true, nil
	}
	if m.pendingDelete {
		m.pendingDelete = false
		m.action = &emailAction{kind: "delete", email: item.email}
		return true, nil
	}
	m.pendingDelete = true
	return true, m.status.setStatus("Press x again to delete", false)
}

func (m *emailListModel) handleSearchInput(msg tea.KeyMsg) (emailListModel, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		m.searchMode = false
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		return *m, nil
	case keyEnter:
		query := m.searchInput.Value()
		if query == "" {
			m.searchMode = false
			m.searchInput.Blur()
			return *m, nil
		}
		m.searchMode = false
		m.searchInput.Blur()
		m.searchQuery = query
		m.searchInput.SetValue("")
		m.loading = true
		if !m.searchActive {
			m.savedItems = m.list.Items()
		}
		m.search = &query
		return *m, nil
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return *m, cmd
}

func (m emailListModel) view() string {
	if m.loading {
		if m.searchQuery != "" {
			return fmt.Sprintf("Searching for %q...", m.searchQuery)
		}
		return fmt.Sprintf("Loading emails from %s...", m.mailbox.Name)
	}
	v := m.list.View()
	if m.searchMode {
		v += "\n" + m.searchInput.View()
	} else if s := m.status.view(); s != "" {
		v += "\n" + s
	}
	return v
}

// emailsLoadedMsg is sent when emails are fetched for a mailbox.
type emailsLoadedMsg struct {
	emails []fastmail.Email
}

// mailboxSelectedMsg signals that a mailbox was selected in the list.
type mailboxSelectedMsg struct {
	mailbox fastmail.Mailbox
}

func emailsToItems(emails []fastmail.Email) []list.Item {
	items := make([]list.Item, len(emails))
	for i, e := range emails {
		items[i] = emailItem{email: e}
	}
	return items
}

// formatAge returns a human-friendly relative time string.
func formatAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 02")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func (i emailItem) richTitle(theme Theme, width int) string {
	// Flag indicator
	var flagStr string
	flagStyle := lipgloss.NewStyle().Foreground(theme.Flagged)
	if i.email.IsFlagged() {
		flagStr = flagStyle.Render("★ ")
	} else {
		flagStr = "  "
	}

	// Attachment indicator
	var attachStr string
	if len(i.email.Attachments) > 0 {
		attachStr = "📎 "
	}

	// From
	from := i.email.From.Email
	if i.email.From.Name != "" {
		from = i.email.From.Name
	}
	from = truncate(from, 20)

	// Subject
	subject := i.email.Subject
	if subject == "" {
		subject = "(no subject)"
	}

	// Date
	age := formatAge(i.email.ReceivedAt)

	// Style based on read state
	textStyle := lipgloss.NewStyle().Foreground(theme.Unread).Bold(true)
	if i.email.IsRead() {
		textStyle = lipgloss.NewStyle().Foreground(theme.Read)
	}

	dateStyle := lipgloss.NewStyle().Foreground(theme.Read).Align(lipgloss.Right)

	fromRendered := textStyle.Width(22).Render(from)
	subjectRendered := textStyle.Render(truncate(subject, width-50))
	dateRendered := dateStyle.Render(age)

	return flagStr + attachStr + fromRendered + " " + subjectRendered + " " + dateRendered
}
