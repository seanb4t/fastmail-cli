# TUI Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform the TUI from fullscreen view-swapping into a persistent pane-based dashboard with theme system, collapsible sidebar, adjustable split, stats bar, and contextual keybinding footer.

**Architecture:** Pane manager coordinates focus/layout. Theme struct provides dark/light palettes via auto-detection. Sub-models become pane-aware (accept constrained dimensions). Root model becomes layout coordinator instead of view switcher.

**Tech Stack:** bubbletea, bubbles (list, viewport, textinput, textarea), lipgloss, glamour. Go 1.25+.

**Design document (immutable):** `docs/plans/2026-02-09-tui-redesign-design.md`

---

## Phase 1: Theme + Layout Shell

### Task 1.1: Create Theme System

**Files:**
- Create: `internal/tui/theme.go`
- Test: `internal/tui/theme_test.go`

**Step 1: Write the failing test**

```go
// internal/tui/theme_test.go
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

    // At minimum, the base colors should differ
    assert.NotEqual(t, dark.StatusBarBg, light.StatusBarBg)
    assert.NotEqual(t, dark.Unread, light.Unread)
}

func TestDetectTheme_ReturnsTheme(t *testing.T) {
    // DetectTheme should always return a valid theme
    theme := DetectTheme()
    assert.NotEqual(t, lipgloss.Color(""), theme.FocusBorder)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestDarkTheme_HasAllColors ./internal/tui/`
Expected: FAIL — `DarkTheme` undefined

**Step 3: Write minimal implementation**

```go
// internal/tui/theme.go
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
        FocusBorder: lipgloss.Color("#b4befe"), // lavender
        KeyBarKey:   lipgloss.Color("#b4befe"),
        KeyBarDesc:  lipgloss.Color("#6c7086"),

        Unread:   lipgloss.Color("#cdd6f4"), // white text
        Read:     lipgloss.Color("#6c7086"), // overlay gray
        Flagged:  lipgloss.Color("#fab387"), // peach
        Selected: lipgloss.Color("#313244"), // surface0

        HeaderLabel: lipgloss.Color("#89b4fa"), // blue
        HeaderValue: lipgloss.Color("#cdd6f4"),
        QuotedText:  lipgloss.Color("#6c7086"),
        Link:        lipgloss.Color("#89b4fa"),

        StatValue: lipgloss.Color("#cdd6f4"),
        QuotaLow:  lipgloss.Color("#a6e3a1"), // green
        QuotaMed:  lipgloss.Color("#f9e2af"), // yellow
        QuotaHigh: lipgloss.Color("#f38ba8"), // red
    }
}

// LightTheme returns a Catppuccin Latte-inspired light palette.
func LightTheme() Theme {
    return Theme{
        StatusBarBg: lipgloss.Color("#e6e9ef"),
        StatusBarFg: lipgloss.Color("#4c4f69"),
        PaneBorder:  lipgloss.Color("#9ca0b0"),
        FocusBorder: lipgloss.Color("#1e66f5"), // blue
        KeyBarKey:   lipgloss.Color("#1e66f5"),
        KeyBarDesc:  lipgloss.Color("#9ca0b0"),

        Unread:   lipgloss.Color("#4c4f69"), // dark text
        Read:     lipgloss.Color("#9ca0b0"), // overlay gray
        Flagged:  lipgloss.Color("#fe640b"), // orange
        Selected: lipgloss.Color("#ccd0da"), // surface0

        HeaderLabel: lipgloss.Color("#1e66f5"),
        HeaderValue: lipgloss.Color("#4c4f69"),
        QuotedText:  lipgloss.Color("#9ca0b0"),
        Link:        lipgloss.Color("#1e66f5"),

        StatValue: lipgloss.Color("#4c4f69"),
        QuotaLow:  lipgloss.Color("#40a02b"), // green
        QuotaMed:  lipgloss.Color("#df8e1d"), // yellow
        QuotaHigh: lipgloss.Color("#d20f39"), // red
    }
}

// DetectTheme returns a theme based on terminal background.
func DetectTheme() Theme {
    if lipgloss.HasDarkBackground() {
        return DarkTheme()
    }
    return LightTheme()
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestDarkTheme ./internal/tui/ && go test -run TestLightTheme ./internal/tui/ && go test -run TestDetectTheme ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add theme system with dark/light auto-detection
```

---

### Task 1.2: Create Pane Manager

**Files:**
- Create: `internal/tui/layout.go`
- Test: `internal/tui/layout_test.go`

**Step 1: Write the failing test**

```go
// internal/tui/layout_test.go
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

    // Sidebar should have width
    assert.Greater(t, layout.sidebarWidth, 0)
    // Main area fills the rest
    assert.Equal(t, 120, layout.sidebarWidth+layout.mainWidth+1) // +1 for border
    // Split divides main height (minus stats bar and key bar)
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
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestPaneID_Constants ./internal/tui/`
Expected: FAIL — `PaneID` undefined

**Step 3: Write minimal implementation**

```go
// internal/tui/layout.go
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

const (
    PaneMailbox   PaneID = iota
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
    if pm.splitPct < splitMin {
        pm.splitPct = splitMin
    }
    if pm.splitPct > splitMax {
        pm.splitPct = splitMax
    }
}

func (pm *paneManager) computeLayout(width, height int) paneLayout {
    var sidebarW int
    if pm.sidebar {
        sidebarW = sidebarDefaultWidth
    }

    // Account for border between sidebar and main
    mainW := width - sidebarW
    if pm.sidebar && mainW > 0 {
        mainW-- // 1 char border
    }

    // Vertical: stats bar + main content + key bar
    contentH := height - statsBarHeight - keyBarHeight
    if contentH < 2 {
        contentH = 2
    }

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
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestPane ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add pane manager with focus cycling and layout computation
```

