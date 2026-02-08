package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVacationGetBuilder(t *testing.T) {
	args := NewVacationGet("account-1").Build()
	assert.Equal(t, "account-1", args["accountId"])
}

func TestVacationSetBuilder_Update(t *testing.T) {
	args := NewVacationSet("account-1").
		Update(map[string]any{
			"isEnabled": true,
			"subject":   "Out of Office",
			"textBody":  "I am on vacation",
		}).Build()

	assert.Equal(t, "account-1", args["accountId"])
	update := args["update"].(map[string]map[string]any)
	assert.True(t, update["singleton"]["isEnabled"].(bool))
	assert.Equal(t, "Out of Office", update["singleton"]["subject"])
}

func TestVacationGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "v-state-1",
		"list": [{
			"id": "singleton",
			"isEnabled": true,
			"fromDate": "2026-01-01T00:00:00Z",
			"toDate": "2026-01-15T00:00:00Z",
			"subject": "Out of Office",
			"textBody": "I am away"
		}],
		"notFound": []
	}`

	var resp VacationGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.Len(t, resp.List, 1)
	assert.True(t, resp.List[0].IsEnabled)
	assert.Equal(t, "Out of Office", resp.List[0].Subject)
	assert.Equal(t, "I am away", resp.List[0].TextBody)
}

func TestVacationSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "v1",
		"newState": "v2",
		"updated": {"singleton": null},
		"notUpdated": {}
	}`

	var resp VacationSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "v2", resp.NewState)
	assert.Contains(t, resp.Updated, "singleton")
}
