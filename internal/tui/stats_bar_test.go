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
		Used:        5 * 1024 * 1024 * 1024,  // 5 GB
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

func TestStatsBar_ShowsBrand(t *testing.T) {
	theme := DarkTheme()
	sb := newStatsBarModel(theme)
	v := sb.view(120)
	assert.Contains(t, v, "Fastmail CLI")
}

func TestStatsBar_NarrowHidesBrand(t *testing.T) {
	theme := DarkTheme()
	sb := newStatsBarModel(theme)
	v := sb.view(30)
	assert.NotContains(t, v, "Fastmail CLI")
}
