package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestDarkTheme_HasAllColors(t *testing.T) {
	theme := DarkTheme()

	assert.NotEqual(t, lipgloss.Color(""), theme.FocusBorder)
	assert.NotEqual(t, lipgloss.Color(""), theme.PaneBorder)
	assert.NotEqual(t, lipgloss.Color(""), theme.Unread)
	assert.NotEqual(t, lipgloss.Color(""), theme.Read)
	assert.NotEqual(t, lipgloss.Color(""), theme.Flagged)
	assert.NotEqual(t, lipgloss.Color(""), theme.StatusBarBg)
	assert.NotEqual(t, lipgloss.Color(""), theme.KeyBarKey)
	assert.NotEqual(t, lipgloss.Color(""), theme.KeyBarDesc)
	assert.NotEqual(t, lipgloss.Color(""), theme.QuotaLow)
	assert.NotEqual(t, lipgloss.Color(""), theme.QuotaMed)
	assert.NotEqual(t, lipgloss.Color(""), theme.QuotaHigh)
}

func TestLightTheme_HasAllColors(t *testing.T) {
	theme := LightTheme()

	assert.NotEqual(t, lipgloss.Color(""), theme.FocusBorder)
	assert.NotEqual(t, lipgloss.Color(""), theme.PaneBorder)
	assert.NotEqual(t, lipgloss.Color(""), theme.Unread)
	assert.NotEqual(t, lipgloss.Color(""), theme.Read)
	assert.NotEqual(t, lipgloss.Color(""), theme.Flagged)
}

func TestDarkTheme_DiffersFromLight(t *testing.T) {
	dark := DarkTheme()
	light := LightTheme()

	assert.NotEqual(t, dark.StatusBarBg, light.StatusBarBg)
	assert.NotEqual(t, dark.Unread, light.Unread)
}

func TestDetectTheme_ReturnsTheme(t *testing.T) {
	theme := DetectTheme()
	assert.NotEqual(t, lipgloss.Color(""), theme.FocusBorder)
}
