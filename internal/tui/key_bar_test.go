package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyBarModel_ForPane_Mailbox(t *testing.T) {
	kb := newKeyBarModel(DetectTheme())
	result := kb.viewForPane(PaneMailbox)

	assert.Contains(t, result, "enter")
	assert.Contains(t, result, "open")
	assert.Contains(t, result, "/")
}

func TestKeyBarModel_ForPane_EmailList(t *testing.T) {
	kb := newKeyBarModel(DetectTheme())
	result := kb.viewForPane(PaneEmailList)

	assert.Contains(t, result, "enter")
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "c")
	assert.Contains(t, result, "/")
}

func TestKeyBarModel_ForPane_Preview(t *testing.T) {
	kb := newKeyBarModel(DetectTheme())
	result := kb.viewForPane(PanePreview)

	assert.Contains(t, result, "j/k")
	assert.Contains(t, result, "r")
	assert.Contains(t, result, "t")
}

func TestKeyBarModel_AlwaysShowsGlobalKeys(t *testing.T) {
	kb := newKeyBarModel(DetectTheme())

	for _, pane := range []PaneID{PaneMailbox, PaneEmailList, PanePreview} {
		result := kb.viewForPane(pane)
		assert.Contains(t, result, "tab", "pane=%d should contain tab", pane)
		assert.Contains(t, result, "?", "pane=%d should contain ?", pane)
		assert.Contains(t, result, "q", "pane=%d should contain q", pane)
	}
}