---

### Task 1.3: Create Keybinding Bar

**Files:**
- Create: `internal/tui/key_bar.go`
- Test: `internal/tui/key_bar_test.go`

**Step 1: Write the failing test**

```go
// internal/tui/key_bar_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestKeyBarModel ./internal/tui/`
Expected: FAIL — `newKeyBarModel` undefined

**Step 3: Write minimal implementation**

```go
// internal/tui/key_bar.go
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
    // Always append global keys
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
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestKeyBarModel ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add context-sensitive keybinding bar
```

---

### Task 1.4: Refactor Root Model to Layout Coordinator

**Files:**
- Modify: `internal/tui/tui.go`
- Modify: `internal/tui/tui_test.go`

This is the largest task in Phase 1. The root `Model` gains `paneManager`, `Theme`, and `keyBarModel` fields. The `View()` method renders the pane layout instead of switching views. The `Update()` method routes to the focused pane.

**Step 1: Write failing tests for the new layout behavior**

Add to `internal/tui/tui_test.go`:

```go
func TestNew_HasPaneManager(t *testing.T) {
    client := fastmail.NewClient("https://api.fastmail.com", "test-token")
    m := New(client)

    assert.Equal(t, PaneEmailList, m.panes.focus)
    assert.True(t, m.panes.sidebar)
}

func TestNew_HasTheme(t *testing.T) {
    client := fastmail.NewClient("https://api.fastmail.com", "test-token")
    m := New(client)

    assert.NotEqual(t, lipgloss.Color(""), m.theme.FocusBorder)
}

func TestUpdate_Tab_CyclesFocus(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.panes.focus = PaneMailbox

    updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
    model := updated.(Model)

    assert.Equal(t, PaneEmailList, model.panes.focus)
}

func TestUpdate_B_TogglesSidebar(t *testing.T) {
    m := New(nil)
    m.connecting = false

    updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
    model := updated.(Model)

    assert.False(t, model.panes.sidebar)
}

func TestUpdate_Plus_AdjustsSplit(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.panes.splitPct = 50

    updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
    model := updated.(Model)

    assert.Equal(t, 60, model.panes.splitPct)
}

func TestUpdate_Minus_AdjustsSplit(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.panes.splitPct = 50

    updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
    model := updated.(Model)

    assert.Equal(t, 40, model.panes.splitPct)
}

func TestView_Layout_RendersPanes(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.mailboxList.loading = false
    m.mailboxList.setSize(20, 35)

    el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
    el.loading = false
    el.setSize(80, 20)
    m.emailList = &el

    v := m.View()
    // Should contain keybinding bar indicators
    assert.Contains(t, v, "tab")
    assert.Contains(t, v, "?")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run TestNew_HasPaneManager ./internal/tui/`
Expected: FAIL — `m.panes` undefined

**Step 3: Implement layout coordinator refactor**

Modify `internal/tui/tui.go`:

1. Add fields to `Model`:
   - `panes paneManager`
   - `theme Theme`
   - `keyBar keyBarModel`

2. Update `New()` to initialize them.

3. Add layout key handling in `handleGlobalKeys()`:
   - `tab` → `pm.cycleFocus()`
   - `b` → `pm.toggleSidebar()` (when not filtering)
   - `+`/`-` → `pm.adjustSplit(±10)` (when not filtering)

4. Update `View()` to compose panes using `lipgloss.JoinHorizontal`/`JoinVertical` with borders. Focused pane gets `theme.FocusBorder`, others get `theme.PaneBorder`.

5. Update `handleWindowSize()` to use `pm.computeLayout()` and dispatch constrained sizes to sub-models.

6. Keep existing fullscreen views (compose, move picker, thread, attachment picker, help overlay) as modal overlays that render on top of the layout when active.

**Step 4: Run all existing tests + new tests**

Run: `go test -race ./internal/tui/`
Expected: PASS — all existing tests still pass, new layout tests pass

**Step 5: Commit**

```
feat(tui): refactor root model to pane-based layout coordinator
```

---

### Task 1.5: Make Mailbox List Pane-Aware

**Files:**
- Modify: `internal/tui/mailbox_list.go`
- Modify: `internal/tui/mailbox_list_test.go`

**Step 1: Write failing test**

Add to `internal/tui/mailbox_list_test.go`:

```go
func TestMailboxList_ConstrainedSize(t *testing.T) {
    ml := newMailboxListModel()
    ml.setSize(20, 30) // narrow sidebar width

    assert.Equal(t, 20, ml.list.Width())
}

func TestMailboxList_UnreadCountStyling(t *testing.T) {
    ml := newMailboxListModel()
    mb := fastmail.Mailbox{Name: "Inbox", UnreadEmails: 5, TotalEmails: 42}
    item := mailboxItem{mailbox: mb}

    title := item.Title()
    assert.Contains(t, title, "Inbox")
    assert.Contains(t, title, "(5)")
}
```

**Step 2: Run tests to verify current state**

