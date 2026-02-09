package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConditionToJMAP_From(t *testing.T) {
	f := NewConditionFilter(&FilterCondition{From: "alice@example.com"})
	m := f.ToJMAP()
	assert.Equal(t, "alice@example.com", m["from"])
	assert.NotContains(t, m, "to")
}

func TestConditionToJMAP_AllFields(t *testing.T) {
	before := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	hasAtt := true

	f := NewConditionFilter(&FilterCondition{
		InMailbox:      "mb-1",
		InMailboxOther: []string{"mb-2"},
		From:           "alice",
		To:             "bob",
		Cc:             "carol",
		Bcc:            "dave",
		Subject:        "meeting",
		Body:           "agenda",
		Text:           "important",
		Before:         &before,
		After:          &after,
		MinSize:        1024,
		MaxSize:        1048576,
		HasAttachment:  &hasAtt,
		HasKeyword:     "$flagged",
		NotKeyword:     "$seen",
		Header:         [2]string{"X-Custom", "value"},
	})

	m := f.ToJMAP()
	assert.Equal(t, "mb-1", m["inMailbox"])
	assert.Equal(t, []string{"mb-2"}, m["inMailboxOtherThan"])
	assert.Equal(t, "alice", m["from"])
	assert.Equal(t, "bob", m["to"])
	assert.Equal(t, "carol", m["cc"])
	assert.Equal(t, "dave", m["bcc"])
	assert.Equal(t, "meeting", m["subject"])
	assert.Equal(t, "agenda", m["body"])
	assert.Equal(t, "important", m["text"])
	assert.Equal(t, "2024-06-01T00:00:00Z", m["before"])
	assert.Equal(t, "2024-01-01T00:00:00Z", m["after"])
	assert.Equal(t, uint64(1024), m["minSize"])
	assert.Equal(t, uint64(1048576), m["maxSize"])
	assert.Equal(t, true, m["hasAttachment"])
	assert.Equal(t, "$flagged", m["hasKeyword"])
	assert.Equal(t, "$seen", m["notKeyword"])
	assert.Equal(t, []string{"X-Custom", "value"}, m["header"])
}

func TestConditionToJMAP_Empty(t *testing.T) {
	f := NewConditionFilter(&FilterCondition{})
	m := f.ToJMAP()
	assert.Empty(t, m)
}

func TestCompoundToJMAP_AND(t *testing.T) {
	f := NewCompoundFilter(OpAND,
		NewConditionFilter(&FilterCondition{From: "alice"}),
		NewConditionFilter(&FilterCondition{Subject: "meeting"}),
	)
	m := f.ToJMAP()

	assert.Equal(t, "AND", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	assert.Len(t, conditions, 2)
	assert.Equal(t, "alice", conditions[0]["from"])
	assert.Equal(t, "meeting", conditions[1]["subject"])
}

func TestCompoundToJMAP_NOT(t *testing.T) {
	f := NewCompoundFilter(OpNOT,
		NewConditionFilter(&FilterCondition{From: "spam@example.com"}),
	)
	m := f.ToJMAP()

	assert.Equal(t, "NOT", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "spam@example.com", conditions[0]["from"])
}

func TestCompoundToJMAP_Nested(t *testing.T) {
	f := NewCompoundFilter(OpAND,
		NewConditionFilter(&FilterCondition{From: "alice"}),
		NewCompoundFilter(OpOR,
			NewConditionFilter(&FilterCondition{Subject: "meeting"}),
			NewConditionFilter(&FilterCondition{Subject: "standup"}),
		),
	)
	m := f.ToJMAP()

	assert.Equal(t, "AND", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	assert.Len(t, conditions, 2)

	nested := conditions[1]
	assert.Equal(t, "OR", nested["operator"])
	nestedConds := nested["conditions"].([]map[string]any)
	assert.Len(t, nestedConds, 2)
}

func TestEmptyFilter(t *testing.T) {
	f := Filter{}
	m := f.ToJMAP()
	assert.Empty(t, m)
}

func TestHasAttachmentFalse(t *testing.T) {
	hasAtt := false
	f := NewConditionFilter(&FilterCondition{HasAttachment: &hasAtt})
	m := f.ToJMAP()
	assert.Equal(t, false, m["hasAttachment"])
}
