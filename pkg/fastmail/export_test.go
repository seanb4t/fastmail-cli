package fastmail

import (
	"bytes"
	"encoding/json"
	"io"
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

	t.Run("returns error on write failure", func(t *testing.T) {
		w := &failWriter{}
		emails := []Email{{ID: "email-1"}}

		err := ExportJSONL(w, emails)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// failWriter always returns an error on write.
type failWriter struct{}

func (f *failWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}
