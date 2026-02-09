package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaneID_Constants(t *testing.T) {
	assert.Equal(t, PaneID(0), PaneMailbox)
	assert.Equal(t, PaneID(1), PaneEmailList)
	assert.Equal(t, PaneID(2), PanePreview)
}

func TestNewPaneManager(t *testing.T) {
	pm := newPaneManager()

	assert.Equal(t, PaneEmailList, pm.focus)
	assert.True(t, pm.sidebar)
	assert.Equal(t, 50, pm.splitPct)
}

func TestPaneManager_CycleFocus(t *testing.T) {
	pm := newPaneManager()
	pm.focus = PaneMailbox

	pm.cycleFocus()
	assert.Equal(t, PaneEmailList, pm.focus)

	pm.cycleFocus()
	assert.Equal(t, PanePreview, pm.focus)

	pm.cycleFocus()
	assert.Equal(t, PaneMailbox, pm.focus)
}

func TestPaneManager_CycleFocus_SkipsSidebarWhenHidden(t *testing.T) {
	pm := newPaneManager()
	pm.sidebar = false
	pm.focus = PaneEmailList

	pm.cycleFocus()
	assert.Equal(t, PanePreview, pm.focus)

	pm.cycleFocus()
	assert.Equal(t, PaneEmailList, pm.focus)
}

func TestPaneManager_ToggleSidebar(t *testing.T) {
	pm := newPaneManager()
	assert.True(t, pm.sidebar)

	pm.toggleSidebar()
	assert.False(t, pm.sidebar)

	// If focus was on sidebar, it should move
	pm.sidebar = true
	pm.focus = PaneMailbox
	pm.toggleSidebar()
	assert.Equal(t, PaneEmailList, pm.focus)
}

func TestPaneManager_AdjustSplit(t *testing.T) {
	pm := newPaneManager()
	pm.splitPct = 50

	pm.adjustSplit(10)
	assert.Equal(t, 60, pm.splitPct)

	pm.adjustSplit(-20)
	assert.Equal(t, 40, pm.splitPct)
}

func TestPaneManager_AdjustSplit_Clamped(t *testing.T) {
	pm := newPaneManager()

	pm.splitPct = 75
	pm.adjustSplit(10)
	assert.Equal(t, 80, pm.splitPct)

	pm.splitPct = 35
	pm.adjustSplit(-10)
	assert.Equal(t, 30, pm.splitPct)
}

func TestPaneManager_ComputeLayout(t *testing.T) {
	pm := newPaneManager()
	pm.sidebar = true
	pm.splitPct = 50

	layout := pm.computeLayout(120, 40)

	assert.Greater(t, layout.sidebarWidth, 0)
	assert.Equal(t, 120, layout.sidebarWidth+layout.mainWidth+1) // +1 for border
	assert.Greater(t, layout.listHeight, 0)
	assert.Greater(t, layout.previewHeight, 0)
}

func TestPaneManager_ComputeLayout_NoSidebar(t *testing.T) {
	pm := newPaneManager()
	pm.sidebar = false

	layout := pm.computeLayout(120, 40)

	assert.Equal(t, 0, layout.sidebarWidth)
	assert.Equal(t, 120, layout.mainWidth)
}

func TestPaneManager_ComputeLayout_NarrowTerminal(t *testing.T) {
	pm := newPaneManager()
	pm.sidebar = true

	layout := pm.computeLayout(60, 30)

	// Below 80 cols, sidebar should auto-collapse
	assert.Equal(t, 0, layout.sidebarWidth)
	assert.Equal(t, 60, layout.mainWidth)
}

func TestPaneManager_ComputeLayout_ShortTerminal(t *testing.T) {
	pm := newPaneManager()

	layout := pm.computeLayout(120, 20)

	// Short terminals should still have valid dimensions
	assert.Greater(t, layout.listHeight, 0)
}

func TestPaneManager_ComputeLayout_VerySmall(t *testing.T) {
	pm := newPaneManager()

	layout := pm.computeLayout(40, 15)

	// Should not panic, should produce valid layout
	assert.GreaterOrEqual(t, layout.listHeight, 0)
	assert.GreaterOrEqual(t, layout.previewHeight, 0)
}
