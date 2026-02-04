// Package output provides TTY-aware formatting for CLI responses.
//
// Behavior:
//   - TTY stdout → human-readable tables
//   - Pipe stdout → JSON
//   - --json flag → forced JSON
//   - --quiet flag → no output
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// Writer provides output formatting with configurable mode.
type Writer struct {
	out       io.Writer
	err       io.Writer
	forceJSON bool
	quiet     bool
}

// Option configures a Writer.
type Option func(*Writer)

// WithJSON forces JSON output regardless of TTY detection.
func WithJSON(force bool) Option {
	return func(w *Writer) {
		w.forceJSON = force
	}
}

// WithQuiet suppresses all output.
func WithQuiet(quiet bool) Option {
	return func(w *Writer) {
		w.quiet = quiet
	}
}

// WithOutput sets the output writers.
func WithOutput(out, errOut io.Writer) Option {
	return func(w *Writer) {
		w.out = out
		w.err = errOut
	}
}

// New creates a Writer with the given options.
func New(opts ...Option) *Writer {
	w := &Writer{
		out: os.Stdout,
		err: os.Stderr,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// IsTerminal returns true if stdout is connected to a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// FormatJSON returns v as indented JSON.
func FormatJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding JSON: %w", err)
	}
	return string(data), nil
}

// shouldUseJSON returns true if output should be JSON.
func (w *Writer) shouldUseJSON() bool {
	if w.forceJSON {
		return true
	}
	return !IsTerminal()
}

// PrintEmail outputs a single email in the appropriate format.
func (w *Writer) PrintEmail(email *jmap.Email) {
	if w.quiet {
		return
	}

	if w.shouldUseJSON() {
		if out, err := FormatJSON(email); err == nil {
			fmt.Fprintln(w.out, out)
		}
		return
	}

	// Human-readable format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:       %s\n", email.ID))
	if email.Subject != "" {
		sb.WriteString(fmt.Sprintf("Subject:  %s\n", email.Subject))
	}
	if email.ReceivedAt != "" {
		sb.WriteString(fmt.Sprintf("Received: %s\n", email.ReceivedAt))
	}
	if email.Preview != "" {
		sb.WriteString(fmt.Sprintf("Preview:  %s\n", truncate(email.Preview, 80)))
	}
	fmt.Fprint(w.out, sb.String())
}

// PrintEmailList outputs a list of emails in the appropriate format.
func (w *Writer) PrintEmailList(emails []jmap.Email) {
	if w.quiet {
		return
	}

	if w.shouldUseJSON() {
		if out, err := FormatJSON(emails); err == nil {
			fmt.Fprintln(w.out, out)
		}
		return
	}

	// Table format
	headers := []string{"ID", "Subject", "Received"}
	rows := make([][]string, len(emails))
	for i, e := range emails {
		subject := truncate(e.Subject, 40)
		received := ""
		if len(e.ReceivedAt) >= 10 {
			received = e.ReceivedAt[:10] // YYYY-MM-DD
		}
		rows[i] = []string{e.ID, subject, received}
	}
	fmt.Fprint(w.out, FormatTable(headers, rows))
}

// Print outputs any value in the appropriate format.
func (w *Writer) Print(v any) {
	if w.quiet {
		return
	}

	if w.shouldUseJSON() {
		if out, err := FormatJSON(v); err == nil {
			fmt.Fprintln(w.out, out)
		}
		return
	}

	// Default: just print the value
	fmt.Fprintf(w.out, "%v\n", v)
}

// Error writes an error message to stderr.
func (w *Writer) Error(format string, args ...any) {
	fmt.Fprintf(w.err, "error: "+format+"\n", args...)
}

// truncate shortens s to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
