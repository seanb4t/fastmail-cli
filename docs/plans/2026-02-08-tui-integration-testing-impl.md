# TUI Integration Testing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the two-layer integration testing strategy from the [TUI Integration Testing Design](2026-02-08-tui-integration-testing-design.md) — teatest golden file tests and VHS visual tests with WireMock.

**Architecture:** Phase 1 adds the `pkg/fastmail` option passthrough needed for test HTTP injection. Phase 2 builds the teatest mock server, fixture, and 7 integration tests with golden files. Phase 3 creates WireMock stubs, VHS tape files, and CI workflows. Phase 4 wires everything into the Taskfile and CI.

**Tech Stack:** charmbracelet/x/exp/teatest, charmbracelet/x/exp/golden, WireMock (Docker), httptest, VHS

---

## Phase 1: Client Option Passthrough (Epic: Client Testability)

> **Prerequisite:** None

### Task 1.1: Add WithHTTPClient to pkg/fastmail

`pkg/fastmail/client.go` currently creates the jmap.Client without forwarding options. We need to add a `ClientOption` type that passes through to `jmap.NewClient`, so integration tests can inject a mock HTTP transport.

**Files:**
- Modify: `pkg/fastmail/client.go`
- Test: `pkg/fastmail/client_test.go`

**Step 1: Write the failing test**

Create `pkg/fastmail/client_test.go` (or append to it if it exists):

```go
package fastmail

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_WithHTTPClient(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {"urn:ietf:params:jmap:core": {}, "urn:ietf:params:jmap:mail": {}},
			"accounts": {"u1": {"name": "test", "isPersonal": true, "isReadOnly": false, "accountCapabilities": {"urn:ietf:params:jmap:mail": {}}}},
			"primaryAccounts": {"urn:ietf:params:jmap:mail": "u1"},
			"username": "test@example.com",
			"apiUrl": "` + srv.URL + `/api/",
			"downloadUrl": "` + srv.URL + `/download/{accountId}/{blobId}/{name}",
			"uploadUrl": "` + srv.URL + `/upload/{accountId}/",
			"eventSourceUrl": "` + srv.URL + `/events/",
			"state": "s1"
		}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token", WithHTTPClient(srv.Client()))
	err := client.Connect(t.Context())
	require.NoError(t, err)
	assert.True(t, called, "custom HTTP client should be used")
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestNewClient_WithHTTPClient ./pkg/fastmail/ -v`
Expected: FAIL — `WithHTTPClient` undefined, `NewClient` doesn't accept options

**Step 3: Write minimal implementation**

Edit `pkg/fastmail/client.go`:

```go
import (
	"context"
	"net/http"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// ClientOption configures a Client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	httpClient *http.Client
}

// WithHTTPClient sets a custom HTTP client, useful for testing.
func WithHTTPClient(c *http.Client) ClientOption {
	return func(cfg *clientConfig) {
		cfg.httpClient = c
	}
}

// NewClient creates a new Fastmail client.
// The endpoint should be the JMAP session URL (e.g., https://api.fastmail.com/jmap/session).
func NewClient(endpoint, accessToken string, opts ...ClientOption) *Client {
	var cfg clientConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var jmapOpts []jmap.ClientOption
	if cfg.httpClient != nil {
		jmapOpts = append(jmapOpts, jmap.WithHTTPClient(cfg.httpClient))
	}

	return &Client{
		jmap: jmap.NewClient(endpoint, accessToken, jmapOpts...),
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestNewClient_WithHTTPClient ./pkg/fastmail/ -v`
Expected: PASS

**Step 5: Run full test suite to verify no regressions**

Run: `go test -race ./...`
Expected: All existing tests pass (callers of `NewClient` without options still work)

**Step 6: Lint**

Run: `golangci-lint run ./pkg/fastmail/`
Expected: 0 issues

**Step 7: Commit**

```bash
git add pkg/fastmail/client.go pkg/fastmail/client_test.go
git commit -m "feat(fastmail): add WithHTTPClient option for test HTTP injection"
```

**Acceptance Criteria:**
- `NewClient(endpoint, token)` still works (backward compatible)
- `NewClient(endpoint, token, WithHTTPClient(c))` injects custom HTTP client
- Custom client is used for session discovery (Authenticate/Connect)
- All existing tests pass without modification

---

## Phase 2: teatest Integration Tests (Epic: Programmatic TUI Testing)

> **Prerequisite:** Phase 1 (WithHTTPClient)

### Task 2.1: Add teatest dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add the dependency**

Run: `go get github.com/charmbracelet/x/exp/teatest@latest`

**Step 2: Tidy**

Run: `go mod tidy`

**Step 3: Verify import works**

