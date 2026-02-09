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
		writeBinding(&b, "/", "Filter emails")
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
		writeBinding(&b, "q/esc", "Back to email list")
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
