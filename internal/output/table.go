package output

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// FormatTable renders data as an ASCII table.
//
// Example output:
//
//	| ID   | Subject      |
//	|------|--------------|
//	| 123  | Hello World  |
//	| 456  | Test Email   |
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if w := utf8.RuneCountInString(cell); w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	var sb strings.Builder

	// Header row
	sb.WriteString("|")
	for i, h := range headers {
		sb.WriteString(" ")
		sb.WriteString(padRight(h, widths[i]))
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Separator row
	sb.WriteString("|")
	for _, w := range widths {
		sb.WriteString(strings.Repeat("-", w+2))
		sb.WriteString("|")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		sb.WriteString("|")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			sb.WriteString(" ")
			sb.WriteString(padRight(cell, widths[i]))
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// padRight pads a string with spaces on the right to reach the specified width.
func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// PrintEmail writes a single email in human-readable format.
func (w *Writer) PrintEmail(email *fastmail.Email) error {
	if w.IsQuiet() {
		return nil
	}

	if w.ShouldOutputJSON() {
		return w.WriteJSON(email)
	}

	// Human-readable format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:      %s\n", email.ID))
	sb.WriteString(fmt.Sprintf("From:    %s\n", email.From.String()))
	sb.WriteString(fmt.Sprintf("To:      %s\n", formatAddressList(email.To)))
	if len(email.Cc) > 0 {
		sb.WriteString(fmt.Sprintf("Cc:      %s\n", formatAddressList(email.Cc)))
	}
	sb.WriteString(fmt.Sprintf("Date:    %s\n", email.Date.Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", email.Subject))
	sb.WriteString(fmt.Sprintf("\n%s\n", email.Preview))

	_, err := w.WriteString(sb.String())
	return err
}

// PrintEmailList writes multiple emails in table format.
func (w *Writer) PrintEmailList(emails []fastmail.Email) error {
	if w.IsQuiet() {
		return nil
	}

	if w.ShouldOutputJSON() {
		return w.WriteJSON(emails)
	}

	if len(emails) == 0 {
		_, err := w.WriteString("No emails found.\n")
		return err
	}

	// Build table
	headers := []string{"ID", "From", "Subject", "Date"}
	rows := make([][]string, len(emails))
	for i, e := range emails {
		rows[i] = []string{
			truncate(e.ID, 12),
			truncate(e.From.String(), 30),
			truncate(e.Subject, 40),
			e.Date.Format("Jan 02 15:04"),
		}
	}

	table := FormatTable(headers, rows)
	_, err := w.WriteString(table)
	return err
}

// formatAddressList formats a slice of email addresses as a comma-separated string.
func formatAddressList(addrs []fastmail.EmailAddress) string {
	if len(addrs) == 0 {
		return ""
	}
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = a.String()
	}
	return strings.Join(parts, ", ")
}

// truncate shortens a string to maxLen, adding ellipsis if needed.
func truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
