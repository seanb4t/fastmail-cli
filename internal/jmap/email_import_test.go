package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailImportBuilder_Build(t *testing.T) {
	args := NewEmailImport("account-1").
		Import("imp1", "blob-123",
			map[string]bool{"mb-inbox": true},
			map[string]bool{"$seen": true},
		).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	emails, ok := args["emails"].(map[string]map[string]any)
	require.True(t, ok, "emails should be a map")
	require.Contains(t, emails, "imp1")

	entry := emails["imp1"]
	assert.Equal(t, "blob-123", entry["blobId"])
	assert.Equal(t, map[string]bool{"mb-inbox": true}, entry["mailboxIds"])
	assert.Equal(t, map[string]bool{"$seen": true}, entry["keywords"])
}

func TestEmailImportBuilder_NoKeywords(t *testing.T) {
	args := NewEmailImport("account-1").
		Import("imp1", "blob-456",
			map[string]bool{"mb-inbox": true},
			nil,
		).
		Build()

	emails := args["emails"].(map[string]map[string]any)
	entry := emails["imp1"]

	_, hasKeywords := entry["keywords"]
	assert.False(t, hasKeywords, "keywords should not be present when nil")
}

func TestEmailImportBuilder_MinimalBuild(t *testing.T) {
	args := NewEmailImport("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	_, hasEmails := args["emails"]
	assert.False(t, hasEmails, "emails should not be present when empty")
}

func TestEmailImportResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"created": {
			"imp1": {
				"id": "email-new-1",
				"blobId": "blob-result-1",
				"threadId": "thread-1",
				"size": 12345
			}
		}
	}`

	var resp EmailImportResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.Contains(t, resp.Created, "imp1")

	result := resp.Created["imp1"]
	assert.Equal(t, "email-new-1", result.ID)
	assert.Equal(t, "blob-result-1", result.BlobID)
	assert.Equal(t, "thread-1", result.ThreadID)
	assert.Equal(t, uint64(12345), result.Size)
}

func TestEmailImportResponse_DecodeNotCreated(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"notCreated": {
			"imp1": {
				"type": "invalidEmail",
				"description": "The message is not valid RFC 5322"
			}
		}
	}`

	var resp EmailImportResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.Contains(t, resp.NotCreated, "imp1")

	errInfo := resp.NotCreated["imp1"]
	assert.Equal(t, "invalidEmail", errInfo.Type)
	assert.Contains(t, errInfo.Description, "RFC 5322")
}
