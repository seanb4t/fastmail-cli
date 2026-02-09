package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

const emailPageSize = 50

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

// emailListModel manages the email list view for a selected mailbox.
type emailListModel struct {
	list     list.Model
	mailbox  fastmail.Mailbox
	loading  bool
	selected *fastmail.Email
	goBack   bool
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

	return emailListModel{
		list:    l,
		mailbox: mailbox,
		loading: true,
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

	case tea.KeyMsg:
		if !m.list.SettingFilter() {
			switch msg.String() {
			case "esc":
				m.goBack = true
				return m, nil
			case "enter":
				if item, ok := m.list.SelectedItem().(emailItem); ok {
					m.selected = &item.email
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m emailListModel) view() string {
	if m.loading {
		return fmt.Sprintf("\n  Loading emails from %s...", m.mailbox.Name)
	}
	return m.list.View()
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