Run: `go build ./...`
Expected: Clean build

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add charmbracelet/x/exp/teatest dependency"
```

**Acceptance Criteria:**
- `go.mod` includes `github.com/charmbracelet/x/exp/teatest`
- `go build ./...` succeeds

---

### Task 2.2: Create JMAP fixture data

The mock JMAP server needs canned responses. This JSON file contains all the response bodies the server will serve.

**Files:**
- Create: `internal/tui/testdata/jmap_fixtures.json`

**Step 1: Create the fixture file**

```json
{
  "session": {
    "capabilities": {
      "urn:ietf:params:jmap:core": {
        "maxSizeUpload": 50000000,
        "maxConcurrentUpload": 4,
        "maxSizeRequest": 10000000,
        "maxConcurrentRequests": 4,
        "maxCallsInRequest": 16,
        "maxObjectsInGet": 500,
        "maxObjectsInSet": 500,
        "collationAlgorithms": ["i;ascii-casemap"]
      },
      "urn:ietf:params:jmap:mail": {}
    },
    "accounts": {
      "u12345": {
        "name": "test@example.com",
        "isPersonal": true,
        "isReadOnly": false,
        "accountCapabilities": {
          "urn:ietf:params:jmap:mail": {}
        }
      }
    },
    "primaryAccounts": {
      "urn:ietf:params:jmap:mail": "u12345"
    },
    "username": "test@example.com",
    "apiUrl": "{{API_URL}}",
    "downloadUrl": "{{API_URL}}/download/{accountId}/{blobId}/{name}",
    "uploadUrl": "{{API_URL}}/upload/{accountId}/",
    "eventSourceUrl": "{{API_URL}}/events/",
    "state": "s1"
  },
  "mailboxes": {
    "accountId": "u12345",
    "state": "m1",
    "list": [
      {"id": "mb-inbox", "name": "Inbox", "role": "inbox", "sortOrder": 1, "totalEmails": 10, "unreadEmails": 3, "totalThreads": 8, "unreadThreads": 2},
      {"id": "mb-drafts", "name": "Drafts", "role": "drafts", "sortOrder": 2, "totalEmails": 2, "unreadEmails": 0, "totalThreads": 2, "unreadThreads": 0},
      {"id": "mb-sent", "name": "Sent", "role": "sent", "sortOrder": 3, "totalEmails": 150, "unreadEmails": 0, "totalThreads": 100, "unreadThreads": 0},
      {"id": "mb-archive", "name": "Archive", "role": "archive", "sortOrder": 4, "totalEmails": 5000, "unreadEmails": 0, "totalThreads": 3000, "unreadThreads": 0},
      {"id": "mb-trash", "name": "Trash", "role": "trash", "sortOrder": 5, "totalEmails": 12, "unreadEmails": 0, "totalThreads": 10, "unreadThreads": 0}
    ],
    "notFound": []
  },
  "emailQuery": {
    "accountId": "u12345",
    "queryState": "q1",
    "canCalculateChanges": false,
    "position": 0,
    "total": 5,
    "ids": ["em-1", "em-2", "em-3", "em-4", "em-5"]
  },
  "emails": {
    "accountId": "u12345",
    "state": "e1",
    "list": [
      {
        "id": "em-1", "threadId": "t1", "mailboxIds": {"mb-inbox": true},
        "from": [{"name": "Alice Smith", "email": "alice@example.com"}],
        "to": [{"name": "Test User", "email": "test@example.com"}],
        "subject": "Weekly team sync notes",
        "receivedAt": "2026-02-08T10:30:00Z",
        "preview": "Here are the notes from today's sync meeting. Action items: review Q1 targets...",
        "keywords": {"$seen": true},
        "size": 4096,
        "hasAttachment": false,
        "textBody": [{"partId": "1", "type": "text/plain"}],
        "bodyValues": {"1": {"value": "Here are the notes from today's sync meeting.\n\nAction items:\n1. Review Q1 targets\n2. Update roadmap\n3. Schedule follow-up\n\nBest,\nAlice"}}
      },
      {
        "id": "em-2", "threadId": "t2", "mailboxIds": {"mb-inbox": true},
        "from": [{"name": "Bob Jones", "email": "bob@example.com"}],
        "to": [{"name": "Test User", "email": "test@example.com"}],
        "subject": "Re: Project proposal draft",
        "receivedAt": "2026-02-08T09:15:00Z",
        "preview": "Looks good! I have a few suggestions on the timeline section...",
        "keywords": {"$flagged": true},
        "size": 2048,
        "hasAttachment": false,
        "textBody": [{"partId": "1", "type": "text/plain"}],
        "bodyValues": {"1": {"value": "Looks good! I have a few suggestions on the timeline section.\n\nBob"}}
      },
      {
        "id": "em-3", "threadId": "t3", "mailboxIds": {"mb-inbox": true},
        "from": [{"name": "Newsletter", "email": "news@example.com"}],
        "to": [{"name": "Test User", "email": "test@example.com"}],
        "subject": "Your weekly digest",
        "receivedAt": "2026-02-08T08:00:00Z",
        "preview": "This week's top stories: new Go release, TUI frameworks...",
        "keywords": {},
        "size": 8192,
        "hasAttachment": true,
        "textBody": [{"partId": "1", "type": "text/plain"}],
        "bodyValues": {"1": {"value": "This week's top stories:\n\n1. New Go release\n2. TUI frameworks comparison\n3. Cloud-native best practices"}}
      },
      {
        "id": "em-4", "threadId": "t4", "mailboxIds": {"mb-inbox": true},
        "from": [{"name": "Carol Davis", "email": "carol@example.com"}],
        "to": [{"name": "Test User", "email": "test@example.com"}],
        "subject": "Lunch tomorrow?",
        "receivedAt": "2026-02-07T16:45:00Z",
        "preview": "Hey! Want to grab lunch tomorrow? I was thinking that new place...",
        "keywords": {"$seen": true},
        "size": 1024,
        "hasAttachment": false,
        "textBody": [{"partId": "1", "type": "text/plain"}],
        "bodyValues": {"1": {"value": "Hey! Want to grab lunch tomorrow?\n\nI was thinking that new place on Main St.\n\nCarol"}}
      },
      {
        "id": "em-5", "threadId": "t5", "mailboxIds": {"mb-inbox": true},
        "from": [{"name": "System", "email": "noreply@example.com"}],
        "to": [{"name": "Test User", "email": "test@example.com"}],
        "subject": "Your account security summary",
        "receivedAt": "2026-02-07T12:00:00Z",
        "preview": "Monthly security report: no suspicious activity detected...",
        "keywords": {"$seen": true},
        "size": 3072,
        "hasAttachment": false,
        "textBody": [{"partId": "1", "type": "text/plain"}],
        "bodyValues": {"1": {"value": "Monthly security report:\n\nNo suspicious activity detected.\nLast login: 2026-02-07 from 192.168.1.1"}}
      }
    ],
    "notFound": []
  }
}
```

**Step 2: Commit**

```bash
git add internal/tui/testdata/jmap_fixtures.json
git commit -m "test(tui): add canned JMAP fixture data for integration tests"
```

**Acceptance Criteria:**
- File contains valid JSON
- 5 mailboxes with realistic roles, counts
- 5 emails with varying states (read/unread/flagged/attachment)
- Session includes all required JMAP fields
- `{{API_URL}}` placeholder for test server URL substitution

---

### Task 2.3: Create mock JMAP server and test fixture

The mock server handles session discovery and JMAP method calls using the fixture data. The test fixture wires it to a `tui.Model` via `teatest.TestModel`.

**Files:**
- Create: `internal/tui/testfixture_test.go`

**Step 1: Write a test that uses the fixture**

```go
package tui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	teatest "github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"

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

