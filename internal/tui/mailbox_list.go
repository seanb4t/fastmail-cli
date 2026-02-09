package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// mailboxItem implements list.DefaultItem for the bubbles list component.
type mailboxItem struct {
	mailbox fastmail.Mailbox
}

func (i mailboxItem) Title() string {
	name := i.mailbox.Name
	if i.mailbox.UnreadEmails > 0 {
		return fmt.Sprintf("%s (%d)", name, i.mailbox.UnreadEmails)
	}
	return name
}

func (i mailboxItem) Description() string {
	return fmt.Sprintf("%d emails", i.mailbox.TotalEmails)
}

func (i mailboxItem) FilterValue() string {
	return i.mailbox.Name
}

// mailboxListModel manages the mailbox list view.
type mailboxListModel struct {
	list     list.Model
	loading  bool
	selected *fastmail.Mailbox
}

func newMailboxListModel() mailboxListModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).
		BorderForeground(lipgloss.Color("12"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("8")).
		BorderForeground(lipgloss.Color("12"))

	l := list.New(nil, delegate, 0, 0)
	l.Title = "Mailboxes"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	return mailboxListModel{
		list:    l,
		loading: true,
	}
}

func (m *mailboxListModel) setSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m mailboxListModel) update(msg tea.Msg) (mailboxListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case mailboxesLoadedMsg:
		items := mailboxesToItems(msg.mailboxes)
		m.loading = false
		cmd := m.list.SetItems(items)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "enter" {
			if item, ok := m.list.SelectedItem().(mailboxItem); ok {
				m.selected = &item.mailbox
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m mailboxListModel) view() string {
	if m.loading {
		return "\n  Loading mailboxes..."
	}
	return m.list.View()
}

// mailboxesLoadedMsg is sent when mailboxes are fetched.
type mailboxesLoadedMsg struct {
	mailboxes []fastmail.Mailbox
}

// mailboxesToItems converts mailboxes to list items, sorted with
// standard roles first (inbox, sent, drafts, etc.) then custom folders.
func mailboxesToItems(mailboxes []fastmail.Mailbox) []list.Item {
	sorted := make([]fastmail.Mailbox, len(mailboxes))
	copy(sorted, mailboxes)

	rolePriority := map[fastmail.MailboxRole]int{
		fastmail.RoleInbox:   0,
		fastmail.RoleDrafts:  1,
		fastmail.RoleSent:    2,
		fastmail.RoleArchive: 3,
		fastmail.RoleTrash:   4,
		fastmail.RoleJunk:    5,
	}

	sort.Slice(sorted, func(i, j int) bool {
		pi, oki := rolePriority[sorted[i].Role]
		pj, okj := rolePriority[sorted[j].Role]
		if oki && okj {
			return pi < pj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return sorted[i].Name < sorted[j].Name
	})

	items := make([]list.Item, len(sorted))
	for idx, mb := range sorted {
		items[idx] = mailboxItem{mailbox: mb}
	}
	return items
}
