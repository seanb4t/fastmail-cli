package fastmail

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
// Format: timestamp.unique.hostname:2,flags
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