func newMockJMAPServer(t *testing.T) *httptest.Server {
	t.Helper()
	fixtures := loadFixtures(t)

	mux := http.NewServeMux()

	// Session discovery — replace {{API_URL}} placeholder with actual server URL
	var srv *httptest.Server
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			sessionJSON := strings.ReplaceAll(string(fixtures.Session), "{{API_URL}}", srv.URL)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(sessionJSON))
			return
		}
		http.NotFound(w, r)
	})

	// JMAP API endpoint — route by method name in request body
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)

		w.Header().Set("Content-Type", "application/json")

		var methodResponses []json.RawMessage

		if strings.Contains(bodyStr, "Mailbox/get") {
			resp, _ := json.Marshal([]any{"Mailbox/get", fixtures.Mailboxes, "c0"})
			methodResponses = append(methodResponses, resp)
		}
		if strings.Contains(bodyStr, "Email/query") {
			resp, _ := json.Marshal([]any{"Email/query", fixtures.EmailQuery, "c1"})
			methodResponses = append(methodResponses, resp)
		}
		if strings.Contains(bodyStr, "Email/get") {
			resp, _ := json.Marshal([]any{"Email/get", fixtures.Emails, "c2"})
			methodResponses = append(methodResponses, resp)
		}

		result := map[string]any{
			"methodResponses":  methodResponses,
			"sessionState":     "s1",
		}
		_ = json.NewEncoder(w).Encode(result)
	})

	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestFixture(t *testing.T, width, height int) *teatest.TestModel {
	t.Helper()
	srv := newMockJMAPServer(t)
	client := fastmail.NewClient(srv.URL, "test-token", fastmail.WithHTTPClient(srv.Client()))
	model := New(client)
	return teatest.NewTestModel(t, model, teatest.WithInitialTermSize(width, height))
}

const defaultTestTimeout = 10 * time.Second
```

**Step 2: Run to verify it compiles**

Run: `go test -run TestMain ./internal/tui/ -v`
Expected: PASS (TestMain runs and exits)

**Step 3: Commit**

```bash
git add internal/tui/testfixture_test.go
git commit -m "test(tui): add mock JMAP server and teatest fixture helper"
```

**Acceptance Criteria:**
- `TestMain` sets `lipgloss.SetColorProfile(termenv.Ascii)` for deterministic golden files
- `newMockJMAPServer` serves session, mailbox, and email responses from fixture JSON
- `newTestFixture` creates a fully wired `teatest.TestModel` at given dimensions
- `{{API_URL}}` placeholder in session JSON is replaced with actual test server URL
- All existing tests still pass (TestMain doesn't break them)

---

### Task 2.4: Write DashboardLoad integration test

The first integration test verifies the TUI starts up, connects, loads mailboxes, and auto-selects inbox.

**Files:**
- Create: `internal/tui/integration_test.go`

**Step 1: Write the failing test**

```go
package tui

