package output

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestIsTerminal(t *testing.T) {
	// When running tests, stdout is typically not a TTY
	// This is a basic sanity check that the function doesn't panic
	isTTY := IsTerminal(int(os.Stdout.Fd()))
	isStdoutTTY := IsStdoutTerminal()

	// In test environment, these should be consistent
	if isTTY != isStdoutTTY {
		t.Errorf("IsTerminal(stdout.Fd()) = %v, IsStdoutTerminal() = %v, expected same", isTTY, isStdoutTTY)
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "simple struct",
			input: struct{ Name string }{Name: "test"},
			want:  "{\n  \"Name\": \"test\"\n}",
		},
		{
			name:  "slice",
			input: []string{"a", "b"},
			want:  "[\n  \"a\",\n  \"b\"\n]",
		},
		{
			name:  "map",
			input: map[string]int{"x": 1},
			want:  "{\n  \"x\": 1\n}",
		},
		{
			name:    "channel (unmarshalable)",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FormatJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTable(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		want    string
	}{
		{
			name:    "empty headers",
			headers: []string{},
			rows:    [][]string{},
			want:    "",
		},
		{
			name:    "single column",
			headers: []string{"Name"},
			rows:    [][]string{{"Alice"}, {"Bob"}},
			want: "| Name  |\n" +
				"|-------|\n" +
				"| Alice |\n" +
				"| Bob   |\n",
		},
		{
			name:    "multiple columns",
			headers: []string{"ID", "Name"},
			rows:    [][]string{{"1", "Alice"}, {"2", "Bob"}},
			want: "| ID | Name  |\n" +
				"|----|-------|\n" +
				"| 1  | Alice |\n" +
				"| 2  | Bob   |\n",
		},
		{
			name:    "unicode content",
			headers: []string{"Name"},
			rows:    [][]string{{"café"}},
			// café has 4 runes, same as Name, so width = 4
			want: "| Name |\n" +
				"|------|\n" +
				"| café |\n",
		},
		{
			name:    "row shorter than headers",
			headers: []string{"A", "B", "C"},
			rows:    [][]string{{"1"}},
			want: "| A | B | C |\n" +
				"|---|---|---|\n" +
				"| 1 |   |   |\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTable(tt.headers, tt.rows)
			if got != tt.want {
				t.Errorf("FormatTable() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestWriter_Modes(t *testing.T) {
	t.Run("quiet mode suppresses output", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeQuiet)

		n, err := w.WriteString("hello")
		if err != nil {
			t.Errorf("WriteString() error = %v", err)
		}
		if n != 5 {
			t.Errorf("WriteString() = %d, want 5", n)
		}
		if buf.Len() != 0 {
			t.Errorf("buffer should be empty in quiet mode, got %q", buf.String())
		}
	})

	t.Run("normal mode writes output", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeAuto)

		_, err := w.WriteString("hello")
		if err != nil {
			t.Errorf("WriteString() error = %v", err)
		}
		if buf.String() != "hello" {
			t.Errorf("buffer = %q, want %q", buf.String(), "hello")
		}
	})

	t.Run("json mode forces JSON detection", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeJSON)

		if !w.ShouldOutputJSON() {
			t.Error("ShouldOutputJSON() should return true in ModeJSON")
		}
	})

	t.Run("auto mode with buffer uses table", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeAuto)

		// Buffers are not TTYs and not files, so default to table
		if w.ShouldOutputJSON() {
			t.Error("ShouldOutputJSON() should return false for non-file buffer")
		}
	})
}

func TestWriter_WriteJSON(t *testing.T) {
	t.Run("writes JSON with newline", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeAuto)

		err := w.WriteJSON(map[string]string{"key": "value"})
		if err != nil {
			t.Errorf("WriteJSON() error = %v", err)
		}

		want := "{\n  \"key\": \"value\"\n}\n"
		if buf.String() != want {
			t.Errorf("WriteJSON() wrote %q, want %q", buf.String(), want)
		}
	})

	t.Run("quiet mode suppresses JSON", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeQuiet)

		err := w.WriteJSON(map[string]string{"key": "value"})
		if err != nil {
			t.Errorf("WriteJSON() error = %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("buffer should be empty in quiet mode")
		}
	})
}