Run: `go test -run TestMailboxList ./internal/tui/`
Expected: These should already pass (existing behavior). If not, adjust.

**Step 3: Minimal implementation**

The mailbox list already accepts constrained sizes via `setSize()`. Main changes:
- Remove fullscreen assumptions in `view()` (no leading `\n  ` padding)
- Ensure the list renders cleanly within its allocated width

**Step 4: Run tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
refactor(tui): make mailbox list pane-aware for constrained rendering
```

---

### Task 1.6: Make Email List Pane-Aware

**Files:**
- Modify: `internal/tui/email_list.go`
- Modify: `internal/tui/email_list_test.go`

**Step 1: Write failing test**

Add to `internal/tui/email_list_test.go`:

```go
func TestEmailList_ConstrainedSize(t *testing.T) {
    el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
    el.setSize(80, 20) // constrained height for split

    assert.Equal(t, 80, el.list.Width())
}
```

**Step 2: Run test**

Run: `go test -run TestEmailList_ConstrainedSize ./internal/tui/`
Expected: Should pass (existing `setSize` works). Adjust if needed.

**Step 3: Minimal implementation**

- Remove fullscreen padding in `view()`
- Ensure status/search renders within allocated space

**Step 4: Run all tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
refactor(tui): make email list pane-aware for constrained rendering
```

---

### Task 1.7: Phase 1 BDD Integration Test

**Files:**
- Create: `internal/tui/bdd_layout_test.go`

**Step 1: Write BDD-style integration test**

```go
// internal/tui/bdd_layout_test.go
package tui

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// BDD: As a user, when I launch the TUI, I see a pane-based layout
// with mailbox sidebar, email list, and keybinding bar.
func TestBDD_Layout_InitialState(t *testing.T) {
    // Given: a connected client with mailboxes loaded
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40

    mailboxes := []fastmail.Mailbox{
        {ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, UnreadEmails: 5},
        {ID: "2", Name: "Sent", Role: fastmail.RoleSent},
    }
    m, _ = applyUpdate(m, mailboxesLoadedMsg{mailboxes: mailboxes})

    // When: I view the TUI
    v := m.View()

    // Then: I see the keybinding bar
    assert.Contains(t, v, "tab")
    assert.Contains(t, v, "?")
}

// BDD: As a user, I can toggle the sidebar with 'b'.
func TestBDD_Layout_ToggleSidebar(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    require.True(t, m.panes.sidebar)

    // When: I press 'b'
    m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

    // Then: sidebar is hidden
    assert.False(t, m.panes.sidebar)
}

// BDD: As a user, I can cycle focus with Tab.
func TestBDD_Layout_CycleFocus(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.panes.focus = PaneMailbox

    // When: I press Tab
    m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyTab})

    // Then: focus moves to email list
    assert.Equal(t, PaneEmailList, m.panes.focus)
}

// BDD: As a user, I can adjust the split with +/-.
func TestBDD_Layout_AdjustSplit(t *testing.T) {
    m := New(nil)
    m.connecting = false
    initial := m.panes.splitPct

    m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
    assert.Greater(t, m.panes.splitPct, initial)

    m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
    assert.Equal(t, initial, m.panes.splitPct)
}

// applyUpdate is a test helper that applies a message and returns the typed model.
func applyUpdate(m Model, msg tea.Msg) (Model, tea.Cmd) {
    updated, cmd := m.Update(msg)
    return updated.(Model), cmd
}
```

**Step 2: Run tests**

Run: `go test -run TestBDD_Layout ./internal/tui/`
Expected: PASS (if Phase 1 tasks are complete)

**Step 3: Commit**

```
test(tui): add BDD integration tests for pane layout (Phase 1)
```

---

## Phase 2: Inline Preview + Split

### Task 2.1: Create Adjustable Horizontal Splitter

**Files:**
- Create: `internal/tui/split.go`
- Test: `internal/tui/split_test.go`

**Step 1: Write the failing test**

```go
// internal/tui/split_test.go
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

    // Should contain a horizontal divider
    assert.Contains(t, result, "─")
}

func TestSplitView_ResizesOnPctChange(t *testing.T) {
    sv1 := newSplitView(80, 30, 30)
    sv2 := newSplitView(80, 30, 70)

    // 30% split should give less top space than 70%
    assert.Less(t, sv1.topHeight, sv2.topHeight)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestSplitView ./internal/tui/`
Expected: FAIL — `newSplitView` undefined

**Step 3: Write minimal implementation**

```go
// internal/tui/split.go
package tui

import (
    "strings"

    "github.com/charmbracelet/lipgloss"
)

// splitView renders two content areas with a horizontal divider.
type splitView struct {
    width     int
    topHeight int
    botHeight int
}

func newSplitView(width, totalHeight, splitPct int) splitView {
    topH := totalHeight * splitPct / 100
    botH := totalHeight - topH - 1 // 1 for divider
    if botH < 0 {
        botH = 0
    }
    return splitView{
        width:     width,
        topHeight: topH,
        botHeight: botH,
    }
}

func (sv splitView) render(top, bottom string, theme Theme) string {
    topStyle := lipgloss.NewStyle().Width(sv.width).Height(sv.topHeight)
    botStyle := lipgloss.NewStyle().Width(sv.width).Height(sv.botHeight)
    divider := lipgloss.NewStyle().
        Foreground(theme.PaneBorder).
        Render(strings.Repeat("─", sv.width))

    return lipgloss.JoinVertical(lipgloss.Left,
        topStyle.Render(top),
        divider,
        botStyle.Render(bottom),
    )
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestSplitView ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add adjustable horizontal split view component
```

