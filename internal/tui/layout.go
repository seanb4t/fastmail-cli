package tui

import "strings"

const (
	sidebarDefaultWidth = 20
	statsBarHeight      = 1
	keyBarHeight        = 1
	splitMin            = 30
	splitMax            = 80
)

// PaneID identifies a pane in the layout.
type PaneID int

// Pane identifiers for the dashboard layout.
const (
	PaneMailbox PaneID = iota
	PaneEmailList
	PanePreview
)

// paneLayout holds computed dimensions for each zone.
type paneLayout struct {
	sidebarWidth  int
	mainWidth     int
	listHeight    int
	previewHeight int
	statsHeight   int
	keyBarHeight  int
	totalWidth    int
	totalHeight   int
}

// paneManager tracks focus, sidebar visibility, and split ratio.
type paneManager struct {
	focus      PaneID
	sidebar    bool
	hasPreview bool // true when an email is selected and preview pane has content
	splitPct   int  // percentage of main area for email list (30-80)
}

func newPaneManager() paneManager {
	return paneManager{
		focus:    PaneEmailList,
		sidebar:  true,
		splitPct: 50,
	}
}

func (pm *paneManager) cycleFocus() {
	panes := pm.visiblePanes()
	for i, p := range panes {
		if p == pm.focus {
			pm.focus = panes[(i+1)%len(panes)]
			return
		}
	}
	pm.focus = panes[0]
}

func (pm *paneManager) visiblePanes() []PaneID {
	var panes []PaneID
	if pm.sidebar {
		panes = append(panes, PaneMailbox)
	}
	panes = append(panes, PaneEmailList)
	if pm.hasPreview {
		panes = append(panes, PanePreview)
	}
	return panes
}

func (pm *paneManager) toggleSidebar() {
	pm.sidebar = !pm.sidebar
	if !pm.sidebar && pm.focus == PaneMailbox {
		pm.focus = PaneEmailList
	}
}

func (pm *paneManager) adjustSplit(delta int) {
	pm.splitPct += delta
	pm.splitPct = max(splitMin, min(splitMax, pm.splitPct))
}

func truncateLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func (pm *paneManager) computeLayout(width, height int) paneLayout {
	var sidebarW int
	// Auto-collapse sidebar on narrow terminals (< 80 cols)
	if pm.sidebar && width >= 80 {
		sidebarW = sidebarDefaultWidth
	}

	mainW := width - sidebarW
	if sidebarW > 0 && mainW > 0 {
		mainW-- // 1 char border between sidebar and main
	}
	mainW = max(mainW, 0)

	contentH := height - statsBarHeight - keyBarHeight - 2 // 2 for pane border (top+bottom)
	contentH = max(contentH, 2)

	listH := contentH * pm.splitPct / 100
	previewH := contentH - listH

	return paneLayout{
		sidebarWidth:  sidebarW,
		mainWidth:     mainW,
		listHeight:    listH,
		previewHeight: previewH,
		statsHeight:   statsBarHeight,
		keyBarHeight:  keyBarHeight,
		totalWidth:    width,
		totalHeight:   height,
	}
}
