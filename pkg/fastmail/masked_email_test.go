package fastmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJMAPSession = "/jmap/session"

func TestMaskedEmailService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testJMAPSession {
			err := json.NewEncoder(w).Encode(map[string]any{
				"apiUrl": "http://" + r.Host + "/jmap/api/",
				"primaryAccounts": map[string]string{
					"urn:ietf:params:jmap:core": "u123",
					"urn:ietf:params:jmap:mail": "u123",
				},
				"accounts": map[string]any{
					"u123": map[string]any{
						"name":       "test@example.com",
						"isPersonal": true,
					},
				},
			})
			require.NoError(t, err)
			return
		}

		// Parse and verify the JMAP request
		var req jmap.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify we're using masked email capability
		assert.Contains(t, req.Using, "https://www.fastmail.com/dev/maskedemail")

		// Return mock masked emails
		response := map[string]any{
			"methodResponses": []any{
				[]any{
					"MaskedEmail/get",
					map[string]any{
						"accountId": "u123",
						"state":     "abc123",
						"list": []map[string]any{
							{
								"id":            "masked-001",
								"email":         "alias1@fastmail.com",
								"state":         "enabled",
								"forDomain":     "example.com",
								"description":   "Test alias",
								"createdAt":     "2024-01-15T10:30:00Z",
								"createdBy":     "user",
								"lastMessageAt": "2024-01-20T15:00:00Z",
							},
							{
								"id":          "masked-002",
								"email":       "alias2@fastmail.com",
								"state":       "disabled",
								"forDomain":   "shop.com",
								"description": "Shopping",
								"createdAt":   "2024-01-10T08:00:00Z",
								"createdBy":   "1password",
							},
						},
						"notFound": []string{},
					},
					"0",
				},
			},
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL+testJMAPSession, "test-token")
	err := client.Connect(context.Background())
	require.NoError(t, err)

	emails, err := client.MaskedEmail().List(context.Background())
	require.NoError(t, err)
	require.Len(t, emails, 2)

	// Verify first masked email
	assert.Equal(t, "masked-001", emails[0].ID)
	assert.Equal(t, "alias1@fastmail.com", emails[0].Email)
	assert.Equal(t, MaskedEmailStateEnabled, emails[0].State)
	assert.Equal(t, "example.com", emails[0].ForDomain)
	assert.Equal(t, "Test alias", emails[0].Description)
	assert.Equal(t, "user", emails[0].CreatedBy)
	assert.False(t, emails[0].CreatedAt.IsZero())
	assert.False(t, emails[0].LastMessageAt.IsZero())

	// Verify second masked email
	assert.Equal(t, "masked-002", emails[1].ID)
	assert.Equal(t, MaskedEmailStateDisabled, emails[1].State)
	assert.True(t, emails[1].LastMessageAt.IsZero())
}

func TestMaskedEmailService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testJMAPSession {
			err := json.NewEncoder(w).Encode(map[string]any{
				"apiUrl": "http://" + r.Host + "/jmap/api/",
				"primaryAccounts": map[string]string{
					"urn:ietf:params:jmap:core": "u123",
					"urn:ietf:params:jmap:mail": "u123",
				},
				"accounts": map[string]any{
					"u123": map[string]any{"name": "test@example.com", "isPersonal": true},
				},
			})
			require.NoError(t, err)
			return
		}

		var req jmap.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify the create request structure
		assert.Contains(t, req.Using, "https://www.fastmail.com/dev/maskedemail")
		assert.Equal(t, "MaskedEmail/set", req.MethodCalls[0].Name)

		response := map[string]any{
			"methodResponses": []any{
				[]any{
					"MaskedEmail/set",
					map[string]any{
						"accountId": "u123",
						"created": map[string]any{
							"new-masked": map[string]any{
								"id":        "masked-new-001",
								"email":     "generated123@fastmail.com",
								"state":     "enabled",
								"forDomain": "newsite.com",
								"createdAt": "2024-01-25T12:00:00Z",
								"createdBy": "api",
							},
						},
						"notCreated": map[string]any{},
					},
					"0",
				},
			},
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL+testJMAPSession, "test-token")
	err := client.Connect(context.Background())
	require.NoError(t, err)

	maskedEmail, err := client.MaskedEmail().Create(context.Background(), "newsite.com")
	require.NoError(t, err)

	assert.Equal(t, "masked-new-001", maskedEmail.ID)
	assert.Equal(t, "generated123@fastmail.com", maskedEmail.Email)
	assert.Equal(t, MaskedEmailStateEnabled, maskedEmail.State)
	assert.Equal(t, "newsite.com", maskedEmail.ForDomain)
}

