package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_SimpleFrom(t *testing.T) {
	f := Parse("from:alice")
	m := f.ToJMAP()
	assert.Equal(t, "alice", m["from"])
}

func TestParse_QuotedSubject(t *testing.T) {
	f := Parse(`subject:"meeting notes"`)
	m := f.ToJMAP()
	assert.Equal(t, "meeting notes", m["subject"])
}

func TestParse_BeforeDate(t *testing.T) {
	f := Parse("before:2024-06-01")
	m := f.ToJMAP()
	assert.Equal(t, "2024-06-01T00:00:00Z", m["before"])
}

func TestParse_AfterDate(t *testing.T) {
	f := Parse("after:2024-01-15")
	m := f.ToJMAP()
	assert.Equal(t, "2024-01-15T00:00:00Z", m["after"])
}

func TestParse_HasAttachment(t *testing.T) {
	f := Parse("has:attachment")
	m := f.ToJMAP()
	assert.Equal(t, true, m["hasAttachment"])
}

func TestParse_IsUnread(t *testing.T) {
	f := Parse("is:unread")
	m := f.ToJMAP()
	assert.Equal(t, "$seen", m["notKeyword"])
}

func TestParse_IsRead(t *testing.T) {
	f := Parse("is:read")
	m := f.ToJMAP()
	assert.Equal(t, "$seen", m["hasKeyword"])
}

func TestParse_IsFlagged(t *testing.T) {
	f := Parse("is:flagged")
	m := f.ToJMAP()
	assert.Equal(t, "$flagged", m["hasKeyword"])
}

func TestParse_IsDraft(t *testing.T) {
	f := Parse("is:draft")
	m := f.ToJMAP()
	assert.Equal(t, "$draft", m["hasKeyword"])
}

func TestParse_FolderSent(t *testing.T) {
	f := Parse("in:Sent")
	m := f.ToJMAP()
	assert.Equal(t, "Sent", m["inMailbox"])
}

func TestParse_FolderAlias(t *testing.T) {
	f := Parse("folder:Archive")
	m := f.ToJMAP()
	assert.Equal(t, "Archive", m["inMailbox"])
}

func TestParse_Negation(t *testing.T) {
	f := Parse("-from:spam")
	m := f.ToJMAP()

	assert.Equal(t, "NOT", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	require.Len(t, conditions, 1)
	assert.Equal(t, "spam", conditions[0]["from"])
}

func TestParse_BareWords(t *testing.T) {
	f := Parse("meeting notes")
	m := f.ToJMAP()
	assert.Equal(t, "meeting notes", m["text"])
}

func TestParse_MultipleConditionsAND(t *testing.T) {
	f := Parse("from:alice subject:meeting")
	m := f.ToJMAP()

	assert.Equal(t, "AND", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	require.Len(t, conditions, 2)
	assert.Equal(t, "alice", conditions[0]["from"])
	assert.Equal(t, "meeting", conditions[1]["subject"])
}

func TestParse_OR(t *testing.T) {
	f := Parse("from:alice OR from:bob")
	m := f.ToJMAP()

	assert.Equal(t, "OR", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	require.Len(t, conditions, 2)
	assert.Equal(t, "alice", conditions[0]["from"])
	assert.Equal(t, "bob", conditions[1]["from"])
}

func TestParse_MixedBareAndFields(t *testing.T) {
	f := Parse("from:alice important update")
	m := f.ToJMAP()

	assert.Equal(t, "AND", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	require.Len(t, conditions, 2)
	assert.Equal(t, "alice", conditions[0]["from"])
	assert.Equal(t, "important update", conditions[1]["text"])
}

func TestParse_EmptyQuery(t *testing.T) {
	f := Parse("")
	m := f.ToJMAP()
	assert.Empty(t, m)
}

func TestParse_To(t *testing.T) {
	f := Parse("to:bob@example.com")
	m := f.ToJMAP()
	assert.Equal(t, "bob@example.com", m["to"])
}

func TestParse_Cc(t *testing.T) {
	f := Parse("cc:carol")
	m := f.ToJMAP()
	assert.Equal(t, "carol", m["cc"])
}

func TestParse_Body(t *testing.T) {
	f := Parse("body:agenda")
	m := f.ToJMAP()
	assert.Equal(t, "agenda", m["body"])
}

func TestParse_ComplexQuery(t *testing.T) {
	f := Parse(`from:alice subject:"project update" has:attachment before:2024-06-01`)
	m := f.ToJMAP()

	assert.Equal(t, "AND", m["operator"])
	conditions := m["conditions"].([]map[string]any)
	require.Len(t, conditions, 4)
	assert.Equal(t, "alice", conditions[0]["from"])
	assert.Equal(t, "project update", conditions[1]["subject"])
	assert.Equal(t, true, conditions[2]["hasAttachment"])
	assert.Equal(t, "2024-06-01T00:00:00Z", conditions[3]["before"])
}

func TestParse_InvalidDate(t *testing.T) {
	// Invalid date token is dropped; from:alice becomes the sole condition
	f := Parse("from:alice before:not-a-date")
	m := f.ToJMAP()

	// Only from:alice survives as a single condition (no AND wrapper)
	assert.Equal(t, "alice", m["from"])
	assert.NotContains(t, m, "before")
	assert.NotContains(t, m, "operator")
}

func TestParse_UnknownField(t *testing.T) {
	// Unknown fields: the token has field="xyzzy" which is not known,
	// so the full "xyzzy:value" is treated as bare text.
	// However, the tokenizer splits on colon and stores field+value separately,
	// so only "value" ends up as the bare word.
	f := Parse("xyzzy:value")
	m := f.ToJMAP()
	assert.Equal(t, "value", m["text"])
}

func TestParse_SingleCondition(t *testing.T) {
	// Single condition should NOT be wrapped in AND
	f := Parse("from:alice")
	assert.NotNil(t, f.Condition)
	assert.Nil(t, f.Compound)
}

func TestParse_QuotedBareText(t *testing.T) {
	f := Parse(`"exact phrase"`)
	m := f.ToJMAP()
	assert.Equal(t, "exact phrase", m["text"])
}
