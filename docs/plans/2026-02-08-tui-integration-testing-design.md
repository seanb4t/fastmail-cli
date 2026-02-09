# TUI Integration Testing Design

**Goal:** Add two complementary integration testing layers — programmatic teatest tests and declarative VHS visual tests — to catch layout regressions, verify user workflows, and generate demo GIFs.

**Architecture:** Teatest drives the bubbletea Model with simulated input and compares rendered output against ASCII golden files. VHS drives the compiled binary against a WireMock mock server and produces both GIFs (for docs) and ASCII captures (for CI golden comparison). Both layers use canned JMAP data and run deterministically in CI.

**Tech Stack:** charmbracelet/x/exp/teatest, charmbracelet/vhs, WireMock (Docker), httptest (for teatest mocks)

---

## Layer 1: teatest Integration Tests

### File Structure

```
internal/tui/
├── testfixture_test.go       # Shared test helper: mock server + model setup
├── integration_test.go       # All integration test cases
├── testdata/
│   ├── jmap_fixtures.json    # Canned JMAP responses (5 mailboxes, 10 emails)
│   ├── DashboardLoad.golden
│   ├── SelectMailbox_JK.golden
│   ├── SelectMailbox_Arrows.golden
│   ├── ReadEmail.golden
│   ├── ComposeEmail.golden
│   ├── ToggleSidebar.golden
│   └── NarrowTerminal.golden
```

### Test Fixture

A shared helper creates a fully wired Model with mock HTTP, reusing the existing `WithHTTPClient` functional option pattern.

```go
func newTestFixture(t *testing.T) (*teatest.TestModel, *httptest.Server) {
    srv := newMockJMAPServer(t)
    client := fastmail.NewClient(srv.URL, "test-token",
        fastmail.WithHTTPClient(srv.Client()))
    model := NewModel(client)
    tm := teatest.NewTestModel(t, model,
        teatest.WithInitialTermSize(120, 40),
    )
    t.Cleanup(func() { srv.Close() })
    return tm, srv
}
```

### Color Profile

Set `lipgloss.SetColorProfile(termenv.Ascii)` in `TestMain` so golden files contain only plain text. The `LIPGLOSS_COLOR_PROFILE=ascii` env var provides the same behavior in CI.

### Mock JMAP Server

A `newMockJMAPServer` function reads canned responses from `testdata/jmap_fixtures.json` and serves them via `httptest.NewServer`. It handles:

- `GET /.well-known/jmap` — session discovery with capabilities
- `POST /jmap` — multiplexed method calls, matched by method name in request body:
  - `Mailbox/get` — 5 mailboxes (Inbox, Drafts, Sent, Archive, Trash)
  - `Email/query` — email IDs for the selected mailbox
  - `Email/get` — 10 emails with realistic fields (flags, dates, subjects, senders, preview)

### Test Cases

| Test | Terminal | Keys | Asserts |
|------|----------|------|---------|
| `DashboardLoad` | 120x40 | (wait for load) | Stats bar visible, sidebar rendered, inbox auto-selected, email list populated |
| `SelectMailbox_JK` | 120x40 | Tab → j → j → Enter | Email list updates to selected mailbox, stats bar updates |
| `SelectMailbox_Arrows` | 120x40 | Tab → ↓ → ↓ → Enter | Same assertions as JK variant (arrow key parity) |
| `ReadEmail` | 120x40 | Enter on first email | Full reader view rendered with subject, sender, body |
| `ComposeEmail` | 120x40 | `c` | Compose overlay appears with To/Subject/Body fields |
| `ToggleSidebar` | 120x40 | `b` | Sidebar hides, main pane expands to full width |
| `NarrowTerminal` | 60x20 | (wait for load) | Sidebar auto-collapsed, key bar abbreviated |

### Golden File Workflow

- **Regenerate:** `go test ./internal/tui/ -run Integration -update`
- **Compare:** `go test ./internal/tui/ -run Integration` (fails on diff)
- **Review:** `git diff internal/tui/testdata/` to confirm changes are intentional

---

## Layer 2: VHS Visual Tests

### File Structure