---

### Task 2.2: Refactor Email Reader as Preview Pane

**Files:**
- Modify: `internal/tui/email_reader.go`
- Modify: `internal/tui/email_reader_test.go` (if exists, else create)

**Step 1: Write failing test**

Create or add to `internal/tui/email_reader_test.go`:

```go
func TestEmailReader_PreviewMode(t *testing.T) {
    email := fastmail.Email{
        ID:      "e1",
        Subject: "Test Subject",
        From:    fastmail.EmailAddress{Name: "Alice", Email: "alice@example.com"},
        Body:    "Hello world",
    }
    er := newEmailReaderModel(email)
    er.setSize(80, 15) // constrained height for preview

    // Simulate body loaded
    er, _ = er.update(emailBodyLoadedMsg{email: email})

    v := er.view()
    assert.Contains(t, v, "Alice")
    assert.Contains(t, v, "Test Subject")
}

func TestEmailReader_QuotedTextPresent(t *testing.T) {
    email := fastmail.Email{
        ID:   "e1",
        Body: "> This is quoted\n\nMy reply",
    }
    er := newEmailReaderModel(email)
    er.setSize(80, 20)
    er, _ = er.update(emailBodyLoadedMsg{email: email})

    v := er.view()
    assert.Contains(t, v, "quoted")
    assert.Contains(t, v, "reply")
}
```

**Step 2: Run tests**

Run: `go test -run TestEmailReader ./internal/tui/`
Expected: FAIL or PASS depending on whether file exists. Adjust.

**Step 3: Implement preview mode**

The email reader already supports constrained sizes. Key changes:
- Ensure `view()` renders cleanly at small heights (15-20 lines)
- Headers should be compact in preview mode
- Viewport height adapts to constrained dimensions
- Add `isPreview bool` field to toggle compact vs fullscreen header rendering

**Step 4: Run all tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
refactor(tui): adapt email reader for inline preview pane rendering
```

---

### Task 2.3: Wire Preview into Layout

**Files:**
- Modify: `internal/tui/tui.go`
- Modify: `internal/tui/tui_test.go`

**Step 1: Write failing test**

Add to `internal/tui/tui_test.go`:

```go
func TestUpdate_EmailSelected_OpensPreview(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.view = viewEmailList
    el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
    el.loading = false
    m.emailList = &el

    email := fastmail.Email{ID: "e1", Subject: "Test"}
    el.selected = &email
    m.emailList = &el

    // When email is selected, preview pane should be populated
    result, cmd := m.updateEmailList(tea.Msg(nil))
    model := result.(Model)

    // Preview should be initialized
    require.NotNil(t, model.emailReader)
    // Should still be in the dashboard layout (not fullscreen reader)
    assert.Equal(t, viewEmailList, model.view)
    // Should fetch the body
    assert.NotNil(t, cmd)
}

func TestUpdate_Enter_InPreview_OpensFullscreen(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.panes.focus = PanePreview
    email := fastmail.Email{ID: "e1", Subject: "Test"}
    er := newEmailReaderModel(email)
    er.setSize(80, 20)
    m.emailReader = &er

    updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}})
    model := updated.(Model)

    assert.Equal(t, viewEmailReader, model.view)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run TestUpdate_EmailSelected_OpensPreview ./internal/tui/`
Expected: FAIL — current behavior switches to `viewEmailReader`

**Step 3: Implement preview wiring**

In `updateEmailList`:
- When `el.selected != nil`, populate `m.emailReader` but do NOT switch `m.view` to `viewEmailReader`. Stay in the layout view.
- `Enter` on the preview pane (when focused) switches to fullscreen `viewEmailReader`.
- Preview pane renders via `emailReader.view()` within the split.

**Step 4: Run all tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): wire email preview into dashboard layout split
```

---

### Task 2.4: Phase 2 BDD Integration Test

**Files:**
- Create or append: `internal/tui/bdd_preview_test.go`

**Step 1: Write BDD test**

```go
// internal/tui/bdd_preview_test.go
package tui

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// BDD: As a user, when I select an email in the list,
// I see a preview in the bottom pane without leaving the dashboard.
func TestBDD_Preview_SelectShowsPreview(t *testing.T) {
    // Given: email list with loaded emails
    m := setupDashboardWithEmails(t)
    require.NotNil(t, m.emailList)

    // When: I select the first email
    email := fastmail.Email{ID: "e1", Subject: "Hello", Body: "World",
        From: fastmail.EmailAddress{Name: "Alice", Email: "alice@test.com"}}
    m.emailList.selected = &email
    m, _ = applyUpdate(m, tea.Msg(nil))

    // Then: preview pane is populated
    require.NotNil(t, m.emailReader)
    // And: I'm still in the dashboard view, not fullscreen reader
    assert.NotEqual(t, viewEmailReader, m.view)
}

// BDD: As a user, I can adjust the split between list and preview.
func TestBDD_Preview_AdjustSplit(t *testing.T) {
    m := setupDashboardWithEmails(t)
    initial := m.panes.splitPct

    m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
    assert.Greater(t, m.panes.splitPct, initial)
}

// BDD: As a user, pressing Enter on preview opens fullscreen reader.
func TestBDD_Preview_EnterOpensFullscreen(t *testing.T) {
    m := setupDashboardWithEmails(t)
    m.panes.focus = PanePreview
    email := fastmail.Email{ID: "e1", Subject: "Test"}
    er := newEmailReaderModel(email)
    er.setSize(80, 20)
    m.emailReader = &er

    m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyEnter})

    assert.Equal(t, viewEmailReader, m.view)
}

func setupDashboardWithEmails(t *testing.T) Model {
    t.Helper()
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40

    el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
    el.loading = false
    el.setSize(80, 20)
    emails := []fastmail.Email{
        {ID: "e1", Subject: "Hello", From: fastmail.EmailAddress{Email: "a@b.com"}},
        {ID: "e2", Subject: "World", From: fastmail.EmailAddress{Email: "c@d.com"}},
    }
    el, _ = el.update(emailsLoadedMsg{emails: emails})
    m.emailList = &el
    m.view = viewEmailList
    return m
}
```

