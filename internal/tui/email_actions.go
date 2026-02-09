package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// emailAction signals the parent model to execute an email action.
type emailAction struct {
	kind  string // "archive", "delete", "toggleRead", "toggleFlag", "move"
	email fastmail.Email
}

// emailActionDoneMsg signals a successful email action.
type emailActionDoneMsg struct {
	action  string // display text: "Archived", "Deleted", etc.
	emailID string
}

// emailActionErrMsg signals a failed email action.
type emailActionErrMsg struct {
	err    error
	action string
}

func archiveEmailCmd(client *fastmail.Client, emailID string) tea.Cmd {
	return func() tea.Msg {
		if err := client.Mail().Move(context.Background(), emailID, "Archive"); err != nil {
			return emailActionErrMsg{err: err, action: "Archive"}
		}
		return emailActionDoneMsg{action: "Archived", emailID: emailID}
	}
}

func deleteEmailCmd(client *fastmail.Client, emailID string) tea.Cmd {
	return func() tea.Msg {
		if err := client.Mail().Delete(context.Background(), emailID); err != nil {
			return emailActionErrMsg{err: err, action: "Delete"}
		}
		return emailActionDoneMsg{action: "Deleted", emailID: emailID}
	}
}

func moveEmailCmd(client *fastmail.Client, emailID, mailboxID string) tea.Cmd {
	return func() tea.Msg {
		if err := client.Mail().Move(context.Background(), emailID, mailboxID); err != nil {
			return emailActionErrMsg{err: err, action: "Move"}
		}
		return emailActionDoneMsg{action: "Moved", emailID: emailID}
	}
}

func toggleReadCmd(client *fastmail.Client, emailID string, isRead bool) tea.Cmd {
	return func() tea.Msg {
		actions := []fastmail.KeywordAction{
			{Keyword: fastmail.KeywordSeen, Set: !isRead},
		}
		if err := client.Mail().SetKeywords(context.Background(), emailID, actions); err != nil {
			return emailActionErrMsg{err: err, action: "Toggle read"}
		}
		label := "Marked read"
		if isRead {
			label = "Marked unread"
		}
		return emailActionDoneMsg{action: label, emailID: emailID}
	}
}

func toggleFlagCmd(client *fastmail.Client, emailID string, isFlagged bool) tea.Cmd {
	return func() tea.Msg {
		actions := []fastmail.KeywordAction{
			{Keyword: fastmail.KeywordFlagged, Set: !isFlagged},
		}
		if err := client.Mail().SetKeywords(context.Background(), emailID, actions); err != nil {
			return emailActionErrMsg{err: err, action: "Toggle flag"}
		}
		label := "Flagged"
		if isFlagged {
			label = "Unflagged"
		}
		return emailActionDoneMsg{action: label, emailID: emailID}
	}
}
