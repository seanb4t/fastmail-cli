package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThreadGetBuilder(t *testing.T) {
	args := NewThreadGet("account-1").Build()
	assert.Equal(t, "account-1", args["accountId"])
	assert.Nil(t, args["ids"])
}

func TestThreadGetBuilder_WithIDs(t *testing.T) {
	args := NewThreadGet("account-1").IDs("t1", "t2").Build()
	assert.Equal(t, []string{"t1", "t2"}, args["ids"])
}

func TestThreadGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "t-state-1",
		"list": [{
			"id": "thread-1",
			"emailIds": ["em1", "em2", "em3"]
		}],
		"notFound": []
	}`

	var resp ThreadGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "thread-1", resp.List[0].ID)
	assert.Equal(t, []string{"em1", "em2", "em3"}, resp.List[0].EmailIDs)
}