func TestMaskedEmailService_Enable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testJMAPSession {
			err := json.NewEncoder(w).Encode(map[string]any{
				"apiUrl": "http://" + r.Host + "/jmap/api/",
				"primaryAccounts": map[string]string{
					"urn:ietf:params:jmap:core": "u123",
					"urn:ietf:params:jmap:mail": "u123",
				},
				"accounts": map[string]any{
					"u123": map[string]any{"name": "test@example.com", "isPersonal": true},
				},
			})
			require.NoError(t, err)
			return
		}

		var req jmap.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify the update request sets state to enabled
		assert.Equal(t, "MaskedEmail/set", req.MethodCalls[0].Name)
		args := req.MethodCalls[0].Args
		update := args["update"].(map[string]any)
		changes := update["masked-123"].(map[string]any)
		assert.Equal(t, "enabled", changes["state"])

		response := map[string]any{
			"methodResponses": []any{
				[]any{
					"MaskedEmail/set",
					map[string]any{
						"accountId": "u123",
						"updated": map[string]any{
							"masked-123": nil,
						},
						"notUpdated": map[string]any{},
					},
					"0",
				},
			},
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL+testJMAPSession, "test-token")
	err := client.Connect(context.Background())
	require.NoError(t, err)

	err = client.MaskedEmail().Enable(context.Background(), "masked-123")
	require.NoError(t, err)
}

func TestMaskedEmailService_Disable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testJMAPSession {
			err := json.NewEncoder(w).Encode(map[string]any{
				"apiUrl": "http://" + r.Host + "/jmap/api/",
				"primaryAccounts": map[string]string{
					"urn:ietf:params:jmap:core": "u123",
					"urn:ietf:params:jmap:mail": "u123",
				},
				"accounts": map[string]any{
					"u123": map[string]any{"name": "test@example.com", "isPersonal": true},
				},
			})
			require.NoError(t, err)
			return
		}

		var req jmap.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		args := req.MethodCalls[0].Args
		update := args["update"].(map[string]any)
		changes := update["masked-456"].(map[string]any)
		assert.Equal(t, "disabled", changes["state"])

		response := map[string]any{
			"methodResponses": []any{
				[]any{
					"MaskedEmail/set",
					map[string]any{
						"accountId":  "u123",
						"updated":    map[string]any{"masked-456": nil},
						"notUpdated": map[string]any{},
					},
					"0",
				},
			},
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL+testJMAPSession, "test-token")
	err := client.Connect(context.Background())
	require.NoError(t, err)

	err = client.MaskedEmail().Disable(context.Background(), "masked-456")
	require.NoError(t, err)
}

func TestMaskedEmailService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == testJMAPSession {
			err := json.NewEncoder(w).Encode(map[string]any{
				"apiUrl": "http://" + r.Host + "/jmap/api/",
				"primaryAccounts": map[string]string{
					"urn:ietf:params:jmap:core": "u123",
					"urn:ietf:params:jmap:mail": "u123",
				},
				"accounts": map[string]any{
					"u123": map[string]any{"name": "test@example.com", "isPersonal": true},
				},
			})
			require.NoError(t, err)
			return
		}

		var req jmap.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify the destroy request
		args := req.MethodCalls[0].Args
		destroy := args["destroy"].([]any)
		assert.Contains(t, destroy, "masked-789")

		response := map[string]any{
			"methodResponses": []any{
				[]any{
					"MaskedEmail/set",
					map[string]any{
						"accountId":    "u123",
						"destroyed":    []string{"masked-789"},
						"notDestroyed": map[string]any{},
					},
					"0",
				},
			},
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL+testJMAPSession, "test-token")
	err := client.Connect(context.Background())
	require.NoError(t, err)

	err = client.MaskedEmail().Delete(context.Background(), "masked-789")
	require.NoError(t, err)
}

func TestMaskedEmail_IsEnabled(t *testing.T) {
	tests := []struct {
		state    MaskedEmailState
		expected bool
	}{
		{MaskedEmailStateEnabled, true},
		{MaskedEmailStateDisabled, false},
		{MaskedEmailStateDeleted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			m := MaskedEmail{State: tt.state}
			assert.Equal(t, tt.expected, m.IsEnabled())
		})
	}
}

func TestMaskedEmail_HasReceivedMail(t *testing.T) {
	t.Run("with messages", func(t *testing.T) {
		m := MaskedEmail{LastMessageAt: time.Now()}
		assert.True(t, m.HasReceivedMail())
	})

	t.Run("no messages", func(t *testing.T) {
		m := MaskedEmail{}
		assert.False(t, m.HasReceivedMail())
	})
}
