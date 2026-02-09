package tui

import "github.com/charmbracelet/lipgloss"

type keyBinding struct {
	key  string
	desc string
}

type keyBarModel struct {
	keyStyle  lipgloss.Style
	descStyle lipgloss.Style
	sepStyle  lipgloss.Style
}

func newKeyBarModel(theme Theme) keyBarModel {
	return keyBarModel{
		keyStyle:  lipgloss.NewStyle().Bold(true).Foreground(theme.KeyBarKey),
		descStyle: lipgloss.NewStyle().Foreground(theme.KeyBarDesc),
		sepStyle:  lipgloss.NewStyle().Foreground(theme.PaneBorder),
	}
}

func (kb keyBarModel) viewForPane(pane PaneID) string {
	bindings := kb.bindingsForPane(pane)
	bindings = append(bindings,
		keyBinding{"tab", "pane"},
		keyBinding{"b", "sidebar"},
		keyBinding{"?", "help"},
		keyBinding{"q", "quit"},
	)
	return kb.render(bindings)
}

func (kb keyBarModel) bindingsForPane(pane PaneID) []keyBinding {
	switch pane {
	case PaneMailbox:
		return []keyBinding{
			{"enter", "open"},
			{"/", "filter"},
		}
	case PaneEmailList:
		return []keyBinding{
			{"enter", "read"},
			{"a", "archive"},
			{"f", "flag"},
			{"c", "compose"},
			{"/", "search"},
		}
	case PanePreview:
		return []keyBinding{
			{"j/k", "scroll"},
			{"r", "reply"},
			{"R", "reply all"},
			{"a", "archive"},
			{"f", "flag"},
			{"t", "thread"},
		}
	}
	return nil
}

func (kb keyBarModel) render(bindings []keyBinding) string {
	var result string
	for i, b := range bindings {
		if i > 0 {
			result += kb.sepStyle.Render("  ")
		}
		result += kb.keyStyle.Render(b.key) + " " + kb.descStyle.Render(b.desc)
	}
	return result
}
