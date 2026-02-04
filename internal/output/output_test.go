package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestIsTerminal_WithRealStdout(_ *testing.T) {
	// This test documents behavior - actual TTY detection depends on environment
	// In tests, stdout is typically not a terminal
	result := IsTerminal(os.Stdout)
	// We just verify it returns a boolean without panicking
	_ = result
}

func TestIsTerminal_WithPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()

	// Pipes are not terminals
	if IsTerminal(w) {
		t.Error("IsTerminal should return false for a pipe")
	}
}

func TestFormatJSON_SingleValue(t *testing.T) {
	data := map[string]string{"key": "value"}

	result, err := FormatJSON(data)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	expected := "{\n  \"key\": \"value\"\n}"
	if result != expected {
		t.Errorf("FormatJSON = %q, want %q", result, expected)
	}
}

func TestFormatJSON_Email(t *testing.T) {
	email := fastmail.Email{
		ID:      "abc123",
		Subject: "Test Subject",
	}

	result, err := FormatJSON(email)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	// Should contain expected fields with indentation
	if !bytes.Contains([]byte(result), []byte(`"ID": "abc123"`)) {
		t.Errorf("FormatJSON output should contain ID field, got: %s", result)
	}
	if !bytes.Contains([]byte(result), []byte(`"Subject": "Test Subject"`)) {
		t.Errorf("FormatJSON output should contain Subject field, got: %s", result)
	}
}

func TestFormatJSON_InvalidValue(t *testing.T) {
	// Channels cannot be marshaled to JSON
	ch := make(chan int)

	_, err := FormatJSON(ch)
	if err == nil {
		t.Error("FormatJSON should return error for unmarshallable value")
	}
}

func TestFormatTable_Empty(t *testing.T) {
	headers := []string{"ID", "Subject"}
	rows := [][]string{}

	result := FormatTable(headers, rows)

	// Should still show headers
	if !bytes.Contains([]byte(result), []byte("ID")) {
		t.Error("FormatTable should include headers even with no rows")
	}
}

func TestFormatTable_WithRows(t *testing.T) {
	headers := []string{"ID", "Subject", "From"}
	rows := [][]string{
		{"abc123", "Hello World", "alice@example.com"},
		{"def456", "Re: Hello World", "bob@example.com"},
	}

	result := FormatTable(headers, rows)

	// Verify headers present
	if !bytes.Contains([]byte(result), []byte("ID")) {
		t.Errorf("FormatTable should include ID header, got: %s", result)
	}
	if !bytes.Contains([]byte(result), []byte("Subject")) {
		t.Errorf("FormatTable should include Subject header, got: %s", result)
	}

	// Verify data present
	if !bytes.Contains([]byte(result), []byte("abc123")) {
		t.Errorf("FormatTable should include row data, got: %s", result)
	}
	if !bytes.Contains([]byte(result), []byte("Hello World")) {
		t.Errorf("FormatTable should include row data, got: %s", result)
	}
}

func TestFormatTable_AlignColumns(t *testing.T) {
	headers := []string{"ID", "Name"}
	rows := [][]string{
		{"a", "Short"},
		{"longer_id", "Much Longer Name"},
	}

	result := FormatTable(headers, rows)

	// The table should have consistent column widths (visual alignment)
	// This is a basic check that longer values don't break formatting
	if len(result) == 0 {
		t.Error("FormatTable returned empty string")
	}
}

func TestPrinter_PrintEmail_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, Options{JSON: true})

	email := fastmail.Email{
		ID:      "test-id",
		Subject: "Test Subject",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
	}

	err := p.PrintEmail(email)
	if err != nil {
		t.Fatalf("PrintEmail returned error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte(`"ID": "test-id"`)) {
		t.Errorf("PrintEmail JSON should contain ID, got: %s", output)
	}
}

func TestPrinter_PrintEmail_Quiet(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, Options{Quiet: true})

	email := fastmail.Email{
		ID:      "test-id",
		Subject: "Test Subject",
	}

	err := p.PrintEmail(email)
	if err != nil {
		t.Fatalf("PrintEmail returned error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("PrintEmail with Quiet should produce no output, got: %s", buf.String())
	}
}

func TestPrinter_PrintEmailList_Table(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, Options{})

	emails := []fastmail.Email{
		{ID: "id1", Subject: "First Email", From: fastmail.EmailAddress{Email: "a@example.com"}},
		{ID: "id2", Subject: "Second Email", From: fastmail.EmailAddress{Email: "b@example.com"}},
	}

	err := p.PrintEmailList(emails)
	if err != nil {
		t.Fatalf("PrintEmailList returned error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("First Email")) {
		t.Errorf("PrintEmailList should contain first subject, got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("Second Email")) {
		t.Errorf("PrintEmailList should contain second subject, got: %s", output)
	}
}

func TestPrinter_PrintEmailList_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, Options{JSON: true})

	emails := []fastmail.Email{
		{ID: "id1", Subject: "First Email"},
		{ID: "id2", Subject: "Second Email"},
	}

	err := p.PrintEmailList(emails)
	if err != nil {
		t.Fatalf("PrintEmailList returned error: %v", err)
	}

	output := buf.String()
	// Should be a JSON array
	if output[0] != '[' {
		t.Errorf("PrintEmailList JSON should start with '[', got: %s", output)
	}
}

func TestPrinter_PrintEmailList_Empty(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, Options{})

	err := p.PrintEmailList([]fastmail.Email{})
	if err != nil {
		t.Fatalf("PrintEmailList returned error: %v", err)
	}

	// Should still produce headers or empty array indication
	if buf.Len() == 0 {
		t.Error("PrintEmailList should produce some output even for empty list")
	}
}