**Step 2: Run tests**

Run: `go test -run TestBDD_Preview ./internal/tui/`
Expected: PASS (if Phase 2 tasks are complete)

**Step 3: Commit**

```
test(tui): add BDD integration tests for preview pane (Phase 2)
```

---

## Phase 3: Rich Email List + Stats Bar

### Task 3.1: Custom Email Row Renderer

**Files:**
- Modify: `internal/tui/email_list.go`
- Modify: `internal/tui/email_list_test.go`

**Step 1: Write failing test**

Add to `internal/tui/email_list_test.go`:

```go
func TestEmailItem_RichTitle_Unread(t *testing.T) {
    theme := DarkTheme()
    item := emailItem{email: fastmail.Email{
        Subject:  "Important message",
        From:     fastmail.EmailAddress{Name: "Alice"},
        Keywords: []string{}, // unread
    }}

    title := item.richTitle(theme, 80)
    // Unread should NOT have the dimmed Read color
    assert.Contains(t, title, "Alice")
    assert.Contains(t, title, "Important message")
}

func TestEmailItem_RichTitle_Flagged(t *testing.T) {
    theme := DarkTheme()
    item := emailItem{email: fastmail.Email{
        Subject:  "Starred",
        From:     fastmail.EmailAddress{Name: "Bob"},
        Keywords: []string{"$flagged"},
    }}

    title := item.richTitle(theme, 80)
    assert.Contains(t, title, "Bob")
    // Should contain a flag indicator
    assert.Contains(t, title, "★")
}

func TestEmailItem_RichTitle_Read(t *testing.T) {
    theme := DarkTheme()
    item := emailItem{email: fastmail.Email{
        Subject:  "Old message",
        From:     fastmail.EmailAddress{Name: "Carol"},
        Keywords: []string{"$seen"},
    }}

    title := item.richTitle(theme, 80)
    assert.Contains(t, title, "Carol")
}
```

**Step 2: Run tests**

Run: `go test -run TestEmailItem_RichTitle ./internal/tui/`
Expected: FAIL — `richTitle` undefined

**Step 3: Implement rich row renderer**

Add to `email_list.go`:

```go
func (i emailItem) richTitle(theme Theme, width int) string {
    // Flag indicator
    var flagStr string
    flagStyle := lipgloss.NewStyle().Foreground(theme.Flagged)
    if i.email.IsFlagged() {
        flagStr = flagStyle.Render("★ ")
    } else {
        flagStr = "  "
    }

    // From
    from := i.email.From.Email
    if i.email.From.Name != "" {
        from = i.email.From.Name
    }
    from = truncate(from, 20)

    // Subject
    subject := i.email.Subject
    if subject == "" {
        subject = "(no subject)"
    }

    // Date
    age := formatAge(i.email.ReceivedAt)

    // Style based on read state
    textStyle := lipgloss.NewStyle().Foreground(theme.Unread).Bold(true)
    if i.email.IsRead() {
        textStyle = lipgloss.NewStyle().Foreground(theme.Read)
    }

    dateStyle := lipgloss.NewStyle().Foreground(theme.Read).Align(lipgloss.Right)

    fromRendered := textStyle.Width(22).Render(from)
    subjectRendered := textStyle.Render(truncate(subject, width-50))
    dateRendered := dateStyle.Render(age)

    return flagStr + fromRendered + " " + subjectRendered + " " + dateRendered
}
```

**Step 4: Run tests**

Run: `go test -run TestEmailItem_RichTitle ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add rich email row renderer with state-based coloring
```

---

### Task 3.2: Create Stats Bar

**Files:**
- Create: `internal/tui/stats_bar.go`
- Test: `internal/tui/stats_bar_test.go`

**Step 1: Write failing test**

