package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme holds all color values for the TUI.
type Theme struct {
	// Chrome
	StatusBarBg lipgloss.Color
	StatusBarFg lipgloss.Color
	PaneBorder  lipgloss.Color
	FocusBorder lipgloss.Color
	KeyBarKey   lipgloss.Color
	KeyBarDesc  lipgloss.Color

	// Email states
	Unread   lipgloss.Color
	Read     lipgloss.Color
	Flagged  lipgloss.Color
	Selected lipgloss.Color

	// Content
	HeaderLabel lipgloss.Color
	HeaderValue lipgloss.Color
	QuotedText  lipgloss.Color
	Link        lipgloss.Color

	// Stats
	StatValue lipgloss.Color
	QuotaLow  lipgloss.Color
	QuotaMed  lipgloss.Color
	QuotaHigh lipgloss.Color
}

// DarkTheme returns a Catppuccin Mocha-inspired dark palette.
func DarkTheme() Theme {
	return Theme{
		StatusBarBg: lipgloss.Color("#1e1e2e"),
		StatusBarFg: lipgloss.Color("#cdd6f4"),
		PaneBorder:  lipgloss.Color("#585b70"),
		FocusBorder: lipgloss.Color("#b4befe"),
		KeyBarKey:   lipgloss.Color("#b4befe"),
		KeyBarDesc:  lipgloss.Color("#6c7086"),

		Unread:   lipgloss.Color("#cdd6f4"),
		Read:     lipgloss.Color("#6c7086"),
		Flagged:  lipgloss.Color("#fab387"),
		Selected: lipgloss.Color("#313244"),

		HeaderLabel: lipgloss.Color("#89b4fa"),
		HeaderValue: lipgloss.Color("#cdd6f4"),
		QuotedText:  lipgloss.Color("#6c7086"),
		Link:        lipgloss.Color("#89b4fa"),

		StatValue: lipgloss.Color("#cdd6f4"),
		QuotaLow:  lipgloss.Color("#a6e3a1"),
		QuotaMed:  lipgloss.Color("#f9e2af"),
		QuotaHigh: lipgloss.Color("#f38ba8"),
	}
}

// LightTheme returns a Catppuccin Latte-inspired light palette.
func LightTheme() Theme {
	return Theme{
		StatusBarBg: lipgloss.Color("#e6e9ef"),
		StatusBarFg: lipgloss.Color("#4c4f69"),
		PaneBorder:  lipgloss.Color("#9ca0b0"),
		FocusBorder: lipgloss.Color("#1e66f5"),
		KeyBarKey:   lipgloss.Color("#1e66f5"),
		KeyBarDesc:  lipgloss.Color("#9ca0b0"),

		Unread:   lipgloss.Color("#4c4f69"),
		Read:     lipgloss.Color("#9ca0b0"),
		Flagged:  lipgloss.Color("#fe640b"),
		Selected: lipgloss.Color("#ccd0da"),

		HeaderLabel: lipgloss.Color("#1e66f5"),
		HeaderValue: lipgloss.Color("#4c4f69"),
		QuotedText:  lipgloss.Color("#9ca0b0"),
		Link:        lipgloss.Color("#1e66f5"),

		StatValue: lipgloss.Color("#4c4f69"),
		QuotaLow:  lipgloss.Color("#40a02b"),
		QuotaMed:  lipgloss.Color("#df8e1d"),
		QuotaHigh: lipgloss.Color("#d20f39"),
	}
}

// DetectTheme returns a theme based on terminal background.
func DetectTheme() Theme {
	if lipgloss.HasDarkBackground() {
		return DarkTheme()
	}
	return LightTheme()
}
