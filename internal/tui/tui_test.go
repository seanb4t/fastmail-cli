package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestNew(t *testing.T) {
	client := fastmail.NewClient("https://api.fastmail.com", "test-token")
	m := New(client)

	assert.Equal(t, viewMailboxList, m.view)
	assert.NotNil(t, m.client)
	assert.False(t, m.quit)
}

func TestUpdate_Quit(t *testing.T) {
	m := New(nil)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestUpdate_CtrlC(t *testing.T) {
	m := New(nil)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestUpdate_WindowSize(t *testing.T) {
	m := New(nil)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	assert.Equal(t, 120, model.width)
	assert.Equal(t, 40, model.height)
}

func TestView_Default(t *testing.T) {
	m := New(nil)
	v := m.View()

	assert.Contains(t, v, "fastmail-cli")
	assert.Contains(t, v, "q: quit")
}

func TestView_Quit(t *testing.T) {
	m := New(nil)
	m.quit = true

	assert.Empty(t, m.View())
}

func TestView_Error(t *testing.T) {
	m := New(nil)
	m.err = assert.AnError

	v := m.View()
	assert.Contains(t, v, "Error:")
}

func TestInit(t *testing.T) {
	m := New(nil)
	cmd := m.Init()

	assert.Nil(t, cmd)
}