```
docs/vhs/
├── dashboard.tape              # Launch → browse mailboxes → read email
├── compose.tape                # Open compose → fill fields
├── navigation.tape             # Sidebar toggle, pane cycling, split adjust
├── wiremock/
│   ├── mappings/
│   │   ├── session.json        # GET /.well-known/jmap stub
│   │   ├── mailbox-get.json    # POST /jmap → Mailbox/get stub
│   │   ├── email-query.json    # POST /jmap → Email/query stub
│   │   └── email-get.json      # POST /jmap → Email/get stub
│   └── __files/
│       ├── session.json        # Session response body
│       ├── mailboxes.json      # 5 mailboxes response body
│       └── emails.json         # 10 emails response body
├── testdata/
│   ├── dashboard.ascii         # ASCII golden (CI comparison)
│   ├── compose.ascii
│   └── navigation.ascii
├── dashboard.gif               # Generated demo GIF (gitignored)
├── compose.gif
└── navigation.gif
```

### WireMock

WireMock runs as a Docker container serving canned JMAP responses. Stub mappings use `bodyPatterns` to match JMAP method names in POST requests.

Example mapping (`mappings/mailbox-get.json`):
```json
{
  "request": {
    "method": "POST",
    "url": "/jmap",
    "bodyPatterns": [{ "contains": "Mailbox/get" }]
  },
  "response": {
    "status": 200,
    "bodyFileName": "mailboxes.json",
    "headers": { "Content-Type": "application/json" }
  }
}
```

### Tape File Pattern

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

# Browse mailboxes
Type "jj"
Sleep 500ms
Enter
Sleep 2s

# Read an email
Type "j"
Sleep 300ms
Enter
Sleep 2s
```

### VHS Tape Inventory

| Tape | Workflow | Duration |
|------|----------|----------|
| `dashboard.tape` | Launch → browse mailboxes → select → read email | ~10s |
| `compose.tape` | Launch → press `c` → fill compose fields | ~8s |
| `navigation.tape` | Sidebar toggle (`b`), pane cycling (Tab), split adjust (+/-) | ~8s |

---

## CI Integration

### Existing Workflow (extend)

teatest integration tests run as part of `task test` — no new workflow needed:

```yaml
# .github/workflows/test.yml (existing, add env var)
- name: Tests
  env:
    LIPGLOSS_COLOR_PROFILE: ascii
  run: task test
```

### New VHS Workflow (PR only)

```yaml
# .github/workflows/vhs.yml
name: VHS Visual Tests
on: [pull_request]
jobs:
  vhs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: go.mod }
      - run: task build
      - name: Start WireMock
        run: task vhs:mock:start
      - uses: charmbracelet/vhs-action@v2
        with:
          path: "docs/vhs/*.tape"
      - name: Stop WireMock
        run: task vhs:mock:stop
      - name: Compare golden ASCII
        run: task vhs:check
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: vhs-recordings
          path: docs/vhs/*.gif
```

GIFs are uploaded as artifacts so PR reviewers can visually inspect.

---

## Taskfile Additions

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
    - docker rm -f fm-mock

vhs:
  desc: Generate all VHS recordings
  deps: [build, vhs:mock:start]
  cmds:
    - defer: task vhs:mock:stop
    - for: [dashboard, compose, navigation]
      cmd: vhs docs/vhs/{{.ITEM}}.tape

vhs:check:
  desc: Verify VHS golden files match
  cmds:
    - for: [dashboard, compose, navigation]
      cmd: diff docs/vhs/testdata/{{.ITEM}}.ascii docs/vhs/expected/{{.ITEM}}.ascii

vhs:update:
  desc: Regenerate VHS golden files (copy testdata to expected)
  cmds:
    - cp docs/vhs/testdata/*.ascii docs/vhs/expected/
```

---

## Developer Workflow

1. Make UI changes
2. `task test:integration` — see what golden files broke
3. `task test:golden:update` — regenerate golden files
4. `git diff internal/tui/testdata/` — review diffs, confirm intentional
5. `task vhs` — regenerate VHS recordings (optional, local Docker required)
6. Commit updated golden files alongside code changes

---

## .gitattributes

```
*.golden -text
docs/vhs/*.gif binary
```

## .gitignore (additions)

```
docs/vhs/*.gif
```

GIFs are generated artifacts — committed only when intentionally updating README demos.
