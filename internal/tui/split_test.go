package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitView_Render(t *testing.T) {
	sv := newSplitView(80, 30, 50)

	topContent := "top pane"
	bottomContent := "bottom pane"

	result := sv.render(topContent, bottomContent, DetectTheme())

	assert.Contains(t, result, "top pane")
	assert.Contains(t, result, "bottom pane")
}

func TestSplitView_DividerLine(t *testing.T) {
	sv := newSplitView(80, 30, 50)
	result := sv.render("top", "bottom", DetectTheme())

	assert.Contains(t, result, "─")
}

func TestSplitView_ResizesOnPctChange(t *testing.T) {
	sv1 := newSplitView(80, 30, 30)
	sv2 := newSplitView(80, 30, 70)

	assert.Less(t, sv1.topHeight, sv2.topHeight)
}