```go
// internal/tui/stats_bar_test.go
package tui

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestStatsBar_Render(t *testing.T) {
    sb := newStatsBarModel(DetectTheme())
    sb.mailboxName = "Inbox"
    sb.unreadCount = 42
    sb.flaggedCount = 7
    sb.todayCount = 3

    result := sb.view(120)
    assert.Contains(t, result, "42")
    assert.Contains(t, result, "7")
    assert.Contains(t, result, "3")
}

func TestStatsBar_QuotaDisplay(t *testing.T) {
    sb := newStatsBarModel(DetectTheme())
    sb.quota = &fastmail.QuotaInfo{
        Used:        5 * 1024 * 1024 * 1024, // 5 GB
        Limit:       10 * 1024 * 1024 * 1024, // 10 GB
        UsedPercent: 50.0,
    }

    result := sb.view(120)
    assert.Contains(t, result, "50")
}

func TestStatsBar_UpdateFromMailboxes(t *testing.T) {
    sb := newStatsBarModel(DetectTheme())
    mailboxes := []fastmail.Mailbox{
        {Name: "Inbox", UnreadEmails: 10},
        {Name: "Work", UnreadEmails: 5},
    }

    sb.updateFromMailboxes(mailboxes)
    assert.Equal(t, uint64(15), sb.unreadCount)
}

func TestStatsBar_QuotaColor_Low(t *testing.T) {
    sb := newStatsBarModel(DarkTheme())
    color := sb.quotaColor(30.0)
    assert.Equal(t, sb.theme.QuotaLow, color)
}

func TestStatsBar_QuotaColor_Med(t *testing.T) {
    sb := newStatsBarModel(DarkTheme())
    color := sb.quotaColor(75.0)
    assert.Equal(t, sb.theme.QuotaMed, color)
}

func TestStatsBar_QuotaColor_High(t *testing.T) {
    sb := newStatsBarModel(DarkTheme())
    color := sb.quotaColor(92.0)
    assert.Equal(t, sb.theme.QuotaHigh, color)
}
```

**Step 2: Run tests**

Run: `go test -run TestStatsBar ./internal/tui/`
Expected: FAIL — `newStatsBarModel` undefined

**Step 3: Implement stats bar**

```go
// internal/tui/stats_bar.go
package tui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

type statsBarModel struct {
    theme        Theme
    mailboxName  string
    unreadCount  uint64
    flaggedCount uint64
    todayCount   uint64
    quota        *fastmail.QuotaInfo
}

func newStatsBarModel(theme Theme) statsBarModel {
    return statsBarModel{theme: theme}
}

func (sb *statsBarModel) updateFromMailboxes(mailboxes []fastmail.Mailbox) {
    var total uint64
    for _, mb := range mailboxes {
        total += mb.UnreadEmails
    }
    sb.unreadCount = total
}

func (sb statsBarModel) quotaColor(pct float64) lipgloss.Color {
    switch {
    case pct >= 90:
        return sb.theme.QuotaHigh
    case pct >= 70:
        return sb.theme.QuotaMed
    default:
        return sb.theme.QuotaLow
    }
}

func (sb statsBarModel) view(width int) string {
    barStyle := lipgloss.NewStyle().
        Background(sb.theme.StatusBarBg).
        Foreground(sb.theme.StatusBarFg).
        Width(width).
        Padding(0, 1)

    labelStyle := lipgloss.NewStyle().Foreground(sb.theme.KeyBarDesc)
    valueStyle := lipgloss.NewStyle().Foreground(sb.theme.StatValue).Bold(true)

    var parts []string

    if sb.mailboxName != "" {
        parts = append(parts, valueStyle.Render(sb.mailboxName))
    }

    parts = append(parts,
        labelStyle.Render("unread ")+valueStyle.Render(fmt.Sprintf("%d", sb.unreadCount)),
        labelStyle.Render("flagged ")+valueStyle.Render(fmt.Sprintf("%d", sb.flaggedCount)),
        labelStyle.Render("today ")+valueStyle.Render(fmt.Sprintf("%d", sb.todayCount)),
    )

    if sb.quota != nil {
        qColor := sb.quotaColor(sb.quota.UsedPercent)
        qStyle := lipgloss.NewStyle().Foreground(qColor).Bold(true)
        parts = append(parts,
            labelStyle.Render("quota ")+qStyle.Render(fmt.Sprintf("%.0f%%", sb.quota.UsedPercent)),
        )
    }

    content := strings.Join(parts, labelStyle.Render("  │  "))
    return barStyle.Render(content)
}
```

**Step 4: Run tests**

Run: `go test -run TestStatsBar ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add stats bar with unread/flagged/quota display
```

---

### Task 3.3: Wire Stats Bar + Collapsible Sidebar into Layout

**Files:**
- Modify: `internal/tui/tui.go`
- Modify: `internal/tui/tui_test.go`

**Step 1: Write failing test**

```go
func TestView_StatsBar_Visible(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.statsBar.unreadCount = 42
    m.statsBar.flaggedCount = 5

    v := m.View()
    assert.Contains(t, v, "42")
    assert.Contains(t, v, "5")
}

func TestUpdate_MailboxesLoaded_UpdatesStats(t *testing.T) {
    m := New(nil)
    m.connecting = false
    mailboxes := []fastmail.Mailbox{
        {ID: "1", Name: "Inbox", UnreadEmails: 10},
        {ID: "2", Name: "Work", UnreadEmails: 5},
    }

    m, _ = applyUpdate(m, mailboxesLoadedMsg{mailboxes: mailboxes})

    assert.Equal(t, uint64(15), m.statsBar.unreadCount)
}
```

**Step 2: Run tests**

Run: `go test -run TestView_StatsBar ./internal/tui/`
Expected: FAIL — `m.statsBar` undefined

**Step 3: Implement**

- Add `statsBar statsBarModel` to `Model`
- Initialize in `New()`
- On `mailboxesLoadedMsg`, call `m.statsBar.updateFromMailboxes()`
- Add quota fetch on connect: `fetchQuotaCmd()` → `quotaLoadedMsg` → `m.statsBar.quota`
- Render stats bar at the top of the layout in `View()`

