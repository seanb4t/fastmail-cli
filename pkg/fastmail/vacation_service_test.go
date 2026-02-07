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

// sessionResponseWithVacation returns a valid JMAP session that includes the vacation capability.
func sessionResponseWithVacation(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"urn:ietf:params:jmap:vacationresponse": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {"urn:ietf:params:jmap:mail": {}}
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

func TestVacationService_Get(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		assert.Equal(t, "VacationResponse/get", methodName)

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
						"fromDate": "2026-01-01T00:00:00Z",
						"toDate": "2026-01-15T00:00:00Z",
						"subject": "Out of Office",
						"textBody": "I am on vacation",
						"htmlBody": "<p>I am on vacation</p>"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithVacation(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	vacation, err := client.Vacation().Get(ctx)
	require.NoError(t, err)

	assert.True(t, vacation.IsEnabled)
	assert.Equal(t, "2026-01-01T00:00:00Z", vacation.FromDate)
	assert.Equal(t, "2026-01-15T00:00:00Z", vacation.ToDate)
	assert.Equal(t, "Out of Office", vacation.Subject)
	assert.Equal(t, "I am on vacation", vacation.TextBody)
	assert.Equal(t, "<p>I am on vacation</p>", vacation.HTMLBody)
}

func TestVacationService_Get_Empty(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/get", {
					"accountId": "acc1",
					"state": "v1",
					"list": [],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithVacation(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	vacation, err := client.Vacation().Get(ctx)
	require.NoError(t, err)
	assert.False(t, vacation.IsEnabled)
	assert.Empty(t, vacation.Subject)
}

func TestVacationService_Set(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "VacationResponse/set", methodName)

		update := args["update"].(map[string]any)
		singleton := update["singleton"].(map[string]any)
		assert.Equal(t, true, singleton["isEnabled"])
		assert.Equal(t, "Out of Office", singleton["subject"])
		assert.Equal(t, "I am away", singleton["textBody"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v2",
					"updated": {"singleton": null},
					"notUpdated": {}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithVacation(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	enabled := true
	err := client.Vacation().Set(ctx, SetVacationOptions{
		IsEnabled: &enabled,
		Subject:   "Out of Office",
		TextBody:  "I am away",
	})
	require.NoError(t, err)
}
