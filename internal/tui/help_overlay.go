package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).MarginBottom(1)
	helpKeyStyle   = lipgloss.NewStyle().Bold(true).Width(12)
	helpDescStyle  = lipgloss.NewStyle()
)

func helpForView(v view) string {
	var b strings.Builder

	switch v {
	case viewMailboxList:
		b.WriteString(helpTitleStyle.Render("Keybindings — Mailbox List"))
		b.WriteString("\n\n")
		writeBinding(&b, "enter", "Open mailbox")
		writeBinding(&b, "/", "Filter mailboxes")
		writeBinding(&b, "q", "Quit")
	case viewEmailList:
		b.WriteString(helpTitleStyle.Render("Keybindings — Email List"))
		b.WriteString("\n\n")
		writeBinding(&b, "enter", "Open email")
		writeBinding(&b, "a", "Archive")
		writeBinding(&b, "x", "Delete (press twice)")
		writeBinding(&b, "r", "Toggle read/unread")
		writeBinding(&b, "f", "Toggle flag")
		writeBinding(&b, "m", "Move to mailbox")
		writeBinding(&b, "c", "Compose")
		writeBinding(&b, "/", "Search emails")
		writeBinding(&b, "esc", "Back to mailboxes")
		writeBinding(&b, "q", "Quit")
	case viewEmailReader:
		b.WriteString(helpTitleStyle.Render("Keybindings — Email Reader"))
		b.WriteString("\n\n")
		writeBinding(&b, "j/k", "Scroll line")
		writeBinding(&b, "d/u", "Half page down/up")
		writeBinding(&b, "g/G", "Top/Bottom")
		writeBinding(&b, "a", "Archive")
		writeBinding(&b, "x", "Delete (press twice)")
		writeBinding(&b, "r", "Toggle read/unread")
		writeBinding(&b, "f", "Toggle flag")
		writeBinding(&b, "m", "Move to mailbox")
		writeBinding(&b, "A", "Attachments")
		writeBinding(&b, "t", "View thread")
		writeBinding(&b, "q/esc", "Back to email list")
	case viewThreadView:
		b.WriteString(helpTitleStyle.Render("Keybindings — Thread View"))
		b.WriteString("\n\n")
		writeBinding(&b, "enter", "View email")
		writeBinding(&b, "q/esc", "Back to reader")
	case viewAttachmentPicker:
		b.WriteString(helpTitleStyle.Render("Keybindings — Attachments"))
		b.WriteString("\n\n")
		writeBinding(&b, "enter", "Download attachment")
		writeBinding(&b, "q/esc", "Back to reader")
	case viewCompose:
		b.WriteString(helpTitleStyle.Render("Keybindings — Compose"))
		b.WriteString("\n\n")
		writeBinding(&b, "tab", "Next field")
		writeBinding(&b, "shift+tab", "Previous field")
		writeBinding(&b, "ctrl+s", "Send email")
		writeBinding(&b, "esc", "Cancel")
	case viewMovePicker:
		return ""
	}

	b.WriteString("\n")
	writeBinding(&b, "?", "Show this help")

	return b.String()
}

func writeBinding(b *strings.Builder, key, desc string) {
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render(key))
	b.WriteString(helpDescStyle.Render(desc))
	b.WriteString("\n")
}
