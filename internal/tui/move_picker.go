package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// movePickerModel lets the user pick a destination mailbox for moving an email.
type movePickerModel struct {
	list     list.Model
	emailID  string
	selected *fastmail.Mailbox
	canceled bool
}

func newMovePickerModel(emailID string, mailboxes []fastmail.Mailbox) movePickerModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).
		BorderForeground(lipgloss.Color("12"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("8")).
		BorderForeground(lipgloss.Color("12"))

	items := mailboxesToItems(mailboxes)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Move to..."
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	return movePickerModel{
		list:    l,
		emailID: emailID,
	}
}

func (m *movePickerModel) setSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m movePickerModel) update(msg tea.Msg) (movePickerModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && !m.list.SettingFilter() {
		switch keyMsg.String() {
		case keyEsc:
			m.canceled = true
			return m, nil
		case keyEnter:
			if item, ok := m.list.SelectedItem().(mailboxItem); ok {
				m.selected = &item.mailbox
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m movePickerModel) view() string {
	return m.list.View()
}
