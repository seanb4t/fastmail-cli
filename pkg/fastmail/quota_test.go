package fastmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sessionResponseWithQuota returns a JMAP session that includes the quota capability.
func sessionResponseWithQuota(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"urn:ietf:params:jmap:quota": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {},
					"urn:ietf:params:jmap:quota": {}
				}
			}
		},
		"primaryAccounts": {"urn:ietf:params:jmap:mail": "acc1"},
		"username": "test@example.com",
		"apiUrl": "` + apiURL + `",
		"downloadUrl": "https://example.com/download",
		"uploadUrl": "https://example.com/upload",
		"eventSourceUrl": "https://example.com/events",
		"state": "s1"
	}`
}

func TestQuotaService_Get(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		assert.Equal(t, "Quota/get", methodName)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Quota/get", {
					"accountId": "acc1",
					"state": "q1",
					"list": [{
						"id": "quota-1",
						"resourceType": "octets",
						"used": 2254857830,
						"hardLimit": 32212254720,
						"scope": "account",
						"name": "Mail storage"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithQuota(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	quota, err := client.Quota().Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, quota)

	assert.Equal(t, uint64(2254857830), quota.Used)
	assert.Equal(t, uint64(32212254720), quota.Limit)
	assert.InDelta(t, 7.0, quota.UsedPercent, 0.1)
}

func TestQuotaService_Get_NoOctetsQuota(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Quota/get", {
					"accountId": "acc1",
					"state": "q1",
					"list": [{
						"id": "quota-count",
						"resourceType": "count",
						"used": 100,
						"hardLimit": 1000,
						"scope": "account",
						"name": "Message count"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithQuota(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	quota, err := client.Quota().Get(ctx)
	assert.Error(t, err)
	assert.Nil(t, quota)
	assert.Contains(t, err.Error(), "no storage quota found")
}

func TestQuotaService_Get_EmptyList(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Quota/get", {
					"accountId": "acc1",
					"state": "q1",
					"list": [],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithQuota(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	quota, err := client.Quota().Get(ctx)
	assert.Error(t, err)
	assert.Nil(t, quota)
	assert.Contains(t, err.Error(), "no storage quota found")
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{name: "zero bytes", bytes: 0, expected: "0 B"},
		{name: "bytes", bytes: 500, expected: "500 B"},
		{name: "one KB", bytes: 1024, expected: "1.0 KB"},
		{name: "kilobytes", bytes: 2048, expected: "2.0 KB"},
		{name: "megabytes", bytes: 1048576, expected: "1.0 MB"},
		{name: "megabytes fractional", bytes: 5242880, expected: "5.0 MB"},
		{name: "gigabytes", bytes: 1073741824, expected: "1.0 GB"},
		{name: "gigabytes fractional", bytes: 2254857830, expected: "2.1 GB"},
		{name: "30 GB", bytes: 32212254720, expected: "30.0 GB"},
		{name: "terabytes", bytes: 1099511627776, expected: "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuotaService_Get_ZeroLimit(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Quota/get", {
					"accountId": "acc1",
					"state": "q1",
					"list": [{
						"id": "quota-1",
						"resourceType": "octets",
						"used": 0,
						"hardLimit": 0,
						"scope": "account",
						"name": "Storage"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithQuota(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	quota, err := client.Quota().Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, quota)

	assert.Equal(t, uint64(0), quota.Used)
	assert.Equal(t, uint64(0), quota.Limit)
	assert.Equal(t, float64(0), quota.UsedPercent)
}
