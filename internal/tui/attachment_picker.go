package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// attachmentDownloadedMsg signals a successful attachment download.
type attachmentDownloadedMsg struct {
	name string
	path string
}

// attachmentDownloadErrMsg signals a failed attachment download.
type attachmentDownloadErrMsg struct {
	err  error
	name string
}

// attachmentItem wraps an Attachment for the bubbles list.
type attachmentItem struct {
	attachment fastmail.Attachment
}

func (i attachmentItem) Title() string {
	return i.attachment.Name
}

func (i attachmentItem) Description() string {
	return fmt.Sprintf("%s  %s", i.attachment.Type, formatSize(i.attachment.Size))
}

func (i attachmentItem) FilterValue() string {
	return i.attachment.Name
}

// attachmentPickerModel lets the user pick an attachment to download.
type attachmentPickerModel struct {
	list     list.Model
	emailID  string
	canceled bool
	selected *fastmail.Attachment
}

func newAttachmentPickerModel(emailID string, attachments []fastmail.Attachment) attachmentPickerModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).
		BorderForeground(lipgloss.Color("12"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("8")).
		BorderForeground(lipgloss.Color("12"))

	items := make([]list.Item, len(attachments))
	for i, a := range attachments {
		items[i] = attachmentItem{attachment: a}
	}

	l := list.New(items, delegate, 0, 0)
	l.Title = "Attachments"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	return attachmentPickerModel{
		list:    l,
		emailID: emailID,
	}
}

func (m *attachmentPickerModel) setSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m attachmentPickerModel) update(msg tea.Msg) (attachmentPickerModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEsc, "q":
			m.canceled = true
			return m, nil
		case keyEnter:
			if item, ok := m.list.SelectedItem().(attachmentItem); ok {
				att := item.attachment
				m.selected = &att
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m attachmentPickerModel) view() string {
	return m.list.View()
}

// formatSize returns a human-friendly file size string.
func formatSize(bytes uint64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
