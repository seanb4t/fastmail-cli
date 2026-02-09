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

func (kb keyBarModel) viewForPaneWidth(pane PaneID, width int) string {
	contextBindings := kb.bindingsForPane(pane)
	globalBindings := []keyBinding{
		{"tab", "pane"},
		{"b", "sidebar"},
		{"?", "help"},
		{"q", "quit"},
	}

	// Try full render
	all := make([]keyBinding, 0, len(contextBindings)+len(globalBindings))
	all = append(all, contextBindings...)
	all = append(all, globalBindings...)
	full := kb.render(all)
	if lipgloss.Width(full) <= width {
		return full
	}

	// Try key-only for context bindings, full for global
	var abbreviated []keyBinding
	for _, b := range contextBindings {
		abbreviated = append(abbreviated, keyBinding{b.key, ""})
	}
	abbreviated = append(abbreviated, globalBindings...)
	abbr := kb.render(abbreviated)
	if lipgloss.Width(abbr) <= width {
		return abbr
	}

	// Drop context bindings one at a time from the end
	for len(contextBindings) > 0 {
		contextBindings = contextBindings[:len(contextBindings)-1]
		abbreviated = nil
		for _, b := range contextBindings {
			abbreviated = append(abbreviated, keyBinding{b.key, ""})
		}
		abbreviated = append(abbreviated, globalBindings...)
		result := kb.render(abbreviated)
		if lipgloss.Width(result) <= width {
			return result
		}
	}

	// Just global bindings with key-only
	var keysOnly []keyBinding
	for _, b := range globalBindings {
		keysOnly = append(keysOnly, keyBinding{b.key, ""})
	}
	return kb.render(keysOnly)
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
		if b.desc != "" {
			result += kb.keyStyle.Render(b.key) + " " + kb.descStyle.Render(b.desc)
		} else {
			result += kb.keyStyle.Render(b.key)
		}
	}
	return result
}
