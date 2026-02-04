package fastmail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExportJSONL(t *testing.T) {
	t.Run("empty slice produces no output", func(t *testing.T) {
		var buf bytes.Buffer
		err := ExportJSONL(&buf, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("expected empty output, got %q", buf.String())
		}
	})

	t.Run("single email produces single line", func(t *testing.T) {
		var buf bytes.Buffer
		emails := []Email{
			{
				ID:      "email-1",
				Subject: "Test Subject",
				From:    EmailAddress{Name: "Alice", Email: "alice@example.com"},
			},
		}

		err := ExportJSONL(&buf, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}

		// Verify it's valid JSON
		var decoded Email
		if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if decoded.ID != "email-1" {
			t.Errorf("expected ID email-1, got %s", decoded.ID)
		}
	})

	t.Run("multiple emails produce multiple lines", func(t *testing.T) {
		var buf bytes.Buffer
		emails := []Email{
			{ID: "email-1", Subject: "First"},
			{ID: "email-2", Subject: "Second"},
			{ID: "email-3", Subject: "Third"},
		}

		err := ExportJSONL(&buf, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}

		// Verify each line is valid JSON with correct ID
		for i, line := range lines {
			var decoded Email
			if err := json.Unmarshal([]byte(line), &decoded); err != nil {
				t.Errorf("line %d: invalid JSON: %v", i, err)
				continue
			}
			expectedID := emails[i].ID
			if decoded.ID != expectedID {
				t.Errorf("line %d: expected ID %s, got %s", i, expectedID, decoded.ID)
			}
		}
	})

	t.Run("preserves all email fields", func(t *testing.T) {
		var buf bytes.Buffer
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		email := Email{
			ID:         "email-full",
			ThreadID:   "thread-1",
			Subject:    "Full Email Test",
			From:       EmailAddress{Name: "Sender", Email: "sender@example.com"},
			To:         []EmailAddress{{Name: "Recipient", Email: "recipient@example.com"}},
			Cc:         []EmailAddress{{Email: "cc@example.com"}},
			Bcc:        []EmailAddress{{Email: "bcc@example.com"}},
			ReceivedAt: now,
			Preview:    "This is a preview...",
			Body:       "Full body content here.",
			Keywords:   []string{"$seen", "$flagged"},
			MailboxIDs: []string{"inbox", "archive"},
			Size:       1234,
		}

		err := ExportJSONL(&buf, []Email{email})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var decoded Email
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// Verify all fields preserved
		if decoded.ID != email.ID {
			t.Errorf("ID: expected %s, got %s", email.ID, decoded.ID)
		}
		if decoded.ThreadID != email.ThreadID {
			t.Errorf("ThreadID: expected %s, got %s", email.ThreadID, decoded.ThreadID)
		}
		if decoded.Subject != email.Subject {
			t.Errorf("Subject: expected %s, got %s", email.Subject, decoded.Subject)
		}
		if decoded.From.Email != email.From.Email {
			t.Errorf("From.Email: expected %s, got %s", email.From.Email, decoded.From.Email)
		}
		if len(decoded.To) != 1 || decoded.To[0].Email != email.To[0].Email {
			t.Errorf("To: expected %v, got %v", email.To, decoded.To)
		}
		if decoded.Body != email.Body {
			t.Errorf("Body: expected %s, got %s", email.Body, decoded.Body)
		}
		if decoded.Size != email.Size {
			t.Errorf("Size: expected %d, got %d", email.Size, decoded.Size)
		}
		if !decoded.ReceivedAt.Equal(email.ReceivedAt) {
			t.Errorf("ReceivedAt: expected %v, got %v", email.ReceivedAt, decoded.ReceivedAt)
		}
	})

	t.Run("each line ends with newline", func(t *testing.T) {
		var buf bytes.Buffer
		emails := []Email{
			{ID: "email-1"},
			{ID: "email-2"},
		}

		err := ExportJSONL(&buf, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.HasSuffix(output, "\n") {
			t.Errorf("output should end with newline")
		}

		// Count newlines - should match number of emails
		count := strings.Count(output, "\n")
		if count != 2 {
			t.Errorf("expected 2 newlines, got %d", count)
		}
	})
}

func TestExportMaildir(t *testing.T) {
	t.Run("creates maildir directory structure", func(t *testing.T) {
		dir := t.TempDir()

		err := ExportMaildir(dir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check standard Maildir subdirectories exist
		for _, subdir := range []string{"cur", "new", "tmp"} {
			path := filepath.Join(dir, subdir)
			info, err := os.Stat(path)
			if err != nil {
				t.Errorf("subdir %s not created: %v", subdir, err)
				continue
			}
			if !info.IsDir() {
				t.Errorf("%s should be a directory", subdir)
			}
		}
	})

	t.Run("writes email to cur directory", func(t *testing.T) {
		dir := t.TempDir()
		emails := []Email{
			{
				ID:         "email-1",
				Subject:    "Test Subject",
				From:       EmailAddress{Name: "Alice", Email: "alice@example.com"},
				To:         []EmailAddress{{Name: "Bob", Email: "bob@example.com"}},
				ReceivedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Body:       "Hello, World!",
			},
		}

		err := ExportMaildir(dir, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that a file was created in cur/
		entries, err := os.ReadDir(filepath.Join(dir, "cur"))
		if err != nil {
			t.Fatalf("failed to read cur directory: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 file in cur/, got %d", len(entries))
		}
	})

	t.Run("filename follows maildir convention", func(t *testing.T) {
		dir := t.TempDir()
		emails := []Email{
			{
				ID:         "email-1",
				ReceivedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Keywords:   []string{"$seen"},
			},
		}

		err := ExportMaildir(dir, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		entries, err := os.ReadDir(filepath.Join(dir, "cur"))
		if err != nil {
			t.Fatalf("failed to read cur directory: %v", err)
		}

		filename := entries[0].Name()
		// Maildir filename format: timestamp.unique.hostname:info
		// The :info part contains flags, :2,S means Seen
		if !strings.Contains(filename, ":") {
			t.Errorf("filename should contain ':' separator: %s", filename)
		}
		if !strings.Contains(filename, ".") {
			t.Errorf("filename should contain '.' separators: %s", filename)
		}
		// Seen emails should have S flag
		if !strings.HasSuffix(filename, "S") && !strings.Contains(filename, "S") {
			t.Errorf("seen email should have S flag in filename: %s", filename)
		}
	})

	t.Run("message content is valid RFC 5322", func(t *testing.T) {
		dir := t.TempDir()
		emails := []Email{
			{
				ID:         "email-1",
				Subject:    "Test Subject",
				From:       EmailAddress{Name: "Alice", Email: "alice@example.com"},
				To:         []EmailAddress{{Name: "Bob", Email: "bob@example.com"}},
				ReceivedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Body:       "Hello, World!",
			},
		}

		err := ExportMaildir(dir, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		entries, err := os.ReadDir(filepath.Join(dir, "cur"))
		if err != nil {
			t.Fatalf("failed to read cur directory: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "cur", entries[0].Name()))
		if err != nil {
			t.Fatalf("failed to read message file: %v", err)
		}

		msg := string(content)

		// RFC 5322 requires headers before body, separated by blank line
		if !strings.Contains(msg, "\r\n\r\n") && !strings.Contains(msg, "\n\n") {
			t.Error("message should have blank line between headers and body")
		}

		// Check required headers
		if !strings.Contains(msg, "From:") {
			t.Error("message should have From header")
		}
		if !strings.Contains(msg, "To:") {
			t.Error("message should have To header")
		}
		if !strings.Contains(msg, "Subject:") {
			t.Error("message should have Subject header")
		}
		if !strings.Contains(msg, "Date:") {
			t.Error("message should have Date header")
		}

		// Check body content
		if !strings.Contains(msg, "Hello, World!") {
			t.Error("message should contain body content")
		}
	})

	t.Run("multiple emails create multiple files", func(t *testing.T) {
		dir := t.TempDir()
		emails := []Email{
			{ID: "email-1", Subject: "First", ReceivedAt: time.Now()},
			{ID: "email-2", Subject: "Second", ReceivedAt: time.Now()},
			{ID: "email-3", Subject: "Third", ReceivedAt: time.Now()},
		}

		err := ExportMaildir(dir, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		entries, err := os.ReadDir(filepath.Join(dir, "cur"))
		if err != nil {
			t.Fatalf("failed to read cur directory: %v", err)
		}
		if len(entries) != 3 {
			t.Errorf("expected 3 files in cur/, got %d", len(entries))
		}
	})

	t.Run("flagged email has F flag in filename", func(t *testing.T) {
		dir := t.TempDir()
		emails := []Email{
			{
				ID:         "email-1",
				ReceivedAt: time.Now(),
				Keywords:   []string{"$flagged"},
			},
		}

		err := ExportMaildir(dir, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		entries, err := os.ReadDir(filepath.Join(dir, "cur"))
		if err != nil {
			t.Fatalf("failed to read cur directory: %v", err)
		}

		filename := entries[0].Name()
		if !strings.Contains(filename, "F") {
			t.Errorf("flagged email should have F flag in filename: %s", filename)
		}
	})

	t.Run("handles special characters in subject", func(t *testing.T) {
		dir := t.TempDir()
		emails := []Email{
			{
				ID:         "email-1",
				Subject:    "Test: Special chars & symbols <here>",
				From:       EmailAddress{Email: "test@example.com"},
				ReceivedAt: time.Now(),
				Body:       "Body content",
			},
		}

		err := ExportMaildir(dir, emails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		entries, err := os.ReadDir(filepath.Join(dir, "cur"))
		if err != nil {
			t.Fatalf("failed to read cur directory: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "cur", entries[0].Name()))
		if err != nil {
			t.Fatalf("failed to read message file: %v", err)
		}

		if !strings.Contains(string(content), "Subject:") {
			t.Error("message should have Subject header")
		}
	})
}
