// Package output provides TTY-aware output formatting for the CLI.
package output

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
	"golang.org/x/term"
)

// Options controls output formatting behavior.
type Options struct {
	// JSON forces JSON output regardless of TTY.
	JSON bool
	// Quiet suppresses all output.
	Quiet bool
}

// Printer writes formatted output.
type Printer struct {
	w    io.Writer
	opts Options
}

// NewPrinter creates a Printer that writes to w with the given options.
func NewPrinter(w io.Writer, opts Options) *Printer {
	return &Printer{w: w, opts: opts}
}

// IsTerminal reports whether f is a terminal.
func IsTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// FormatJSON returns v as indented JSON.
func FormatJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// PrintEmail writes a single email to the output.
func (p *Printer) PrintEmail(email fastmail.Email) error {
	if p.opts.Quiet {
		return nil
	}

	if p.opts.JSON {
		s, err := FormatJSON(email)
		if err != nil {
			return err
		}
		_, err = io.WriteString(p.w, s+"\n")
		return err
	}

	// Table format for single email
	headers := []string{"ID", "Subject", "From", "Date"}
	rows := [][]string{
		{email.ID, email.Subject, email.From.String(), formatDate(email.Date)},
	}
	_, err := io.WriteString(p.w, FormatTable(headers, rows))
	return err
}

// PrintEmailList writes a list of emails to the output.
func (p *Printer) PrintEmailList(emails []fastmail.Email) error {
	if p.opts.Quiet {
		return nil
	}

	if p.opts.JSON {
		s, err := FormatJSON(emails)
		if err != nil {
			return err
		}
		_, err = io.WriteString(p.w, s+"\n")
		return err
	}

	// Table format
	headers := []string{"ID", "Subject", "From", "Date"}
	rows := make([][]string, len(emails))
	for i, e := range emails {
		rows[i] = []string{e.ID, e.Subject, e.From.String(), formatDate(e.Date)}
	}
	_, err := io.WriteString(p.w, FormatTable(headers, rows))
	return err
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
