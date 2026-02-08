package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVacationResponse_JSON(t *testing.T) {
	vr := VacationResponse{
		ID:        "singleton",
		IsEnabled: true,
		FromDate:  "2024-01-15T00:00:00Z",
		ToDate:    "2024-01-20T00:00:00Z",
		Subject:   "Out of Office",
		TextBody:  "I'm away until Jan 20",
		HTMLBody:  "<p>I'm away until Jan 20</p>",
	}

	data, err := json.Marshal(vr)
	require.NoError(t, err)

	var parsed VacationResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, vr.ID, parsed.ID)
	assert.Equal(t, vr.IsEnabled, parsed.IsEnabled)
	assert.Equal(t, vr.FromDate, parsed.FromDate)
	assert.Equal(t, vr.ToDate, parsed.ToDate)
	assert.Equal(t, vr.Subject, parsed.Subject)
	assert.Equal(t, vr.TextBody, parsed.TextBody)
	assert.Equal(t, vr.HTMLBody, parsed.HTMLBody)
}

func TestVacationResponse_JSON_OmitEmpty(t *testing.T) {
	vr := VacationResponse{
		ID:        "singleton",
		IsEnabled: false,
	}

	data, err := json.Marshal(vr)
	require.NoError(t, err)

	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "id")
	assert.Contains(t, raw, "isEnabled")
	assert.NotContains(t, raw, "fromDate")
	assert.NotContains(t, raw, "toDate")
	assert.NotContains(t, raw, "subject")
	assert.NotContains(t, raw, "textBody")
	assert.NotContains(t, raw, "htmlBody")
}

func TestVacationResponseGetBuilder_Basic(t *testing.T) {
	args := NewVacationResponseGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"singleton"}, args["ids"])
}

func TestVacationResponseGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"state": "v1",
		"list": [
			{
				"id": "singleton",
				"isEnabled": true,
				"fromDate": "2024-01-15T00:00:00Z",
				"toDate": "2024-01-20T00:00:00Z",
				"subject": "Out of Office",
				"textBody": "I'm away until Jan 20"
			}
		],
		"notFound": []
	}`

	var resp VacationResponseGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "acc1", resp.AccountID)
	assert.Equal(t, "v1", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "singleton", resp.List[0].ID)
	assert.True(t, resp.List[0].IsEnabled)
	assert.Equal(t, "2024-01-15T00:00:00Z", resp.List[0].FromDate)
	assert.Equal(t, "2024-01-20T00:00:00Z", resp.List[0].ToDate)
	assert.Equal(t, "Out of Office", resp.List[0].Subject)
	assert.Equal(t, "I'm away until Jan 20", resp.List[0].TextBody)
	assert.Empty(t, resp.NotFound)
}

func TestVacationResponseGetResponse_Disabled(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"state": "v1",
		"list": [
			{
				"id": "singleton",
				"isEnabled": false
			}
		],
		"notFound": []
	}`

	var resp VacationResponseGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Len(t, resp.List, 1)
	assert.False(t, resp.List[0].IsEnabled)
	assert.Empty(t, resp.List[0].Subject)
	assert.Empty(t, resp.List[0].TextBody)
}

func TestVacationResponseSetBuilder_Update(t *testing.T) {
	args := NewVacationResponseSet("account-1").
		Update("singleton", map[string]any{
			"isEnabled": true,
			"subject":   "Out of Office",
			"textBody":  "I'm away",
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "singleton")
	assert.Equal(t, true, update["singleton"]["isEnabled"])
	assert.Equal(t, "Out of Office", update["singleton"]["subject"])
	assert.Equal(t, "I'm away", update["singleton"]["textBody"])
}

func TestVacationResponseSetBuilder_Disable(t *testing.T) {
	args := NewVacationResponseSet("account-1").
		Update("singleton", map[string]any{
			"isEnabled": false,
		}).
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "singleton")
	assert.Equal(t, false, update["singleton"]["isEnabled"])
}

func TestVacationResponseSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"oldState": "v1",
		"newState": "v2",
		"updated": {"singleton": null},
		"notUpdated": {}
	}`

	var resp VacationResponseSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "acc1", resp.AccountID)
	assert.Equal(t, "v1", resp.OldState)
	assert.Equal(t, "v2", resp.NewState)
	assert.Contains(t, resp.Updated, "singleton")
}

func TestVacationResponseSetResponse_WithError(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"oldState": "v1",
		"newState": "v1",
		"updated": {},
		"notUpdated": {
			"singleton": {
				"type": "invalidProperties",
				"description": "fromDate is not a valid date"
			}
		}
	}`

	var resp VacationResponseSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Contains(t, resp.NotUpdated, "singleton")
	assert.Equal(t, "invalidProperties", resp.NotUpdated["singleton"].Type)
	assert.Equal(t, "fromDate is not a valid date", resp.NotUpdated["singleton"].Description)
}

func TestVacationResponseIntegration_GetAndSet(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)

	// Get current vacation response
	getCallID := req.Invoke("VacationResponse/get",
		NewVacationResponseGet("account-1").Build(),
	)

	// Update vacation response
	req.Invoke("VacationResponse/set",
		NewVacationResponseSet("account-1").
			Update("singleton", map[string]any{
				"isEnabled": true,
				"subject":   "Out of Office",
				"textBody":  "I'm away",
			}).
			Build(),
	)

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify capabilities include mail
	using := parsed["using"].([]any)
	assert.Contains(t, using, CapMail)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 2)

	// Verify VacationResponse/get
	getCall := methodCalls[0].([]any)
	assert.Equal(t, "VacationResponse/get", getCall[0])
	assert.Equal(t, getCallID, getCall[2])

	// Verify VacationResponse/set
	setCall := methodCalls[1].([]any)
	assert.Equal(t, "VacationResponse/set", setCall[0])
	setArgs := setCall[1].(map[string]any)
	update := setArgs["update"].(map[string]any)
	singletonUpdate := update["singleton"].(map[string]any)
	assert.Equal(t, true, singletonUpdate["isEnabled"])
}
