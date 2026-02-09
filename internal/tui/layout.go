package tui

const (
	sidebarDefaultWidth = 20
	statsBarHeight      = 2
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
	focus    PaneID
	sidebar  bool
	splitPct int // percentage of main area for email list (30-80)
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
	if pm.sidebar {
		return []PaneID{PaneMailbox, PaneEmailList, PanePreview}
	}
	return []PaneID{PaneEmailList, PanePreview}
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

	contentH := height - statsBarHeight - keyBarHeight
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
