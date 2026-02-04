package fastmail

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ExportJSONL writes emails in JSON Lines format to the given writer.
// Each email is written as a single JSON object on its own line.
// This format is streaming-friendly and efficient for large exports.
func ExportJSONL(w io.Writer, emails []Email) error {
	enc := json.NewEncoder(w)
	for _, email := range emails {
		if err := enc.Encode(email); err != nil {
			return err
		}
	}
	return nil
}

// ExportMaildir exports emails to a Maildir directory structure.
// It creates the standard cur/, new/, and tmp/ subdirectories and writes
// each email as an RFC 5322 message file with proper Maildir filename conventions.
func ExportMaildir(dir string, emails []Email) error {
	// Create Maildir directory structure
	for _, subdir := range []string{"cur", "new", "tmp"} {
		if err := os.MkdirAll(filepath.Join(dir, subdir), 0700); err != nil {
			return fmt.Errorf("creating %s directory: %w", subdir, err)
		}
	}

	// Write each email to cur/ directory
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	for i, email := range emails {
		filename := maildirFilename(email, i, hostname)
		path := filepath.Join(dir, "cur", filename)

		content := formatRFC5322(email)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return fmt.Errorf("writing email %s: %w", email.ID, err)
		}
	}

	return nil
}

// maildirFilename generates a Maildir-compliant filename.
// Format: timestamp.unique.hostname:2,flags.
func maildirFilename(email Email, seq int, hostname string) string {
	timestamp := email.ReceivedAt.Unix()
	if timestamp <= 0 {
		timestamp = 1
	}

	unique := fmt.Sprintf("P%dQ%d", os.Getpid(), seq)
	flags := maildirFlags(email)

	return fmt.Sprintf("%d.%s.%s:2,%s", timestamp, unique, hostname, flags)
}

// maildirFlags returns the Maildir flags string based on email keywords.
// Flags are alphabetically sorted per Maildir spec.
func maildirFlags(email Email) string {
	var flags []string

	if email.HasKeyword(KeywordDraft) {
		flags = append(flags, "D")
	}
	if email.HasKeyword(KeywordFlagged) {
		flags = append(flags, "F")
	}
	if email.HasKeyword(KeywordAnswered) {
		flags = append(flags, "R")
	}
	if email.HasKeyword(KeywordSeen) {
		flags = append(flags, "S")
	}

	sort.Strings(flags)
	return strings.Join(flags, "")
}

// ExportMbox writes emails in standard Mboxo format to the given writer.
// Each email is preceded by a "From " envelope line and any lines in the
// message body starting with "From " are quoted with ">".
// This format is streaming-friendly and compatible with standard Unix mail tools.
func ExportMbox(w io.Writer, emails []Email) error {
	for _, email := range emails {
		// Write "From " envelope line (not a header, but mbox separator)
		// Format: "From sender@domain Day Mon DD HH:MM:SS YYYY"
		sender := email.From.Email
		if sender == "" {
			sender = "MAILER-DAEMON"
		}
		timestamp := email.ReceivedAt.Format(time.ANSIC)
		if _, err := fmt.Fprintf(w, "From %s %s\n", sender, timestamp); err != nil {
			return err
		}

		// Write RFC 5322 content with From-quoting for mbox safety
		content := formatRFC5322(email)
		quoted := mboxQuoteFrom(content)
		if _, err := io.WriteString(w, quoted); err != nil {
			return err
		}

		// Ensure message ends with blank line (mbox message separator)
		if !strings.HasSuffix(quoted, "\n") {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

// mboxQuoteFrom quotes lines starting with "From " by prefixing with ">".
// This is the Mboxo quoting convention to prevent false message boundaries.
func mboxQuoteFrom(content string) string {
	// Use LF line endings for mbox (Unix convention)
	content = strings.ReplaceAll(content, "\r\n", "\n")

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "From ") {
			lines[i] = ">" + line
		}
	}
	return strings.Join(lines, "\n")
}

// formatRFC5322 converts an Email to RFC 5322 message format.
func formatRFC5322(email Email) string {
	var b strings.Builder

	// Headers
	if email.From.Email != "" {
		b.WriteString(fmt.Sprintf("From: %s\r\n", email.From.String()))
	}

	if len(email.To) > 0 {
		addrs := make([]string, len(email.To))
		for i, addr := range email.To {
			addrs[i] = addr.String()
		}
		b.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(addrs, ", ")))
	}

	if len(email.Cc) > 0 {
		addrs := make([]string, len(email.Cc))
		for i, addr := range email.Cc {
			addrs[i] = addr.String()
		}
		b.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(addrs, ", ")))
	}

	b.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	b.WriteString(fmt.Sprintf("Date: %s\r\n", email.ReceivedAt.Format("Mon, 02 Jan 2006 15:04:05 -0700")))
	b.WriteString(fmt.Sprintf("Message-ID: <%s@exported>\r\n", email.ID))

	// Blank line separates headers from body
	b.WriteString("\r\n")

	// Body
	b.WriteString(email.Body)
	if !strings.HasSuffix(email.Body, "\n") {
		b.WriteString("\r\n")
	}

	return b.String()
}
