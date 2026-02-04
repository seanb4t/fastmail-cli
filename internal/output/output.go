// Package output provides formatting utilities for CLI output.
//
// Output automatically adapts based on context:
//   - TTY stdout → human-readable tables
//   - Piped stdout → JSON
//   - --json flag → forced JSON
//   - --quiet flag → no output
package output

import (
	"encoding/json"
	"io"
	"os"

	"golang.org/x/term"
)

// Mode determines how output is formatted.
type Mode int

const (
	// ModeAuto detects TTY and formats accordingly.
	ModeAuto Mode = iota
	// ModeJSON forces JSON output.
	ModeJSON
	// ModeQuiet suppresses all output.
	ModeQuiet
)

// Writer handles formatted output to a destination.
type Writer struct {
	out  io.Writer
	mode Mode
}

// NewWriter creates a Writer with the given destination and mode.
func NewWriter(out io.Writer, mode Mode) *Writer {
	return &Writer{out: out, mode: mode}
}

// IsTerminal reports whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

// IsStdoutTerminal reports whether stdout is a terminal.
func IsStdoutTerminal() bool {
	return IsTerminal(int(os.Stdout.Fd()))
}

// FormatJSON converts any value to indented JSON.
func FormatJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ShouldOutputJSON determines if JSON output should be used based on mode and TTY.
func (w *Writer) ShouldOutputJSON() bool {
	if w.mode == ModeJSON {
		return true
	}
	if w.mode == ModeQuiet {
		return false
	}
	// ModeAuto: use JSON when not a TTY
	if f, ok := w.out.(*os.File); ok {
		return !IsTerminal(int(f.Fd()))
	}
	// Non-file writers (buffers, etc.) default to table
	return false
}

// IsQuiet returns true if output should be suppressed.
func (w *Writer) IsQuiet() bool {
	return w.mode == ModeQuiet
}

// Write writes formatted output. Returns immediately if quiet mode.
func (w *Writer) Write(data []byte) (int, error) {
	if w.IsQuiet() {
		return len(data), nil
	}
	return w.out.Write(data)
}

// WriteString writes a string to output. Returns immediately if quiet mode.
func (w *Writer) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// WriteJSON writes value as JSON. Returns immediately if quiet mode.
func (w *Writer) WriteJSON(v any) error {
	if w.IsQuiet() {
		return nil
	}
	data, err := FormatJSON(v)
	if err != nil {
		return err
	}
	_, err = w.WriteString(data + "\n")
	return err
}
