package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii)
	os.Exit(m.Run())
}

type jmapFixtures struct {
	Session    json.RawMessage `json:"session"`
	Mailboxes  json.RawMessage `json:"mailboxes"`
	EmailQuery json.RawMessage `json:"emailQuery"`
	Emails     json.RawMessage `json:"emails"`
}

func loadFixtures(t *testing.T) jmapFixtures {
	t.Helper()
	data, err := os.ReadFile("testdata/jmap_fixtures.json")
	require.NoError(t, err)
	var f jmapFixtures
	require.NoError(t, json.Unmarshal(data, &f))
	return f
}

func newMockJMAPServer(t *testing.T, fixtures jmapFixtures) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			session := strings.ReplaceAll(string(fixtures.Session), "{{API_URL}}", server.URL)
			_, _ = w.Write([]byte(session))
		case http.MethodPost:
			var req jmap.Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var methodResponses []json.RawMessage
			for _, inv := range req.MethodCalls {
				switch inv.Name {
				case "Mailbox/get":
					methodResponses = append(methodResponses, buildMethodResponse(inv.Name, fixtures.Mailboxes, inv.CallID))
				case "Email/query":
					methodResponses = append(methodResponses, buildMethodResponse(inv.Name, fixtures.EmailQuery, inv.CallID))
				case "Email/get":
					methodResponses = append(methodResponses, buildMethodResponse(inv.Name, fixtures.Emails, inv.CallID))
				}
			}

			resp := map[string]any{
				"sessionState":    "test-state",
				"methodResponses": methodResponses,
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func buildMethodResponse(method string, data json.RawMessage, callID string) json.RawMessage {
	resp, _ := json.Marshal([]any{method, data, callID})
	return resp
}

func newTestFixture(t *testing.T, width, height int) *teatest.TestModel {
	t.Helper()
	fixtures := loadFixtures(t)
	server := newMockJMAPServer(t, fixtures)
	client := fastmail.NewClient(server.URL, "test-token", fastmail.WithHTTPClient(server.Client()))
	model := New(client)
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(width, height))
	t.Cleanup(func() {
		_ = tm.Quit()
		tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
	})
	return tm
}

func TestFixtureInfrastructure(t *testing.T) {
	fixtures := loadFixtures(t)
	require.NotEmpty(t, fixtures.Session)
	require.NotEmpty(t, fixtures.Mailboxes)
	require.NotEmpty(t, fixtures.EmailQuery)
	require.NotEmpty(t, fixtures.Emails)

	server := newMockJMAPServer(t, fixtures)
	require.NotNil(t, server)

	client := fastmail.NewClient(server.URL, "test-token", fastmail.WithHTTPClient(server.Client()))
	require.NoError(t, client.Connect(t.Context()))
}

func TestNewTestFixture(t *testing.T) {
	tm := newTestFixture(t, 120, 40)
	require.NotNil(t, tm)
}
