package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func makeTestAttachments() []fastmail.Attachment {
	return []fastmail.Attachment{
		{BlobID: "b1", Name: "report.pdf", Type: "application/pdf", Size: 1048576, Disposition: "attachment"},
		{BlobID: "b2", Name: "photo.jpg", Type: "image/jpeg", Size: 2097152, Disposition: "attachment"},
		{BlobID: "b3", Name: "logo.png", Type: "image/png", Size: 512, Disposition: "inline"},
	}
}

func TestAttachmentItem_Title(t *testing.T) {
	att := fastmail.Attachment{BlobID: "b1", Name: "report.pdf", Type: "application/pdf", Size: 1024}
	item := attachmentItem{attachment: att}

	assert.Equal(t, "report.pdf", item.Title())
}

func TestAttachmentItem_Description(t *testing.T) {
	att := fastmail.Attachment{BlobID: "b1", Name: "report.pdf", Type: "application/pdf", Size: 1572864}
	item := attachmentItem{attachment: att}

	desc := item.Description()
	assert.Contains(t, desc, "application/pdf")
	assert.Contains(t, desc, "1.5 MB")
}

func TestAttachmentItem_FilterValue(t *testing.T) {
	att := fastmail.Attachment{BlobID: "b1", Name: "report.pdf", Type: "application/pdf", Size: 1024}
	item := attachmentItem{attachment: att}

	assert.Equal(t, "report.pdf", item.FilterValue())
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{1536, "1.5 KB"},
		{1572864, "1.5 MB"},
		{1610612736, "1.5 GB"},
		{0, "0 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatSize(tt.bytes))
		})
	}
}

func TestNewAttachmentPickerModel(t *testing.T) {
	attachments := makeTestAttachments()
	m := newAttachmentPickerModel("e1", attachments)

	assert.Equal(t, "e1", m.emailID)
	assert.Equal(t, "Attachments", m.list.Title)
	assert.Len(t, m.list.Items(), 3)
	assert.False(t, m.canceled)
	assert.Nil(t, m.selected)
}

func TestAttachmentPickerModel_Esc(t *testing.T) {
	attachments := makeTestAttachments()
	m := newAttachmentPickerModel("e1", attachments)
	m.setSize(80, 24)

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEscape})

	assert.True(t, m.canceled)
	assert.Nil(t, m.selected)
}

func TestAttachmentPickerModel_QuitKey(t *testing.T) {
	attachments := makeTestAttachments()
	m := newAttachmentPickerModel("e1", attachments)
	m.setSize(80, 24)

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	assert.True(t, m.canceled)
	assert.Nil(t, m.selected)
}

func TestAttachmentPickerModel_Enter(t *testing.T) {
	attachments := makeTestAttachments()
	m := newAttachmentPickerModel("e1", attachments)
	m.setSize(80, 24)

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEnter})

	require.NotNil(t, m.selected)
	assert.Equal(t, "report.pdf", m.selected.Name)
	assert.Equal(t, "b1", m.selected.BlobID)
}

func TestAttachmentPickerModel_View(t *testing.T) {
	attachments := makeTestAttachments()
	m := newAttachmentPickerModel("e1", attachments)
	m.setSize(80, 24)

	v := m.view()
	assert.NotEmpty(t, v)
	assert.Contains(t, v, "Attachments")
}

func TestNewAttachmentPickerModel_Empty(t *testing.T) {
	m := newAttachmentPickerModel("e1", []fastmail.Attachment{})

	assert.Empty(t, m.list.Items())
	assert.Equal(t, "Attachments", m.list.Title)
}
