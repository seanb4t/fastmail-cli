# TUI Redesign: Rich Dashboard Layout

**Bead:** fastmail-cli-8ul
**Date:** 2026-02-09
**Status:** Approved

## Summary

Redesign the TUI from a fullscreen view-swapping model to a persistent pane-based dashboard with a collapsible mailbox sidebar, adjustable email list/preview split, rich stats bar, contextual keybinding footer, and auto-detecting dark/light color theme.

## Layout

```
┌──────────────────────────────────────────────────────────────┐
│ ▌INBOX 142 unread │ ★ 23 flagged │ 📧 today 8 │ ▐ 73% quota │  ← Stats Bar
├────────────┬─────────────────────────────────────────────────┤
│ Inbox  (142)│ ★  alice@ex… │ Re: Project update  │ 10:23 AM  │
│ Drafts   (3)│    bob@co…  │ Invoice #4521       │  9:15 AM  │  ← Email List
│ Sent       │    team@…   │ Weekly standup notes │  Yesterday│
│ Archive    │ ★  ceo@…    │ Q4 Planning          │  Yesterday│
│ Trash      ├─────────────────────────────────────────────────┤
│ ────────── │ From: alice@example.com                         │
│ Work       │ To: sean@fastmail.com                           │
│ Personal   │ Date: Feb 8, 2026 10:23 AM                     │  ← Preview Pane
│ Receipts   │                                                 │
│            │ Hi Sean,                                        │
│            │ Just following up on the project timeline...    │
├────────────┴─────────────────────────────────────────────────┤
│ ↹ pane  / search  c compose  a archive  f flag  ? help      │  ← Keybinding Bar
└──────────────────────────────────────────────────────────────┘
```

### Four Zones

1. **Stats bar** (1-2 lines) — Unread count, flagged count, today's email count, storage quota with mini bar. Refreshed every 60 seconds via tick.
2. **Mailbox sidebar** (left, collapsible with `b`) — Bubbletea list with styled unread counts per mailbox. Default width ~15 chars, collapses to 0.
3. **Main area** (right, horizontal split) — Email list on top, preview pane on bottom. Divider position adjustable with `+`/`-` keys (percentage-based, 30-80% range).
4. **Keybinding bar** (1 line) — Context-sensitive, updates based on which pane has focus.

### Focus Model

`Tab` cycles focus: mailbox sidebar → email list → preview pane. Focused pane gets a highlighted border color and bold title. Keys dispatch to the focused pane only, except globals (`q`, `?`, `ctrl+c`).

## Theme System

### Auto-Detection

On startup, `lipgloss.HasDarkBackground()` selects the active palette. All styling references the palette through a `Theme` struct — no hardcoded colors.

### Palette Structure

```go
type Theme struct {
    // Chrome
    StatusBarBg    lipgloss.Color
    StatusBarFg    lipgloss.Color
    PaneBorder     lipgloss.Color  // unfocused pane border
    FocusBorder    lipgloss.Color  // focused pane border (bright accent)
    KeyBarKey      lipgloss.Color  // keybinding letter (bright)
    KeyBarDesc     lipgloss.Color  // keybinding description (muted)

    // Email states
    Unread         lipgloss.Color  // bold/bright for unread rows
    Read           lipgloss.Color  // dimmed for read rows
    Flagged        lipgloss.Color  // star/flag indicator
    Selected       lipgloss.Color  // cursor row highlight bg

    // Content
    HeaderLabel    lipgloss.Color  // "From:", "To:", "Date:" labels
    HeaderValue    lipgloss.Color  // header values
    QuotedText     lipgloss.Color  // > quoted reply text
    Link           lipgloss.Color  // URLs in body

    // Stats
    StatValue      lipgloss.Color  // numbers in stats bar
    QuotaLow       lipgloss.Color  // green
    QuotaMed       lipgloss.Color  // yellow
    QuotaHigh      lipgloss.Color  // red
}
```

### Color Palettes

- **Dark** — Catppuccin Mocha-inspired: deep base (`#1e1e2e`), lavender accent for focus, peach for flagged, green/yellow/red for quota. Unread in white bold, read in overlay gray.
- **Light** — Catppuccin Latte-inspired: off-white base, blue accent for focus, orange for flagged. Unread in dark bold, read in muted gray.

### Visual Details

- All panes wrapped in lipgloss rounded borders. Focused pane gets `FocusBorder` color + bold title.
- Email list rows: flag indicator (colored dot) + sender (truncated, bold if unread) + subject + relative date (right-aligned). Alternating row backgrounds.

