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