import (
	"testing"
	"time"

	teatest "github.com/charmbracelet/x/exp/teatest"
)

func TestIntegration_DashboardLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for dashboard to load with mailboxes and emails
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return strings.Contains(s, "Inbox") && strings.Contains(s, "Drafts")
	}, teatest.WithDuration(defaultTestTimeout))

	// Send quit and capture final output
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))

	// Verify golden file
	golden.RequireEqual(t, out)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestIntegration_DashboardLoad ./internal/tui/ -v`
Expected: FAIL — golden file doesn't exist yet

**Step 3: Generate the golden file**

Run: `go test -run TestIntegration_DashboardLoad ./internal/tui/ -v -update`
Expected: Creates `internal/tui/testdata/TestIntegration_DashboardLoad.golden`

**Step 4: Inspect the golden file**

Run: `cat internal/tui/testdata/TestIntegration_DashboardLoad.golden`
Expected: Contains "Inbox", "Drafts", stats bar, email list content

**Step 5: Run test again to verify it passes**

Run: `go test -run TestIntegration_DashboardLoad ./internal/tui/ -v`
Expected: PASS — output matches golden file

**Step 6: Commit**

```bash
git add internal/tui/integration_test.go internal/tui/testdata/TestIntegration_DashboardLoad.golden
git commit -m "test(tui): add DashboardLoad integration test with golden file"
```

**Acceptance Criteria:**
- Test creates model at 120x40, waits for mailbox data to load
- Golden file captures: stats bar (unread count, "Fastmail CLI"), sidebar (5 mailboxes), inbox auto-selected, email list with 5 emails, key bar
- Test is skippable with `-short` flag
- `golden.RequireEqual` compares output deterministically (ASCII profile)

---

### Task 2.5: Write SelectMailbox_JK and SelectMailbox_Arrows integration tests

Two tests verifying mailbox selection works with both j/k and arrow keys.

**Files:**
- Modify: `internal/tui/integration_test.go`

**Step 1: Write the failing tests**

```go
func TestIntegration_SelectMailbox_JK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Inbox")
	}, teatest.WithDuration(defaultTestTimeout))

	// Tab to sidebar, move down twice with j, select with Enter
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(500 * time.Millisecond)

	// Quit and capture
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	golden.RequireEqual(t, out)
}