## Components

### New Files

| File | Purpose |
|------|---------|
| `theme.go` | `Theme` struct, dark/light palettes, `DetectTheme()` |
| `layout.go` | Pane manager: focus cycling, border rendering, resize dispatch |
| `stats_bar.go` | Stats bar model: fetches unread/flagged/quota, renders stat groups |
| `key_bar.go` | Context-sensitive keybinding footer, updates on focus change |
| `split.go` | Adjustable horizontal splitter between email list and preview |

### Heavy Refactors

| File | Changes |
|------|---------|
| `tui.go` | Root model becomes a layout coordinator, not a view switcher. Holds pane manager, stats bar, key bar. `Update()` routes to focused pane. |
| `mailbox_list.go` | Strip fullscreen logic, become pane-aware sub-model. Accept constrained width/height. Styled unread counts per row. |
| `email_list.go` | Replace `list.Model` rendering with custom lipgloss row renderer for colored states (unread/flagged/read). Columnar layout: flag + sender + subject + date. |
| `email_reader.go` | Becomes the preview pane. Renders inline instead of fullscreen. Header block (From/To/Date) + viewport body. Colored quoted text. |

### Minor Changes

| File | Changes |
|------|---------|
| `help_overlay.go` | Renders as centered modal over the layout |
| `compose.go` | Renders as modal overlay on top of the layout |
| `email_actions.go` | No change — action logic is independent of rendering |
| `status.go` | Merge into `stats_bar.go` or keep for transient flash messages |

### Pane Manager

```go
type PaneManager struct {
    focus    PaneID           // which pane has focus
    panes    []Pane           // ordered for tab cycling
    sidebar  bool             // sidebar visible?
    splitPct int              // email list / preview split (30-80%)
}
```

`Tab` advances focus. `b` toggles sidebar. `+`/`-` adjusts `splitPct` by 10%. On `WindowSizeMsg`, the manager recalculates each pane's dimensions and calls `SetSize()` on sub-models.

### Message Flow

Root `Update()` checks globals first, then routes to pane manager for focus/resize keys, then delegates to the focused pane's `Update()`. Stats bar refreshes on a `tickMsg` every 60 seconds.

## Implementation Phases

### Phase 1 — Theme + Layout Shell

Foundation work. TUI remains functional throughout.

- Create `theme.go` with dark/light palettes and auto-detection
- Create `layout.go` pane manager with focus cycling (`Tab`)
- Create `key_bar.go` with static keybinding footer
- Refactor `tui.go` from view-switcher to layout coordinator
- Render mailbox list and email list as side-by-side panes with borders
- No preview pane yet — selecting an email still opens fullscreen reader

**Result:** Colored borders, focus indicators, keybinding bar, theme system in place.

### Phase 2 — Inline Preview + Split

The big payoff — dashboard feel comes alive.

- Create `split.go` adjustable horizontal splitter
- Refactor `email_reader.go` to render as constrained preview pane
- Selecting an email populates the preview instead of switching views
- `+`/`-` adjusts split ratio, `Enter` opens fullscreen reader for long emails
- Colored quoted text (`>` lines) and header labels in preview

**Result:** Scan and read without leaving the view.

### Phase 3 — Rich Email List + Stats Bar

Visual polish to match the GitVision inspiration.

- Custom row renderer for email list: flag dot, bold unread, truncated sender, subject, relative date
- Alternating row shading
- Create `stats_bar.go` with unread/flagged/today counts and quota bar
- Periodic refresh tick for stats
- Collapsible sidebar with `b` key

**Result:** Rich, informative, colored dashboard.

### Phase 4 — Polish + Edge Cases

Production-ready finishing touches.

- Compose and help as modal overlays on the layout
- Thread view integration with preview pane
- Graceful handling of small terminals (collapse sidebar below 80 cols, hide preview below 40 rows)
- Attachment indicators in email list rows
- Search mode visual treatment (dimmed non-matching rows or filter)

**Result:** Handles all edge cases, fully polished.

## Keybindings

All existing vim-style keybindings are preserved. New keys added:

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus between panes |
| `b` | Toggle mailbox sidebar |
| `+` / `-` | Adjust email list / preview split ratio |
| `Enter` (in email list) | Open fullscreen reader (preview already shows content) |

Context-sensitive keybinding bar shows relevant keys for the focused pane.
