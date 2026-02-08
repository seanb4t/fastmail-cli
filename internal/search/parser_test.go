package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_FreeText(t *testing.T) {
	filter := Parse("quarterly report")
	assert.Equal(t, map[string]any{"text": "quarterly report"}, filter)
}

func TestParse_FromToken(t *testing.T) {
	filter := Parse("from:alice")
	assert.Equal(t, map[string]any{"from": "alice"}, filter)
}

func TestParse_ToToken(t *testing.T) {
	filter := Parse("to:bob")
	assert.Equal(t, map[string]any{"to": "bob"}, filter)
}

func TestParse_SubjectToken(t *testing.T) {
	filter := Parse("subject:meeting")
	assert.Equal(t, map[string]any{"subject": "meeting"}, filter)
}

func TestParse_CompoundQuery(t *testing.T) {
	filter := Parse("from:alice subject:meeting is:unread")
	assert.Equal(t, map[string]any{
		"operator": "AND",
		"conditions": []map[string]any{
			{"from": "alice"},
			{"subject": "meeting"},
			{"notKeyword": "$seen"},
		},
	}, filter)
}

func TestParse_HasAttachment(t *testing.T) {
	filter := Parse("has:attachment")
	assert.Equal(t, map[string]any{"hasAttachment": true}, filter)
}

func TestParse_IsUnread(t *testing.T) {
	filter := Parse("is:unread")
	assert.Equal(t, map[string]any{"notKeyword": "$seen"}, filter)
}

func TestParse_IsFlagged(t *testing.T) {
	filter := Parse("is:flagged")
	assert.Equal(t, map[string]any{"hasKeyword": "$flagged"}, filter)
}

func TestParse_DateFilters(t *testing.T) {
	filter := Parse("after:2026-01-01 before:2026-02-01")
	assert.Equal(t, map[string]any{
		"operator": "AND",
		"conditions": []map[string]any{
			{"after": "2026-01-01"},
			{"before": "2026-02-01"},
		},
	}, filter)
}

func TestParse_InMailbox(t *testing.T) {
	filter := Parse("in:drafts")
	assert.Equal(t, map[string]any{"inMailbox": "drafts"}, filter)
}

func TestParse_MixedFreeTextAndTokens(t *testing.T) {
	filter := Parse("from:alice quarterly report")
	assert.Equal(t, map[string]any{
		"operator": "AND",
		"conditions": []map[string]any{
			{"from": "alice"},
			{"text": "quarterly report"},
		},
	}, filter)
}

func TestParse_EmptyQuery(t *testing.T) {
	filter := Parse("")
	assert.Equal(t, map[string]any{}, filter)
}
