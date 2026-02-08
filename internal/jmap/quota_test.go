package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuotaGetArgs_MinimalBuild(t *testing.T) {
	args := NewQuotaGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	_, hasIDs := args["ids"]
	assert.False(t, hasIDs, "ids should not be present when none are set")
}

func TestQuotaGetArgs_WithIDs(t *testing.T) {
	args := NewQuotaGet("account-1").
		IDs("quota-1", "quota-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"quota-1", "quota-2"}, args["ids"])
}

func TestQuotaGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"state": "q1",
		"list": [
			{
				"id": "quota-1",
				"resourceType": "octets",
				"used": 2254857830,
				"hardLimit": 32212254720,
				"scope": "account",
				"name": "Mail storage"
			}
		],
		"notFound": []
	}`

	var resp QuotaGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "acc1", resp.AccountID)
	assert.Equal(t, "q1", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "quota-1", resp.List[0].ID)
	assert.Equal(t, "octets", resp.List[0].ResourceType)
	assert.Equal(t, uint64(2254857830), resp.List[0].Used)
	assert.Equal(t, uint64(32212254720), resp.List[0].HardLimit)
	assert.Equal(t, "account", resp.List[0].Scope)
	assert.Equal(t, "Mail storage", resp.List[0].Name)
	assert.Empty(t, resp.NotFound)
}

func TestQuotaGetResponse_DecodeWithOptionalFields(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"state": "q2",
		"list": [
			{
				"id": "quota-full",
				"resourceType": "octets",
				"used": 1000000,
				"hardLimit": 5000000,
				"scope": "account",
				"name": "Storage",
				"types": ["Mail", "Calendar"],
				"warnLimit": 4000000,
				"softLimit": 4500000,
				"description": "Total storage quota"
			}
		],
		"notFound": ["quota-missing"]
	}`

	var resp QuotaGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "acc1", resp.AccountID)
	require.Len(t, resp.List, 1)

	q := resp.List[0]
	assert.Equal(t, "quota-full", q.ID)
	assert.Equal(t, []string{"Mail", "Calendar"}, q.Types)
	assert.Equal(t, uint64(4000000), q.WarnLimit)
	assert.Equal(t, uint64(4500000), q.SoftLimit)
	assert.Equal(t, "Total storage quota", q.Description)
	assert.Equal(t, []string{"quota-missing"}, resp.NotFound)
}

func TestQuotaGetResponse_DecodeMultipleQuotas(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"state": "q3",
		"list": [
			{
				"id": "quota-octets",
				"resourceType": "octets",
				"used": 1000,
				"hardLimit": 5000,
				"scope": "account",
				"name": "Bytes quota"
			},
			{
				"id": "quota-count",
				"resourceType": "count",
				"used": 100,
				"hardLimit": 500,
				"scope": "account",
				"name": "Message count quota"
			}
		],
		"notFound": []
	}`

	var resp QuotaGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Len(t, resp.List, 2)
	assert.Equal(t, "octets", resp.List[0].ResourceType)
	assert.Equal(t, "count", resp.List[1].ResourceType)
}
