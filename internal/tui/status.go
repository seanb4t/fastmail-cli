package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statusSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	statusErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// statusClearMsg is fired by a tick to auto-clear the status.
type statusClearMsg struct{}

// statusModel provides temporary feedback messages.
type statusModel struct {
	message string
	isError bool
	visible bool
}

func (s *statusModel) setStatus(msg string, isError bool) tea.Cmd {
	s.message = msg
	s.isError = isError
	s.visible = true
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (s *statusModel) update(msg tea.Msg) {
	if _, ok := msg.(statusClearMsg); ok {
		s.visible = false
		s.message = ""
	}
}

func (s statusModel) view() string {
	if !s.visible || s.message == "" {
		return ""
	}
	style := statusSuccessStyle
	if s.isError {
		style = statusErrorStyle
	}
	return style.Render("  " + s.message)
}