**Step 4: Run all tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): wire stats bar and quota fetch into dashboard layout
```

---

### Task 3.4: Alternating Row Shading

**Files:**
- Modify: `internal/tui/email_list.go`
- Test: `internal/tui/email_list_test.go`

**Step 1: Write failing test**

```go
func TestEmailList_AlternatingRows(t *testing.T) {
    theme := DarkTheme()
    item1 := emailItem{email: fastmail.Email{Subject: "First"}}
    item2 := emailItem{email: fastmail.Email{Subject: "Second"}}

    row1 := item1.richTitle(theme, 80)
    row2 := item2.richTitle(theme, 80)

    // Both should render without error
    assert.Contains(t, row1, "First")
    assert.Contains(t, row2, "Second")
}
```

**Step 2: Implement alternating background**

Use a custom `list.ItemDelegate` that applies a subtle background shade to even-numbered rows. The `Theme.Selected` color for the cursor row, and a slightly different shade for alternating.

**Step 3: Run tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 4: Commit**

```
feat(tui): add alternating row shading to email list
```

---

### Task 3.5: Phase 3 BDD Integration Test

**Files:**
- Create: `internal/tui/bdd_stats_test.go`

**Step 1: Write BDD test**

```go
// internal/tui/bdd_stats_test.go
package tui

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// BDD: As a user, I see unread/flagged counts in the stats bar.
func TestBDD_Stats_ShowsCounts(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40

    mailboxes := []fastmail.Mailbox{
        {ID: "1", Name: "Inbox", UnreadEmails: 42},
        {ID: "2", Name: "Work", UnreadEmails: 8},
    }
    m, _ = applyUpdate(m, mailboxesLoadedMsg{mailboxes: mailboxes})

    v := m.View()
    assert.Contains(t, v, "50") // 42 + 8 unread
}

// BDD: As a user, I see quota percentage with color coding.
func TestBDD_Stats_ShowsQuota(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.statsBar.quota = &fastmail.QuotaInfo{UsedPercent: 73.5}

    v := m.View()
    assert.Contains(t, v, "73")
}

// BDD: As a user, I can collapse the sidebar with 'b'.
func TestBDD_Stats_CollapseSidebar(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    require.True(t, m.panes.sidebar)

    m, _ = applyUpdate(m, keyMsg("b"))

    assert.False(t, m.panes.sidebar)
    // Main area should now have full width
    layout := m.panes.computeLayout(120, 40)
    assert.Equal(t, 120, layout.mainWidth)
}

func keyMsg(s string) tea.KeyMsg {
    return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
```

**Step 2: Run tests**

Run: `go test -run TestBDD_Stats ./internal/tui/`
Expected: PASS

**Step 3: Commit**

```
test(tui): add BDD integration tests for stats bar (Phase 3)
```

---

## Phase 4: Polish + Edge Cases

### Task 4.1: Modal Overlays for Compose and Help

**Files:**
- Modify: `internal/tui/help_overlay.go`
- Modify: `internal/tui/compose.go`
- Modify: `internal/tui/tui.go`

**Step 1: Write failing test**

```go
func TestView_ComposeOverlay_RendersOnTopOfLayout(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.view = viewCompose
    cm := newComposeModel()
    cm.setSize(120, 40)
    m.composeView = &cm

    v := m.View()
    // Should still see some dashboard elements underneath
    // but compose should be visible
    assert.Contains(t, v, "To:")
}

func TestView_HelpOverlay_RendersOnTopOfLayout(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.helpOverlay = true

    v := m.View()
    assert.Contains(t, v, "help")
}
```

**Step 2: Implement modal overlay rendering**

Instead of replacing the layout, render compose/help as a centered lipgloss box on top of the layout content. Use `lipgloss.Place()` to center the modal within the full terminal dimensions.

**Step 3: Run tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 4: Commit**

```
feat(tui): render compose and help as modal overlays on layout
```

---

### Task 4.2: Thread View Integration with Preview

**Files:**
- Modify: `internal/tui/thread_view.go`
- Modify: `internal/tui/tui.go`

**Step 1: Write failing test**

```go
func TestUpdate_ThreadView_RendersInPreviewPane(t *testing.T) {
    // Thread view should render within the preview pane area
    // when triggered from the dashboard, not as fullscreen
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.panes.focus = PanePreview

    email := fastmail.Email{ID: "e1", ThreadID: "t1"}
    er := newEmailReaderModel(email)
    er.setSize(80, 20)
    er.showThread = true
    m.emailReader = &er

    // Thread view should be set up within the layout
    m, _ = applyUpdate(m, tea.Msg(nil))
    // Thread view should exist
    // (exact behavior TBD based on implementation)
}
```

**Step 2: Implement**

Thread view renders in the preview pane area when launched from the dashboard. If launched from fullscreen reader, it stays fullscreen.

**Step 3: Run tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 4: Commit**

```
feat(tui): integrate thread view with dashboard preview pane
```

---

### Task 4.3: Responsive Terminal Handling

**Files:**
- Modify: `internal/tui/layout.go`
- Modify: `internal/tui/layout_test.go`

**Step 1: Write failing test**

```go
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

    // Below 40 rows, preview should be hidden (0 height)
    // and email list gets full content height
    assert.Greater(t, layout.listHeight, 0)
}