func TestWriter_PrintEmail(t *testing.T) {
	email := &fastmail.Email{
		ID:       "msg123",
		Subject:  "Test Subject",
		From:     fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
		To:       []fastmail.EmailAddress{{Email: "bob@example.com"}},
		Cc:       []fastmail.EmailAddress{{Name: "Charlie", Email: "charlie@example.com"}},
		Date:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Preview:  "This is the preview text.",
		Keywords: []string{"$seen"},
	}

	t.Run("human readable format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeAuto)

		err := w.PrintEmail(email)
		if err != nil {
			t.Errorf("PrintEmail() error = %v", err)
		}

		output := buf.String()
		checks := []string{
			"ID:      msg123",
			"From:    Alice <alice@example.com>",
			"To:      bob@example.com",
			"Cc:      Charlie <charlie@example.com>",
			"Subject: Test Subject",
			"This is the preview text.",
		}
		for _, check := range checks {
			if !bytes.Contains(buf.Bytes(), []byte(check)) {
				t.Errorf("PrintEmail() output missing %q\nGot:\n%s", check, output)
			}
		}
	})

	t.Run("JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeJSON)

		err := w.PrintEmail(email)
		if err != nil {
			t.Errorf("PrintEmail() error = %v", err)
		}

		output := buf.String()
		if !bytes.Contains(buf.Bytes(), []byte(`"ID": "msg123"`)) {
			t.Errorf("PrintEmail() JSON missing ID field\nGot:\n%s", output)
		}
	})

	t.Run("quiet mode", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeQuiet)

		err := w.PrintEmail(email)
		if err != nil {
			t.Errorf("PrintEmail() error = %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("buffer should be empty in quiet mode")
		}
	})
}

func TestWriter_PrintEmailList(t *testing.T) {
	emails := []fastmail.Email{
		{
			ID:      "msg1",
			Subject: "First Email",
			From:    fastmail.EmailAddress{Email: "alice@example.com"},
			Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			ID:      "msg2",
			Subject: "Second Email",
			From:    fastmail.EmailAddress{Name: "Bob", Email: "bob@example.com"},
			Date:    time.Date(2024, 1, 16, 14, 0, 0, 0, time.UTC),
		},
	}

	t.Run("table format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeAuto)

		err := w.PrintEmailList(emails)
		if err != nil {
			t.Errorf("PrintEmailList() error = %v", err)
		}

		output := buf.String()
		checks := []string{"ID", "From", "Subject", "Date", "msg1", "msg2", "First Email", "Second Email"}
		for _, check := range checks {
			if !bytes.Contains(buf.Bytes(), []byte(check)) {
				t.Errorf("PrintEmailList() output missing %q\nGot:\n%s", check, output)
			}
		}
	})

	t.Run("JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeJSON)

		err := w.PrintEmailList(emails)
		if err != nil {
			t.Errorf("PrintEmailList() error = %v", err)
		}

		output := buf.String()
		if !bytes.Contains(buf.Bytes(), []byte(`"ID": "msg1"`)) {
			t.Errorf("PrintEmailList() JSON missing first email\nGot:\n%s", output)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf, ModeAuto)

		err := w.PrintEmailList([]fastmail.Email{})
		if err != nil {
			t.Errorf("PrintEmailList() error = %v", err)
		}

		if !bytes.Contains(buf.Bytes(), []byte("No emails found")) {
			t.Errorf("PrintEmailList() should show empty message, got: %s", buf.String())
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"ab", 2, "ab"},
		{"abc", 2, "ab"},
		{"日本語テスト", 5, "日本..."}, // 6 runes truncated to 5 = 2 chars + "..."
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}
