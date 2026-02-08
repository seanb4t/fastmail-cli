package fastmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vacationSessionResponse returns a valid JMAP session with mail capability.
func vacationSessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {}
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

func TestVacationService_GetStatus_Enabled(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "VacationResponse/get", methodName)
		assert.Equal(t, "acc1", args["accountId"])

		// Verify singleton ID is requested
		ids := args["ids"].([]any)
		assert.Contains(t, ids, "singleton")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/get", {
					"accountId": "acc1",
					"state": "v1",
					"list": [{
						"id": "singleton",
						"isEnabled": true,
						"fromDate": "2024-01-15T00:00:00Z",
						"toDate": "2024-01-20T00:00:00Z",
						"subject": "Out of Office",
						"textBody": "I'm away until Jan 20"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	status, err := client.Vacation().GetStatus(ctx)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.True(t, status.IsEnabled)
	assert.Equal(t, "Out of Office", status.Subject)
	assert.Equal(t, "I'm away until Jan 20", status.TextBody)
	require.NotNil(t, status.FromDate)
	assert.Equal(t, 2024, status.FromDate.Year())
	assert.Equal(t, 1, int(status.FromDate.Month()))
	assert.Equal(t, 15, status.FromDate.Day())
	require.NotNil(t, status.ToDate)
	assert.Equal(t, 20, status.ToDate.Day())
}

func TestVacationService_GetStatus_Disabled(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/get", {
					"accountId": "acc1",
					"state": "v1",
					"list": [{
						"id": "singleton",
						"isEnabled": false
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	status, err := client.Vacation().GetStatus(ctx)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.False(t, status.IsEnabled)
	assert.Empty(t, status.Subject)
	assert.Nil(t, status.FromDate)
	assert.Nil(t, status.ToDate)
}

func TestVacationService_Enable(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "VacationResponse/set", methodName)

		// Verify update payload
		update := args["update"].(map[string]any)
		singletonUpdate := update["singleton"].(map[string]any)
		assert.Equal(t, true, singletonUpdate["isEnabled"])
		assert.Equal(t, "Out of Office", singletonUpdate["subject"])
		assert.Equal(t, "I'm away until Jan 20", singletonUpdate["textBody"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v2",
					"updated": {"singleton": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Vacation().Enable(ctx, "Out of Office", "I'm away until Jan 20", nil, nil)
	require.NoError(t, err)
}

func TestVacationService_Enable_WithDates(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		update := args["update"].(map[string]any)
		singletonUpdate := update["singleton"].(map[string]any)
		assert.Equal(t, true, singletonUpdate["isEnabled"])
		assert.Contains(t, singletonUpdate, "fromDate")
		assert.Contains(t, singletonUpdate, "toDate")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v2",
					"updated": {"singleton": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	from := mustParseTime(t, "2024-01-15T00:00:00Z")
	to := mustParseTime(t, "2024-01-20T00:00:00Z")

	err := client.Vacation().Enable(ctx, "Out of Office", "I'm away", &from, &to)
	require.NoError(t, err)
}

func TestVacationService_Disable(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "VacationResponse/set", methodName)

		// Verify update payload disables vacation
		update := args["update"].(map[string]any)
		singletonUpdate := update["singleton"].(map[string]any)
		assert.Equal(t, false, singletonUpdate["isEnabled"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v2",
					"updated": {"singleton": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Vacation().Disable(ctx)
	require.NoError(t, err)
}

func TestVacationService_Enable_Error(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v1",
					"notUpdated": {
						"singleton": {
							"type": "invalidProperties",
							"description": "fromDate is not a valid date"
						}
					}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Vacation().Enable(ctx, "OOO", "away", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalidProperties")
}

// mustParseTime parses a time string in RFC3339 format, failing the test on error.
func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return parsed
}