func TestIntegration_SelectMailbox_Arrows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Inbox")
	}, teatest.WithDuration(defaultTestTimeout))

	// Tab to sidebar, move down twice with arrows, select with Enter
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(500 * time.Millisecond)

	// Quit and capture
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	golden.RequireEqual(t, out)
}
```

**Step 2: Generate golden files**

Run: `go test -run 'TestIntegration_SelectMailbox' ./internal/tui/ -v -update`
Expected: Creates two `.golden` files

**Step 3: Verify both golden files show the same selected mailbox**

Run: `diff internal/tui/testdata/TestIntegration_SelectMailbox_JK.golden internal/tui/testdata/TestIntegration_SelectMailbox_Arrows.golden`
Expected: Identical (or near-identical) — proving j/k and arrows behave the same

**Step 4: Run tests**

Run: `go test -run 'TestIntegration_SelectMailbox' ./internal/tui/ -v`
Expected: Both PASS

**Step 5: Commit**

```bash
git add internal/tui/integration_test.go internal/tui/testdata/TestIntegration_SelectMailbox_*.golden
git commit -m "test(tui): add mailbox selection integration tests for j/k and arrow keys"
```

**Acceptance Criteria:**
- Both tests navigate to the same mailbox (2 positions down from Inbox)
- Golden files are identical or near-identical, proving navigation parity
- Arrow key test uses `tea.KeyDown`, j/k test uses `tea.KeyRunes`

---

### Task 2.6: Write ReadEmail integration test

**Files:**
- Modify: `internal/tui/integration_test.go`

**Step 1: Write the failing test**

```go
func TestIntegration_ReadEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for emails to load
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Weekly team sync")
	}, teatest.WithDuration(defaultTestTimeout))

	// Open first email
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(500 * time.Millisecond)

	// Quit and capture
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	golden.RequireEqual(t, out)
}
```

**Step 2: Generate golden file and run**

Run: `go test -run TestIntegration_ReadEmail ./internal/tui/ -v -update`
Run: `go test -run TestIntegration_ReadEmail ./internal/tui/ -v`
Expected: PASS — golden file shows email reader with subject, sender, body

**Step 3: Commit**

```bash
git add internal/tui/integration_test.go internal/tui/testdata/TestIntegration_ReadEmail.golden
git commit -m "test(tui): add ReadEmail integration test"
```

**Acceptance Criteria:**
- Golden file shows the full email reader view
- Subject "Weekly team sync notes" visible
- Sender "Alice Smith" visible
- Body content rendered

---

### Task 2.7: Write ComposeEmail integration test

**Files:**
- Modify: `internal/tui/integration_test.go`

**Step 1: Write the failing test**

```go
func TestIntegration_ComposeEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for dashboard
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Inbox")
	}, teatest.WithDuration(defaultTestTimeout))

	// Press c to compose
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	time.Sleep(500 * time.Millisecond)

	// Quit and capture
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	golden.RequireEqual(t, out)
}
```

**Step 2: Generate and run**

Run: `go test -run TestIntegration_ComposeEmail ./internal/tui/ -v -update`
Run: `go test -run TestIntegration_ComposeEmail ./internal/tui/ -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/integration_test.go internal/tui/testdata/TestIntegration_ComposeEmail.golden
git commit -m "test(tui): add ComposeEmail integration test"
```

**Acceptance Criteria:**
- Golden file shows compose overlay with To/Subject/Body fields
- Compose overlay appears over dashboard

---

### Task 2.8: Write ToggleSidebar integration test

**Files:**
- Modify: `internal/tui/integration_test.go`

**Step 1: Write the failing test**

```go
func TestIntegration_ToggleSidebar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for dashboard
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Inbox")
	}, teatest.WithDuration(defaultTestTimeout))

	// Toggle sidebar off
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	time.Sleep(300 * time.Millisecond)

	// Quit and capture
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	golden.RequireEqual(t, out)
}
```

**Step 2: Generate and run**

Run: `go test -run TestIntegration_ToggleSidebar ./internal/tui/ -v -update`
Run: `go test -run TestIntegration_ToggleSidebar ./internal/tui/ -v`
Expected: PASS — sidebar hidden, main pane takes full width

**Step 3: Commit**

```bash
git add internal/tui/integration_test.go internal/tui/testdata/TestIntegration_ToggleSidebar.golden
git commit -m "test(tui): add ToggleSidebar integration test"
```

**Acceptance Criteria:**
- Golden file shows dashboard WITHOUT sidebar
- Main pane spans full terminal width
- "Mailboxes" header NOT visible in output

---

### Task 2.9: Write NarrowTerminal integration test

**Files:**
- Modify: `internal/tui/integration_test.go`

**Step 1: Write the failing test**

```go
func TestIntegration_NarrowTerminal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 60, 20)

	// Wait for dashboard
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Inbox")
	}, teatest.WithDuration(defaultTestTimeout))

	// Quit and capture
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	golden.RequireEqual(t, out)
}
```

**Step 2: Generate and run**

Run: `go test -run TestIntegration_NarrowTerminal ./internal/tui/ -v -update`
Run: `go test -run TestIntegration_NarrowTerminal ./internal/tui/ -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/integration_test.go internal/tui/testdata/TestIntegration_NarrowTerminal.golden
git commit -m "test(tui): add NarrowTerminal integration test for responsive layout"
```

**Acceptance Criteria:**
- Terminal is 60x20 (below 80-col sidebar auto-collapse threshold)
- Sidebar is auto-collapsed (not visible)
- Key bar is abbreviated (shorter key descriptions or key-only)
- Layout fits entirely within 60x20 without overflow

---

### Task 2.10: Write BDD tests for integration layer

Add BDD-style tests that describe the integration testing scenarios in plain language.

**Files:**
- Create: `internal/tui/bdd_integration_test.go`

**Step 1: Write BDD tests**

```go
package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	teatest "github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
)

func TestBDD_Integration_DashboardShowsAllPanes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Given: a TUI connected to a JMAP server with 5 mailboxes and 5 emails
	tm := newTestFixture(t, 120, 40)

	// When: the dashboard loads
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return strings.Contains(s, "Inbox") && strings.Contains(s, "Weekly team sync")
	}, teatest.WithDuration(defaultTestTimeout))

	// Then: all three panes are visible
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	output := string(out)

	assert.Contains(t, output, "Mailboxes", "sidebar should show mailbox header")
	assert.Contains(t, output, "Inbox", "sidebar should list Inbox")
	assert.Contains(t, output, "Fastmail CLI", "stats bar should show branding")
	assert.Contains(t, output, "tab", "key bar should show tab binding")
}

func TestBDD_Integration_KeyboardNavParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Given: two TUI instances
	tmJK := newTestFixture(t, 120, 40)
	tmArrow := newTestFixture(t, 120, 40)

	// When: both load
	for _, tm := range []*teatest.TestModel{tmJK, tmArrow} {
		teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
			return strings.Contains(string(bts), "Inbox")
		}, teatest.WithDuration(defaultTestTimeout))
	}

	// And: one uses j/k, the other uses arrows to navigate
	tmJK.Send(tea.KeyMsg{Type: tea.KeyTab})
	tmArrow.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)

	tmJK.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	tmArrow.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	// Then: both end up at the same selection
	tmJK.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tmArrow.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	outJK := string(tmJK.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second)))
	outArrow := string(tmArrow.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second)))

	assert.Equal(t, outJK, outArrow, "j/k and arrow navigation should produce identical output")
}

