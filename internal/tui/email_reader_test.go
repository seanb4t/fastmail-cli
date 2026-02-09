package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestEmailReader_PreviewMode(t *testing.T) {
	email := fastmail.Email{
		ID:      "e1",
		Subject: "Test Subject",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
		Body:    "Hello world",
	}
	er := newEmailReaderModel(email)
	er.isPreview = true
	er.setSize(80, 15)

	er, _ = er.update(emailBodyLoadedMsg{email: email})

	v := er.view()
	assert.Contains(t, v, "Alice")
	assert.Contains(t, v, "Test Subject")
}

func TestEmailReader_PreviewMode_CompactHeaders(t *testing.T) {
	email := fastmail.Email{
		ID:      "e1",
		Subject: "Test Subject",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
		Body:    "Hello world",
	}
	er := newEmailReaderModel(email)
	er.isPreview = true
	er.setSize(80, 15)

	er, _ = er.update(emailBodyLoadedMsg{email: email})

	// Preview mode should not show the help bar
	v := er.view()
	assert.NotContains(t, v, "j/k scroll")
}

func TestEmailReader_FullscreenMode_ShowsHelpBar(t *testing.T) {
	email := fastmail.Email{
		ID:      "e1",
		Subject: "Test Subject",
		From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
		Body:    "Hello world",
	}
	er := newEmailReaderModel(email)
	er.setSize(80, 24)

	er, _ = er.update(emailBodyLoadedMsg{email: email})

	v := er.view()
	assert.Contains(t, v, "j/k scroll")
}

func TestEmailReader_QuotedTextPresent(t *testing.T) {
	email := fastmail.Email{
		ID:   "e1",
		Body: "> This is quoted\n\nMy reply",
	}
	er := newEmailReaderModel(email)
	er.setSize(80, 20)
	er, _ = er.update(emailBodyLoadedMsg{email: email})

	v := er.view()
	assert.Contains(t, v, "quoted")
	assert.Contains(t, v, "reply")
}