func TestPaneManager_ComputeLayout_VerySmall(t *testing.T) {
    pm := newPaneManager()

    layout := pm.computeLayout(40, 15)

    // Should not panic, should still produce valid layout
    assert.GreaterOrEqual(t, layout.listHeight, 0)
    assert.GreaterOrEqual(t, layout.previewHeight, 0)
}
```

**Step 2: Run tests**

Run: `go test -run TestPaneManager_ComputeLayout_Narrow ./internal/tui/`
Expected: FAIL — no responsive logic yet

**Step 3: Implement responsive breakpoints**

In `computeLayout()`:
- Width < 80: auto-collapse sidebar regardless of `pm.sidebar`
- Height < 25: hide preview pane (list gets full height)
- Ensure no negative dimensions

**Step 4: Run tests**

Run: `go test -race ./internal/tui/`
Expected: PASS

**Step 5: Commit**

```
feat(tui): add responsive terminal size handling for narrow/short terminals
```

---

### Task 4.4: Attachment Indicator in Email List

**Files:**
- Modify: `internal/tui/email_list.go`
- Test: `internal/tui/email_list_test.go`

**Step 1: Write failing test**

```go
func TestEmailItem_RichTitle_WithAttachment(t *testing.T) {
    theme := DarkTheme()
    item := emailItem{email: fastmail.Email{
        Subject:     "With file",
        From:        fastmail.EmailAddress{Name: "Dave"},
        Attachments: []fastmail.Attachment{{Name: "report.pdf"}},
    }}

    title := item.richTitle(theme, 80)
    // Should contain an attachment indicator
    assert.Contains(t, title, "📎")
}
```

**Step 2: Implement**

Add `📎` indicator in `richTitle()` when `len(i.email.Attachments) > 0`.

**Step 3: Run tests**

Run: `go test -run TestEmailItem_RichTitle_WithAttachment ./internal/tui/`
Expected: PASS

**Step 4: Commit**

```
feat(tui): add attachment indicator to email list rows
```

---

### Task 4.5: Phase 4 BDD Integration Test

**Files:**
- Create: `internal/tui/bdd_polish_test.go`

**Step 1: Write BDD test**

```go
// internal/tui/bdd_polish_test.go
package tui

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
)

// BDD: As a user on a narrow terminal, the sidebar auto-collapses.
func TestBDD_Polish_NarrowTerminalCollapsesSidebar(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.panes.sidebar = true

    m, _ = applyUpdate(m, tea.WindowSizeMsg{Width: 60, Height: 30})

    layout := m.panes.computeLayout(60, 30)
    assert.Equal(t, 0, layout.sidebarWidth)
}

// BDD: As a user, compose renders as a modal overlay.
func TestBDD_Polish_ComposeIsOverlay(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    cm := newComposeModel()
    cm.setSize(100, 30)
    m.composeView = &cm
    m.view = viewCompose

    v := m.View()
    assert.Contains(t, v, "To:")
}

// BDD: As a user, help renders as a modal overlay.
func TestBDD_Polish_HelpIsOverlay(t *testing.T) {
    m := New(nil)
    m.connecting = false
    m.width = 120
    m.height = 40
    m.helpOverlay = true

    v := m.View()
    assert.Contains(t, v, "help")
}

// BDD: As a user, emails with attachments show an indicator.
func TestBDD_Polish_AttachmentIndicator(t *testing.T) {
    theme := DarkTheme()
    item := emailItem{email: fastmail.Email{
        Subject:     "Report",
        From:        fastmail.EmailAddress{Name: "Boss"},
        Attachments: []fastmail.Attachment{{Name: "q4.pdf"}},
    }}

    title := item.richTitle(theme, 80)
    assert.Contains(t, title, "📎")
}
```

**Step 2: Run tests**

Run: `go test -run TestBDD_Polish ./internal/tui/`
Expected: PASS

**Step 3: Commit**

```
test(tui): add BDD integration tests for polish (Phase 4)
```

---

## Estimated File Changes Summary

| File | Action | Phase |
|------|--------|-------|
| `internal/tui/theme.go` | Create | 1 |
| `internal/tui/theme_test.go` | Create | 1 |
| `internal/tui/layout.go` | Create | 1, 4 |
| `internal/tui/layout_test.go` | Create | 1, 4 |
| `internal/tui/key_bar.go` | Create | 1 |
| `internal/tui/key_bar_test.go` | Create | 1 |
| `internal/tui/split.go` | Create | 2 |
| `internal/tui/split_test.go` | Create | 2 |
| `internal/tui/stats_bar.go` | Create | 3 |
| `internal/tui/stats_bar_test.go` | Create | 3 |
| `internal/tui/tui.go` | Modify | 1, 2, 3, 4 |
| `internal/tui/tui_test.go` | Modify | 1, 2, 3 |
| `internal/tui/mailbox_list.go` | Modify | 1 |
| `internal/tui/email_list.go` | Modify | 1, 3, 4 |
| `internal/tui/email_list_test.go` | Modify | 1, 3, 4 |
| `internal/tui/email_reader.go` | Modify | 2 |
| `internal/tui/help_overlay.go` | Modify | 4 |
| `internal/tui/compose.go` | Modify | 4 |
| `internal/tui/thread_view.go` | Modify | 4 |
| `internal/tui/bdd_layout_test.go` | Create | 1 |
| `internal/tui/bdd_preview_test.go` | Create | 2 |
| `internal/tui/bdd_stats_test.go` | Create | 3 |
| `internal/tui/bdd_polish_test.go` | Create | 4 |