func TestBDD_Integration_ResponsiveLayout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Given: a narrow terminal (60 cols, below 80-col threshold)
	tm := newTestFixture(t, 60, 20)

	// When: the dashboard loads
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Inbox")
	}, teatest.WithDuration(defaultTestTimeout))

	// Then: sidebar is auto-collapsed
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	out := string(tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second)))

	assert.NotContains(t, out, "Mailboxes", "sidebar should be auto-collapsed on narrow terminal")
	assert.Contains(t, out, "Inbox", "email list should still show Inbox content")
}
```

**Step 2: Run BDD tests**

Run: `go test -run TestBDD_Integration ./internal/tui/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add internal/tui/bdd_integration_test.go
git commit -m "test(tui): add BDD integration tests for dashboard, navigation parity, responsive layout"
```

**Acceptance Criteria:**
- `TestBDD_Integration_DashboardShowsAllPanes` — stats bar, sidebar, email list, key bar all present
- `TestBDD_Integration_KeyboardNavParity` — j/k and arrow keys produce identical output
- `TestBDD_Integration_ResponsiveLayout` — sidebar auto-collapses below 80 cols

---

### Task 2.11: Add .gitattributes for golden files

**Files:**
- Create: `.gitattributes`

**Step 1: Create .gitattributes**

```
*.golden -text
docs/vhs/*.gif binary
```

**Step 2: Commit**

```bash
git add .gitattributes
git commit -m "build: add .gitattributes for golden files and VHS GIFs"
```

**Acceptance Criteria:**
- `.golden` files marked as `-text` (no line-ending normalization)
- VHS GIFs marked as `binary`

---

## Phase 3: VHS Visual Tests (Epic: VHS Tape Files & WireMock)

> **Prerequisite:** Phase 2 (teatest tests working, fixture data exists)

### Task 3.1: Create WireMock stub mappings

WireMock serves the same canned JMAP data as the teatest mock server, but as a standalone Docker container for VHS tape files.

**Files:**
- Create: `docs/vhs/wiremock/mappings/session.json`
- Create: `docs/vhs/wiremock/mappings/mailbox-get.json`
- Create: `docs/vhs/wiremock/mappings/email-query.json`
- Create: `docs/vhs/wiremock/mappings/email-get.json`
- Create: `docs/vhs/wiremock/__files/session.json`
- Create: `docs/vhs/wiremock/__files/mailboxes.json`
- Create: `docs/vhs/wiremock/__files/emails.json`

**Step 1: Create mappings**

`docs/vhs/wiremock/mappings/session.json`:
```json
{
  "request": {
    "method": "GET",
    "url": "/"
  },
  "response": {
    "status": 200,
    "bodyFileName": "session.json",
    "headers": { "Content-Type": "application/json" }
  }
}
```

`docs/vhs/wiremock/mappings/mailbox-get.json`:
```json
{
  "request": {
    "method": "POST",
    "url": "/api/",
    "bodyPatterns": [{ "contains": "Mailbox/get" }]
  },
  "response": {
    "status": 200,
    "bodyFileName": "mailboxes.json",
    "headers": { "Content-Type": "application/json" }
  }
}
```

`docs/vhs/wiremock/mappings/email-query.json`:
```json
{
  "request": {
    "method": "POST",
    "url": "/api/",
    "bodyPatterns": [{ "contains": "Email/query" }]
  },
  "response": {
    "status": 200,
    "bodyFileName": "emails.json",
    "headers": { "Content-Type": "application/json" }
  }
}
```

`docs/vhs/wiremock/mappings/email-get.json`:
```json
{
  "request": {
    "method": "POST",
    "url": "/api/",
    "bodyPatterns": [{ "contains": "Email/get" }]
  },
  "response": {
    "status": 200,
    "bodyFileName": "emails.json",
    "headers": { "Content-Type": "application/json" }
  }
}
```

**Step 2: Create response bodies**

`docs/vhs/wiremock/__files/session.json`: Copy the session object from `jmap_fixtures.json`, replacing `{{API_URL}}` with `http://localhost:8080`.

`docs/vhs/wiremock/__files/mailboxes.json`: Wrap the mailboxes fixture in a JMAP response envelope:
```json
{
  "methodResponses": [
    ["Mailbox/get", {"accountId": "u12345", "state": "m1", "list": [...], "notFound": []}, "c0"]
  ],
  "sessionState": "s1"
}
```

`docs/vhs/wiremock/__files/emails.json`: Wrap both Email/query and Email/get in a combined response:
```json
{
  "methodResponses": [
    ["Email/query", {"accountId": "u12345", "queryState": "q1", ...}, "c1"],
    ["Email/get", {"accountId": "u12345", "state": "e1", "list": [...], "notFound": []}, "c2"]
  ],
  "sessionState": "s1"
}
```

**Step 3: Test WireMock locally**

Run: `docker run -d --name fm-mock -p 8080:8080 -v $(pwd)/docs/vhs/wiremock:/home/wiremock wiremock/wiremock:latest`
Run: `curl -s http://localhost:8080/ | jq .username`
Expected: `"test@example.com"`
Run: `docker rm -f fm-mock`

**Step 4: Commit**

```bash
git add docs/vhs/wiremock/
git commit -m "test(vhs): add WireMock stub mappings and response bodies"
```

**Acceptance Criteria:**
- `curl http://localhost:8080/` returns session JSON
- `curl -X POST http://localhost:8080/api/ -d '{"methodCalls": [["Mailbox/get", {}, "c0"]]}'` returns mailbox data
- WireMock matches on `bodyPatterns.contains` for method routing

---

### Task 3.2: Create VHS tape files

Three tape files for the key workflows defined in the design.

**Files:**
- Create: `docs/vhs/dashboard.tape`
- Create: `docs/vhs/compose.tape`
- Create: `docs/vhs/navigation.tape`

**Step 1: Create dashboard.tape**

```
Output docs/vhs/dashboard.gif
Output docs/vhs/testdata/dashboard.ascii

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set TypingSpeed 50ms

Type "FASTMAIL_ENDPOINT=http://localhost:8080 ./bin/fastmail-cli tui"
Enter
Sleep 3s

# Browse emails
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Read an email
Enter
Sleep 2s

# Go back
Type "q"
Sleep 500ms
```

**Step 2: Create compose.tape**

```
Output docs/vhs/compose.gif
Output docs/vhs/testdata/compose.ascii

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set TypingSpeed 50ms

Type "FASTMAIL_ENDPOINT=http://localhost:8080 ./bin/fastmail-cli tui"
Enter
Sleep 3s

# Open compose
Type "c"
Sleep 1s

# Escape compose
Escape
Sleep 500ms

Type "q"
Sleep 500ms
```

**Step 3: Create navigation.tape**

```
Output docs/vhs/navigation.gif
Output docs/vhs/testdata/navigation.ascii

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set TypingSpeed 50ms

Type "FASTMAIL_ENDPOINT=http://localhost:8080 ./bin/fastmail-cli tui"
Enter
Sleep 3s

# Toggle sidebar
Type "b"
Sleep 1s

# Toggle back
Type "b"
Sleep 1s

# Cycle panes
Tab
Sleep 500ms
Tab
Sleep 500ms
Tab
Sleep 500ms

Type "q"
Sleep 500ms
```

**Step 4: Create testdata directory**

Run: `mkdir -p docs/vhs/testdata`

**Step 5: Commit**

```bash
git add docs/vhs/*.tape
git commit -m "test(vhs): add dashboard, compose, and navigation tape files"
```

**Acceptance Criteria:**
- Each tape outputs both `.gif` and `.ascii` to `testdata/`
- Tape files use `FASTMAIL_ENDPOINT` env var to point to WireMock
- Each tape ends with quit to cleanly exit

---

### Task 3.3: Generate initial VHS golden files

Run the tapes locally and generate baseline golden files.

**Step 1: Build the binary**

Run: `task build`

**Step 2: Start WireMock**

Run: `docker run -d --name fm-mock -p 8080:8080 -v $(pwd)/docs/vhs/wiremock:/home/wiremock wiremock/wiremock:latest`
Run: `sleep 2`

**Step 3: Run VHS**

Run: `vhs docs/vhs/dashboard.tape`
Run: `vhs docs/vhs/compose.tape`
Run: `vhs docs/vhs/navigation.tape`

**Step 4: Stop WireMock**

Run: `docker rm -f fm-mock`

**Step 5: Create expected directory and copy golden files**

Run: `mkdir -p docs/vhs/expected`
Run: `cp docs/vhs/testdata/*.ascii docs/vhs/expected/`

**Step 6: Add GIFs to .gitignore**

Append to `.gitignore`:
```
docs/vhs/*.gif
```

**Step 7: Commit**

```bash
git add docs/vhs/testdata/ docs/vhs/expected/ .gitignore
git commit -m "test(vhs): generate initial golden ASCII files from VHS recordings"
```

**Acceptance Criteria:**
- `docs/vhs/testdata/` contains ASCII captures from each tape
- `docs/vhs/expected/` contains copies as baselines
- GIFs are gitignored

---

### Task 3.4: Write BDD test for VHS workflow

A BDD-style test that validates the VHS golden file comparison works.

**Files:**
- Create: `docs/vhs/vhs_test.go` (or as a shell script)

Since VHS tests are script-based, add a BDD description as a test that verifies the expected golden files exist and match:

**Step 1: Create a simple shell-based check**

This is better as a Taskfile target (Task 4.2) rather than a Go test, since it depends on Docker and VHS being installed. Skip this task — the BDD coverage comes from the teatest BDD tests in Task 2.10.

---

## Phase 4: CI Integration & Taskfile (Epic: CI Wiring)

> **Prerequisite:** Phase 3 (VHS tapes and WireMock stubs exist)

### Task 4.1: Add Taskfile entries

**Files:**
- Modify: `Taskfile.yaml`

**Step 1: Add the new task targets**

Append to `Taskfile.yaml`:

```yaml
  test:integration:
    desc: Run teatest integration tests only
    cmds:
      - go test -race -run Integration ./internal/tui/ -v

  test:golden:update:
    desc: Regenerate all teatest golden files
    cmds:
      - go test ./internal/tui/ -run Integration -update

  vhs:mock:start:
    desc: Start WireMock for VHS recording
    cmds:
      - docker run -d --name fm-mock -p 8080:8080
        -v {{.ROOT_DIR}}/docs/vhs/wiremock:/home/wiremock
        wiremock/wiremock:latest
      - sleep 2

  vhs:mock:stop:
    desc: Stop WireMock
    cmds:
      - docker rm -f fm-mock 2>/dev/null || true

  vhs:
    desc: Generate all VHS recordings
    deps: [build]
    cmds:
      - task: vhs:mock:start
      - defer: { task: vhs:mock:stop }
      - for: [dashboard, compose, navigation]
        cmd: vhs docs/vhs/{{.ITEM}}.tape

  vhs:check:
    desc: Verify VHS golden files match
    cmds:
      - for: [dashboard, compose, navigation]
        cmd: diff docs/vhs/testdata/{{.ITEM}}.ascii docs/vhs/expected/{{.ITEM}}.ascii

  vhs:update:
    desc: Regenerate VHS golden files
    cmds:
      - cp docs/vhs/testdata/*.ascii docs/vhs/expected/
```

**Step 2: Verify task list**

Run: `task --list`
Expected: New tasks appear in listing

**Step 3: Commit**

```bash
git add Taskfile.yaml
git commit -m "build: add Taskfile entries for integration tests and VHS workflows"
```

**Acceptance Criteria:**
- `task test:integration` runs teatest integration tests
- `task test:golden:update` regenerates golden files
- `task vhs` builds binary, starts WireMock, runs all tapes, stops WireMock
- `task vhs:check` diffs ASCII golden files
- `task vhs:mock:stop` doesn't fail if container doesn't exist (2>/dev/null || true)

---

### Task 4.2: Update CI workflow

**Files:**
- Modify: `.github/workflows/ci.yaml`
- Create: `.github/workflows/vhs.yaml`

**Step 1: Add LIPGLOSS_COLOR_PROFILE to CI test step**

Edit `.github/workflows/ci.yaml`, update the test step:

```yaml
      - name: Run tests
        env:
          LIPGLOSS_COLOR_PROFILE: ascii
        run: go test -race ./...
```

**Step 2: Create VHS workflow**

Create `.github/workflows/vhs.yaml`:

```yaml
name: VHS Visual Tests

on:
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  vhs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install Task
        uses: arduino/setup-task@v2
        with:
          version: 3.x

      - name: Build
        run: task build

      - name: Start WireMock
        run: task vhs:mock:start

      - name: Run VHS
        uses: charmbracelet/vhs-action@v2
        with:
          path: "docs/vhs/*.tape"

      - name: Stop WireMock
        if: always()
        run: task vhs:mock:stop

      - name: Compare golden ASCII
        run: task vhs:check

      - name: Upload recordings
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: vhs-recordings
          path: docs/vhs/*.gif
```

**Step 3: Commit**

```bash
git add .github/workflows/ci.yaml .github/workflows/vhs.yaml
git commit -m "ci: add LIPGLOSS_COLOR_PROFILE and VHS visual test workflow"
```

**Acceptance Criteria:**
- CI test step sets `LIPGLOSS_COLOR_PROFILE=ascii` for deterministic golden files
- VHS workflow runs on PRs only
- WireMock stop runs even if VHS fails (`if: always()`)
- GIFs uploaded as artifacts for PR review

---

### Task 4.3: Final verification

**Step 1: Run all tests**

Run: `task all`
Expected: lint passes, all tests pass (unit + integration), build succeeds

**Step 2: Run integration tests specifically**

Run: `task test:integration`
Expected: All 7+ integration tests pass

**Step 3: Verify golden file update workflow**

Run: `task test:golden:update`
Run: `git diff internal/tui/testdata/`
Expected: No diff (golden files already up to date)

**Step 4: Run VHS locally (if Docker available)**

Run: `task vhs`
Run: `task vhs:check`
Expected: No diff in ASCII golden files

**Step 5: Final commit (if any cleanup needed)**

```bash
git add -A
git commit -m "test(tui): finalize integration testing setup"
```

**Acceptance Criteria:**
- `task all` passes cleanly
- `task test:integration` passes all integration tests
- `task vhs` generates recordings without error
- `task vhs:check` reports no diffs
- No lint warnings
