# Full API Parity Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bring the CLI and MCP server to full CRUD parity with the Fastmail JMAP, CardDAV, and CalDAV APIs.

**Architecture:** Five phases building bottom-up — JMAP types/builders → service layer → CLI commands + MCP tools in lock-step. Each phase ships independently.

**Tech Stack:** Go 1.21+, Cobra CLI, JMAP (RFC 8620/8621), CardDAV/CalDAV, MCP JSON-RPC 2.0, testify, httptest

**Design doc:** `docs/plans/2026-02-06-full-api-parity-design.md`

---

## Phase 1 — Mailbox + Mail Read

### Task 1: MailboxSetBuilder (JMAP types)

**Files:**
- Modify: `internal/jmap/mailbox.go` (after line 65, after `MailboxGetResponse`)
- Test: `internal/jmap/mailbox_test.go`

**Step 1: Write the failing tests**

Add to `internal/jmap/mailbox_test.go`:

```go
func TestMailboxSetBuilder_Create(t *testing.T) {
	args := NewMailboxSet("account-1").
		Create("mb1", map[string]any{"name": "Projects", "parentId": nil}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	create := args["create"].(map[string]map[string]any)
	assert.Equal(t, "Projects", create["mb1"]["name"])
}

func TestMailboxSetBuilder_Update(t *testing.T) {
	args := NewMailboxSet("account-1").
		Update("mb-123", map[string]any{"name": "Renamed"}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	update := args["update"].(map[string]map[string]any)
	assert.Equal(t, "Renamed", update["mb-123"]["name"])
}

func TestMailboxSetBuilder_Destroy(t *testing.T) {
	args := NewMailboxSet("account-1").
		Destroy("mb-123", "mb-456").
		Build()

	assert.Equal(t, []string{"mb-123", "mb-456"}, args["destroy"])
}

func TestMailboxSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "s1",
		"newState": "s2",
		"created": {"mb1": {"id": "Mab999"}},
		"updated": {"mb-123": null},
		"destroyed": ["mb-456"],
		"notCreated": {},
		"notUpdated": {},
		"notDestroyed": {}
	}`

	var resp MailboxSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "s2", resp.NewState)
	require.Contains(t, resp.Created, "mb1")
	assert.Equal(t, "Mab999", resp.Created["mb1"].ID)
	assert.Equal(t, []string{"mb-456"}, resp.Destroyed)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run 'TestMailboxSet' ./internal/jmap/`
Expected: FAIL — `NewMailboxSet` and types undefined

**Step 3: Implement MailboxSetBuilder**

Add to `internal/jmap/mailbox.go` after `MailboxGetResponse`:

```go
// MailboxSetBuilder builds arguments for Mailbox/set.
type MailboxSetBuilder struct {
	accountID string
	create    map[string]map[string]any
	update    map[string]map[string]any
	destroy   []string
}

// NewMailboxSet creates a new Mailbox/set builder.
func NewMailboxSet(accountID string) *MailboxSetBuilder {
	return &MailboxSetBuilder{
		accountID: accountID,
		create:    make(map[string]map[string]any),
		update:    make(map[string]map[string]any),
	}
}

// Create adds a mailbox to be created.
func (b *MailboxSetBuilder) Create(clientID string, mailbox map[string]any) *MailboxSetBuilder {
	b.create[clientID] = mailbox
	return b
}

// Update adds a mailbox to be updated.
func (b *MailboxSetBuilder) Update(mailboxID string, patch map[string]any) *MailboxSetBuilder {
	b.update[mailboxID] = patch
	return b
}

// Destroy adds mailbox IDs to be destroyed.
func (b *MailboxSetBuilder) Destroy(ids ...string) *MailboxSetBuilder {
	b.destroy = append(b.destroy, ids...)
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *MailboxSetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}
	if len(b.create) > 0 {
		args["create"] = b.create
	}
	if len(b.update) > 0 {
		args["update"] = b.update
	}
	if len(b.destroy) > 0 {
		args["destroy"] = b.destroy
	}
	return args
}

// MailboxSetResponse represents the response from Mailbox/set.
type MailboxSetResponse struct {
	AccountID    string                 `json:"accountId"`
	OldState     string                 `json:"oldState"`
	NewState     string                 `json:"newState"`
	Created      map[string]Mailbox     `json:"created"`
	Updated      map[string]any         `json:"updated"`
	Destroyed    []string               `json:"destroyed"`
	NotCreated   map[string]MethodError `json:"notCreated"`
	NotUpdated   map[string]MethodError `json:"notUpdated"`
	NotDestroyed map[string]MethodError `json:"notDestroyed"`
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -run 'TestMailboxSet' ./internal/jmap/`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/jmap/mailbox.go internal/jmap/mailbox_test.go
git commit -m "feat(jmap): add MailboxSetBuilder for Mailbox/set operations"
```

---

### Task 2: MailboxService

**Files:**
- Create: `pkg/fastmail/mailbox_service.go`
- Test: `pkg/fastmail/mailbox_service_test.go`
- Modify: `pkg/fastmail/client.go` (add `Mailbox()` accessor)

**Step 1: Write the failing tests**

Create `pkg/fastmail/mailbox_service_test.go`:

```go
package fastmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

func TestMailboxService_List(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Session response
			json.NewEncoder(w).Encode(jmap.Session{
				APIURL:          "http://" + r.Host + "/api",
				PrimaryAccounts: map[string]string{jmap.CapMail: "A1"},
			})
			return
		}
		// Mailbox/get response
		resp := map[string]any{
			"methodResponses": []any{
				[]any{"Mailbox/get", map[string]any{
					"accountId": "A1",
					"state":     "s1",
					"list": []map[string]any{
						{"id": "mb-1", "name": "Inbox", "role": "inbox", "totalEmails": 150, "unreadEmails": 10},
						{"id": "mb-2", "name": "Sent", "role": "sent", "totalEmails": 50, "unreadEmails": 0},
					},
					"notFound": []string{},
				}, "0"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	require.NoError(t, client.Connect(context.Background()))

	mailboxes, err := client.Mailbox().List(context.Background())
	require.NoError(t, err)
	require.Len(t, mailboxes, 2)

	assert.Equal(t, "mb-1", mailboxes[0].ID)
	assert.Equal(t, "Inbox", mailboxes[0].Name)
	assert.Equal(t, MailboxRole("inbox"), mailboxes[0].Role)
	assert.Equal(t, uint64(150), mailboxes[0].TotalEmails)
	assert.Equal(t, uint64(10), mailboxes[0].UnreadEmails)
}

func TestMailboxService_Create(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(jmap.Session{
				APIURL:          "http://" + r.Host + "/api",
				PrimaryAccounts: map[string]string{jmap.CapMail: "A1"},
			})
			return
		}
		resp := map[string]any{
			"methodResponses": []any{
				[]any{"Mailbox/set", map[string]any{
					"accountId": "A1",
					"oldState":  "s1",
					"newState":  "s2",
					"created":   map[string]any{"mb1": map[string]any{"id": "Mab999"}},
				}, "0"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	require.NoError(t, client.Connect(context.Background()))

	mb, err := client.Mailbox().Create(context.Background(), "Projects", "")
	require.NoError(t, err)
	assert.Equal(t, "Mab999", mb.ID)
}

func TestMailboxService_Rename(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(jmap.Session{
				APIURL:          "http://" + r.Host + "/api",
				PrimaryAccounts: map[string]string{jmap.CapMail: "A1"},
			})
			return
		}
		resp := map[string]any{
			"methodResponses": []any{
				[]any{"Mailbox/set", map[string]any{
					"accountId": "A1",
					"oldState":  "s1",
					"newState":  "s2",
					"updated":   map[string]any{"mb-123": nil},
				}, "0"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	require.NoError(t, client.Connect(context.Background()))

	err := client.Mailbox().Rename(context.Background(), "mb-123", "NewName")
	require.NoError(t, err)
}

func TestMailboxService_Delete(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(jmap.Session{
				APIURL:          "http://" + r.Host + "/api",
				PrimaryAccounts: map[string]string{jmap.CapMail: "A1"},
			})
			return
		}
		resp := map[string]any{
			"methodResponses": []any{
				[]any{"Mailbox/set", map[string]any{
					"accountId": "A1",
					"oldState":  "s1",
					"newState":  "s2",
					"destroyed": []string{"mb-123"},
				}, "0"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	require.NoError(t, client.Connect(context.Background()))

	err := client.Mailbox().Delete(context.Background(), "mb-123")
	require.NoError(t, err)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run 'TestMailboxService' ./pkg/fastmail/`
Expected: FAIL — `client.Mailbox()` undefined

**Step 3: Implement MailboxService**

Create `pkg/fastmail/mailbox_service.go`:

```go
package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// MailboxService provides mailbox operations.
type MailboxService struct {
	client *Client
}

// List returns all mailboxes.
func (s *MailboxService) List(ctx context.Context) ([]Mailbox, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewMailboxGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("mailbox get failed: %s", result.Error())
	}

	var getResp jmap.MailboxGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	return convertMailboxes(getResp.List), nil
}

// Create creates a new mailbox with the given name and optional parent ID.
func (s *MailboxService) Create(ctx context.Context, name, parentID string) (*Mailbox, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	data := map[string]any{"name": name}
	if parentID != "" {
		data["parentId"] = parentID
	}

	setBuilder := jmap.NewMailboxSet(accountID).Create("mb1", data)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("mailbox set failed: %s", result.Error())
	}

	var setResp jmap.MailboxSetResponse
	if err := result.Decode(&setResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotCreated["mb1"]; ok {
		return nil, oops.Errorf("failed to create mailbox: %s", errInfo.Error())
	}

	created, ok := setResp.Created["mb1"]
	if !ok {
		return nil, oops.Errorf("mailbox not returned in created map")
	}

	mb := convertMailbox(created)
	mb.Name = name
	return &mb, nil
}

// Rename changes the name of a mailbox.
func (s *MailboxService) Rename(ctx context.Context, id, newName string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewMailboxSet(accountID).
		Update(id, map[string]any{"name": newName})

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("mailbox set failed: %s", result.Error())
	}

	var setResp jmap.MailboxSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to rename mailbox: %s", errInfo.Error())
	}

	return nil
}

// Delete destroys a mailbox by ID.
func (s *MailboxService) Delete(ctx context.Context, id string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewMailboxSet(accountID).Destroy(id)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("mailbox set failed: %s", result.Error())
	}

	var setResp jmap.MailboxSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete mailbox: %s", errInfo.Error())
	}

	return nil
}

func convertMailboxes(jmapMailboxes []jmap.Mailbox) []Mailbox {
	result := make([]Mailbox, len(jmapMailboxes))
	for i, mb := range jmapMailboxes {
		result[i] = convertMailbox(mb)
	}
	return result
}

func convertMailbox(mb jmap.Mailbox) Mailbox {
	return Mailbox{
		ID:            mb.ID,
		Name:          mb.Name,
		Role:          MailboxRole(mb.Role),
		ParentID:      mb.ParentID,
		TotalEmails:   mb.TotalEmails,
		UnreadEmails:  mb.UnreadEmails,
		TotalThreads:  mb.TotalThreads,
		UnreadThreads: mb.UnreadThreads,
	}
}
```

**Step 4: Add `Mailbox()` accessor to Client**

Add to `pkg/fastmail/client.go` after the `MaskedEmail()` method:

```go
// Mailbox returns the mailbox service for folder operations.
func (c *Client) Mailbox() *MailboxService {
	return &MailboxService{client: c}
}
```

**Step 5: Run tests to verify they pass**

Run: `go test -run 'TestMailboxService' ./pkg/fastmail/`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/fastmail/mailbox_service.go pkg/fastmail/mailbox_service_test.go pkg/fastmail/client.go
git commit -m "feat(fastmail): add MailboxService with List/Create/Rename/Delete"
```

---

### Task 3: Mailbox CLI commands

**Files:**
- Create: `cli/mailbox.go`
- Test: `cli/mailbox_test.go`
- Modify: `cli/root.go` (add `newMailboxCommand()`)

**Step 1: Write the failing tests**

Create `cli/mailbox_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestMailboxHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox --help should not error: %v", err)
	}

	output := buf.String()
	for _, sub := range []string{"list", "create", "rename", "delete"} {
		if !strings.Contains(output, sub) {
			t.Errorf("expected %q subcommand in help, got: %q", sub, output)
		}
	}
}

func TestMailboxCreate_RequiresName(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox create without --name should error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailboxRename_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "rename", "--name", "NewName"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox rename without ID should error")
	}
	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailboxDelete_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox delete without ID should error")
	}
}

func TestMailboxDelete_NoForceShowsWarning(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "delete", "mb-123"})

	// Without --force, should show confirmation message (not error)
	_ = cmd.Execute()
	output := buf.String()
	if !strings.Contains(output, "force") {
		t.Errorf("expected force confirmation in output, got: %q", output)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run 'TestMailbox' ./cli/`
Expected: FAIL — `mailbox` command not registered

**Step 3: Implement mailbox CLI**

Create `cli/mailbox.go` following the pattern from `cli/contacts.go`:

```go
package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newMailboxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailbox",
		Short: "Mailbox operations",
		Long:  "Commands for managing email mailboxes/folders.",
	}

	cmd.AddCommand(newMailboxListCommand())
	cmd.AddCommand(newMailboxCreateCommand())
	cmd.AddCommand(newMailboxRenameCommand())
	cmd.AddCommand(newMailboxDeleteCommand())

	return cmd
}

func newMailboxListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mailboxes",
		Long:  "List all email mailboxes/folders.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			mailboxes, err := client.Mailbox().List(ctx)
			if err != nil {
				return fmt.Errorf("listing mailboxes: %w", err)
			}

			return outputMailboxes(cmd, mailboxes)
		},
	}
	return cmd
}

func newMailboxCreateCommand() *cobra.Command {
	var name string
	var parentID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a mailbox",
		Long:  "Create a new mailbox/folder.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			mb, err := client.Mailbox().Create(ctx, name, parentID)
			if err != nil {
				return fmt.Errorf("creating mailbox: %w", err)
			}

			return outputMailboxCreated(cmd, mb)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "folder name (required)")
	cmd.Flags().StringVar(&parentID, "parent", "", "parent mailbox ID (optional)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newMailboxRenameCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "rename ID",
		Short: "Rename a mailbox",
		Long:  "Rename an existing mailbox/folder.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mailbox().Rename(ctx, id, name); err != nil {
				return fmt.Errorf("renaming mailbox: %w", err)
			}

			return outputMailboxStatus(cmd, id, "Renamed")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new folder name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newMailboxDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a mailbox",
		Long:  "Delete a mailbox/folder. Refuses to delete well-known roles (inbox, drafts, sent, trash, junk).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !force {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete mailbox %s? Use --force to confirm.\n", id)
				return nil
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mailbox().Delete(ctx, id); err != nil {
				return fmt.Errorf("deleting mailbox: %w", err)
			}

			return outputMailboxStatus(cmd, id, "Deleted")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}

// Output helpers — follow pattern from cli/contacts.go

func outputMailboxes(cmd *cobra.Command, mailboxes []fastmail.Mailbox) error {
	if IsQuiet() {
		return nil
	}
	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(mailboxes)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-14s %-20s %-10s %7s %7s\n", "ID", "Name", "Role", "Unread", "Total")
	for _, mb := range mailboxes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-14s %-20s %-10s %7d %7d\n",
			mb.ID, mb.Name, mb.Role, mb.UnreadEmails, mb.TotalEmails)
	}
	return nil
}

func outputMailboxCreated(cmd *cobra.Command, mb *fastmail.Mailbox) error {
	if IsQuiet() {
		return nil
	}
	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(mb)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s (%s)\n", mb.Name, mb.ID)
	return nil
}

func outputMailboxStatus(cmd *cobra.Command, id, status string) error {
	if IsQuiet() {
		return nil
	}
	if IsJSONOutput() {
		result := map[string]string{"id": id, "status": status}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", status, id)
	return nil
}
```

Note: The `outputMailboxes` and related functions need `"github.com/seanb4t/fastmail-cli/pkg/fastmail"` import added.

**Step 4: Register in root command**

In `cli/root.go`, add `cmd.AddCommand(newMailboxCommand())` alongside the existing `newMailCommand()` call.

**Step 5: Run tests to verify they pass**

Run: `go test -run 'TestMailbox' ./cli/`
Expected: PASS

**Step 6: Run full test suite**

Run: `go test -race ./...`
Expected: PASS

**Step 7: Commit**

```bash
git add cli/mailbox.go cli/mailbox_test.go cli/root.go
git commit -m "feat(cli): add mailbox list/create/rename/delete commands"
```

---

### Task 4: Mailbox MCP tools

**Files:**
- Create: `mcp/tools_mailbox.go`
- Test: `mcp/tools_mailbox_test.go`
- Modify: `mcp/tools_mail.go` (add `registerMailboxTools` call in `RegisterMailTools`)

**Step 1: Write the failing tests**

Create `mcp/tools_mailbox_test.go` following the pattern from `mcp/tools_mail_test.go`. Test that tools register and handlers validate required args.

**Step 2: Implement mailbox MCP tools**

Create `mcp/tools_mailbox.go`:

```go
package mcp

import (
	"context"

	"github.com/samber/oops"
)

func registerMailboxTools(s *Server, cfg ToolsConfig) {
	s.RegisterTool(
		NewTool("mailbox_list", "List all email mailboxes/folders"),
		makeMailboxListHandler(cfg),
	)

	s.RegisterTool(
		NewTool("mailbox_create", "Create a new mailbox/folder").
			WithProperty("name", "string", "Folder name").
			WithProperty("parent_id", "string", "Parent mailbox ID (optional)").
			WithRequired("name"),
		makeMailboxCreateHandler(cfg),
	)

	s.RegisterTool(
		NewTool("mailbox_rename", "Rename a mailbox/folder").
			WithProperty("id", "string", "Mailbox ID").
			WithProperty("name", "string", "New folder name").
			WithRequired("id", "name"),
		makeMailboxRenameHandler(cfg),
	)

	s.RegisterTool(
		NewTool("mailbox_delete", "Delete a mailbox/folder").
			WithProperty("id", "string", "Mailbox ID").
			WithRequired("id"),
		makeMailboxDeleteHandler(cfg),
	)
}

func makeMailboxListHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		mailboxes, err := cfg.Client.Mailbox().List(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "listing mailboxes")
		}
		result := make([]map[string]any, len(mailboxes))
		for i, mb := range mailboxes {
			result[i] = map[string]any{
				"id": mb.ID, "name": mb.Name, "role": string(mb.Role),
				"parent_id": mb.ParentID, "total_emails": mb.TotalEmails,
				"unread_emails": mb.UnreadEmails,
			}
		}
		return result, nil
	}
}

func makeMailboxCreateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		name := getStringArg(args, "name", "")
		if name == "" {
			return nil, oops.Errorf("name is required")
		}
		parentID := getStringArg(args, "parent_id", "")

		mb, err := cfg.Client.Mailbox().Create(ctx, name, parentID)
		if err != nil {
			return nil, oops.Wrapf(err, "creating mailbox")
		}
		return map[string]any{"id": mb.ID, "name": mb.Name, "status": "created"}, nil
	}
}

func makeMailboxRenameHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}
		name := getStringArg(args, "name", "")
		if name == "" {
			return nil, oops.Errorf("name is required")
		}

		if err := cfg.Client.Mailbox().Rename(ctx, id, name); err != nil {
			return nil, oops.Wrapf(err, "renaming mailbox")
		}
		return map[string]any{"id": id, "name": name, "status": "renamed"}, nil
	}
}

func makeMailboxDeleteHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Client.Mailbox().Delete(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "deleting mailbox")
		}
		return map[string]any{"id": id, "status": "deleted"}, nil
	}
}
```

**Step 3: Wire into RegisterMailTools**

In `mcp/tools_mail.go`, add `registerMailboxTools(s, cfg)` to `RegisterMailTools()`:

```go
func RegisterMailTools(s *Server, cfg ToolsConfig) {
	registerMailTools(s, cfg)
	registerMaskedEmailTools(s, cfg)
	registerContactTools(s, cfg)
	registerCalendarTools(s, cfg)
	registerMailboxTools(s, cfg)  // NEW
}
```

**Step 4: Run tests**

Run: `go test -race ./mcp/`
Expected: PASS

**Step 5: Commit**

```bash
git add mcp/tools_mailbox.go mcp/tools_mailbox_test.go mcp/tools_mail.go
git commit -m "feat(mcp): add mailbox_list/create/rename/delete tools"
```

---

### Task 5: MailService.GetFull — expanded email properties

**Files:**
- Modify: `pkg/fastmail/mail.go` (add `GetFull` method)
- Modify: `pkg/fastmail/email.go` (add `Attachments` field)
- Test: `pkg/fastmail/mail_test.go`

**Step 1: Write the failing test**

Add to `pkg/fastmail/mail_test.go`:

```go
func TestMailService_GetFull(t *testing.T) {
	// Setup httptest server returning session + Email/get with full properties
	// Verify that From, To, Cc, Bcc, Body, Attachments are populated
}
```

**Step 2: Implement GetFull**

Add `Attachments` field to `pkg/fastmail/email.go`:

```go
type Attachment struct {
	Name   string
	Type   string
	Size   uint64
	BlobID string
}
```

Add `GetFull` to `pkg/fastmail/mail.go` — same as `Get` but requests additional properties:
`"from", "to", "cc", "bcc", "bodyValues", "textBody", "htmlBody", "attachments"` and sets `fetchTextBodyValues: true`.

Convert the additional JMAP fields into domain Email fields (From, To, Body, Attachments).

**Step 3: Run tests, verify pass**

Run: `go test -run 'TestMailService_GetFull' ./pkg/fastmail/`

**Step 4: Commit**

```bash
git add pkg/fastmail/mail.go pkg/fastmail/email.go pkg/fastmail/mail_test.go
git commit -m "feat(fastmail): add GetFull for expanded email properties"
```

---

### Task 6: MailService.GetRaw — blob download

**Files:**
- Modify: `internal/jmap/client.go` (add `DownloadBlob` method)
- Test: `internal/jmap/client_test.go`
- Modify: `pkg/fastmail/mail.go` (add `GetRaw` method)
- Test: `pkg/fastmail/mail_test.go`

**Step 1: Write the failing test for DownloadBlob**

The JMAP session has a `downloadUrl` template like `https://api.fastmail.com/jmap/download/{accountId}/{blobId}/{name}?type={type}`. Test that `DownloadBlob` builds the correct URL and returns an `io.ReadCloser`.

**Step 2: Implement DownloadBlob on jmap.Client**

```go
func (c *Client) DownloadBlob(ctx context.Context, accountID, blobID string) (io.ReadCloser, error) {
	session, err := c.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	url := strings.ReplaceAll(session.DownloadURL, "{accountId}", accountID)
	url = strings.ReplaceAll(url, "{blobId}", blobID)
	url = strings.ReplaceAll(url, "{name}", "raw")
	url = strings.ReplaceAll(url, "{type}", "application/octet-stream")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading blob: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode}
	}
	return resp.Body, nil
}
```

**Step 3: Implement GetRaw on MailService**

```go
func (s *MailService) GetRaw(ctx context.Context, id string) (io.ReadCloser, error) {
	// Get email to find blobId
	// Call s.client.jmap.DownloadBlob(ctx, accountID, email.BlobID)
}
```

Note: `jmap.Email` already has a `BlobID` field. Need to request it in the properties.

**Step 4: Run tests, commit**

```bash
git add internal/jmap/client.go internal/jmap/client_test.go pkg/fastmail/mail.go pkg/fastmail/mail_test.go
git commit -m "feat(jmap): add DownloadBlob and MailService.GetRaw for raw email retrieval"
```

---

### Task 7: `mail show` CLI command

**Files:**
- Modify: `cli/mail.go` (add `newMailShowCommand`)
- Test: `cli/mail_test.go`

**Step 1: Write the failing test**

```go
func TestMailShow_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "show"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail show without ID should error")
	}
}
```

**Step 2: Implement `mail show`**

Add `newMailShowCommand` to `cli/mail.go`:
- Flags: `--full`, `--headers`, `--raw`
- Default: calls `GetFull`, displays From/To/Date/Subject/Body/Attachments
- `--raw`: calls `GetRaw`, pipes `io.Reader` to `cmd.OutOrStdout()`
- Register in `newMailCommand`: `cmd.AddCommand(newMailShowCommand())`

**Step 3: Run tests, commit**

```bash
git add cli/mail.go cli/mail_test.go
git commit -m "feat(cli): add mail show command with --full/--headers/--raw flags"
```

---

### Task 8: Search query parser

**Files:**
- Create: `internal/search/parser.go`
- Test: `internal/search/parser_test.go`

**Step 1: Write the failing tests**

```go
func TestParse_FreeText(t *testing.T) {
	filter := Parse("quarterly report")
	assert.Equal(t, map[string]any{"text": "quarterly report"}, filter)
}

func TestParse_FromToken(t *testing.T) {
	filter := Parse("from:alice")
	assert.Equal(t, map[string]any{"from": "alice"}, filter)
}

func TestParse_CompoundQuery(t *testing.T) {
	filter := Parse("from:alice subject:meeting is:unread")
	assert.Equal(t, map[string]any{
		"operator":   "AND",
		"conditions": []map[string]any{
			{"from": "alice"},
			{"subject": "meeting"},
			{"notKeyword": "$seen"},
		},
	}, filter)
}

func TestParse_HasAttachment(t *testing.T) {
	filter := Parse("has:attachment")
	assert.Equal(t, map[string]any{"hasAttachment": true}, filter)
}

func TestParse_IsUnread(t *testing.T) {
	filter := Parse("is:unread")
	assert.Equal(t, map[string]any{"notKeyword": "$seen"}, filter)
}

func TestParse_IsFlagged(t *testing.T) {
	filter := Parse("is:flagged")
	assert.Equal(t, map[string]any{"hasKeyword": "$flagged"}, filter)
}

func TestParse_DateFilters(t *testing.T) {
	filter := Parse("after:2026-01-01 before:2026-02-01")
	// Returns AND condition with after/before
}
```

**Step 2: Implement parser**

Create `internal/search/parser.go` — tokenizes the query string, maps tokens to JMAP FilterCondition fields per the design doc table. Single conditions return a flat map; multiple conditions wrap in `{"operator": "AND", "conditions": [...]}`.

**Step 3: Run tests, commit**

```bash
git add internal/search/parser.go internal/search/parser_test.go
git commit -m "feat(search): add query parser for JMAP Email/query filter conditions"
```

---

### Task 9: Refactor MailService.Search for structured filters

**Files:**
- Modify: `pkg/fastmail/mail.go` (refactor `Search` or add `SearchWithFilter`)
- Test: `pkg/fastmail/mail_test.go`

**Step 1: Write the failing test**

Test that `SearchWithFilter` accepts a `map[string]any` filter and passes it to Email/query.

**Step 2: Implement**

Add `SearchWithFilter(ctx context.Context, filter map[string]any, limit uint64) ([]Email, error)` to `MailService`. Same as existing `Search` but accepts structured filter instead of building `{"text": query}`.

**Step 3: Run tests, commit**

```bash
git add pkg/fastmail/mail.go pkg/fastmail/mail_test.go
git commit -m "feat(fastmail): add SearchWithFilter for structured JMAP query filters"
```

---

### Task 10: `mail search` CLI command

**Files:**
- Modify: `cli/mail.go` (add `newMailSearchCommand`)
- Test: `cli/mail_test.go`

**Step 1: Write the failing test**

```go
func TestMailSearch_RequiresQuery(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "search"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail search without query should error")
	}
}
```

**Step 2: Implement**

Add `newMailSearchCommand` — takes positional arg(s) as query, calls `search.Parse()`, then `client.Mail().SearchWithFilter()`. Register in `newMailCommand`.

**Step 3: Run tests, commit**

```bash
git add cli/mail.go cli/mail_test.go
git commit -m "feat(cli): add mail search command with query parser syntax"
```

---

## Phase 2 — Mail Write Operations

### Task 11: `mail move` CLI command

**Files:**
- Modify: `cli/mail.go` (add `newMailMoveCommand`)
- Test: `cli/mail_test.go`

**Step 1: Write tests for flag validation**

Test that `--folder` is required and ID arg is required.

**Step 2: Implement**

Wraps existing `client.Mail().Move(ctx, id, folder)`. Register in `newMailCommand`.

**Step 3: Commit**

```bash
git add cli/mail.go cli/mail_test.go
git commit -m "feat(cli): add mail move command"
```

---

### Task 12: `mail delete` CLI command

**Files:**
- Modify: `cli/mail.go` (add `newMailDeleteCommand`)
- Test: `cli/mail_test.go`

**Step 1: Write tests**

Test `--force` flag behavior (same pattern as `contacts delete`).

**Step 2: Implement**

Wraps existing `client.Mail().Delete(ctx, id)`. Without `--force`, shows confirmation prompt.

**Step 3: Commit**

```bash
git add cli/mail.go cli/mail_test.go
git commit -m "feat(cli): add mail delete command with --force flag"
```

---

### Task 13: MailService.SetKeywords + `mail flag` CLI + MCP

**Files:**
- Modify: `pkg/fastmail/mail.go` (add `SetKeywords`)
- Test: `pkg/fastmail/mail_test.go`
- Modify: `cli/mail.go` (add `newMailFlagCommand`)
- Test: `cli/mail_test.go`
- Modify: `mcp/tools_mail.go` (add `mail_flag` tool)

**Step 1: Write test for SetKeywords**

Test that it calls `Email/set` with keyword patches like `{"keywords/$seen": true}`.

**Step 2: Implement SetKeywords**

Uses existing `EmailSetBuilder.Update()` with keyword patch map.

**Step 3: Implement CLI command**

`mail flag <ID> --read/--unread --flagged/--unflagged` — validates at least one flag pair provided.

**Step 4: Add MCP tool**

Register `mail_flag` in `registerMailTools`.

**Step 5: Commit**

```bash
git add pkg/fastmail/mail.go pkg/fastmail/mail_test.go cli/mail.go cli/mail_test.go mcp/tools_mail.go
git commit -m "feat: add mail flag command for setting email keywords"
```

---

## Phase 3 — Calendar CLI

### Task 14: Calendar CLI commands

**Files:**
- Create: `cli/calendar.go`
- Test: `cli/calendar_test.go`
- Modify: `cli/root.go` (add `newCalendarCommand()`)
- Modify: `internal/config/config.go` (add CalDAV config fields if not present)

**Step 1: Write failing tests**

Test help output shows all 6 subcommands (list, show, create, update, delete, calendars).

**Step 2: Implement**

Create `cli/calendar.go` following the `cli/contacts.go` pattern for DAV client creation:

```go
func createCalendarClient() (*fastmail.CalendarService, error) {
	// Follow createContactsClient() pattern:
	// Load config, get token, derive CalDAV endpoint, create dav.Client
}
```

Config needs `CalDAVEndpoint` and `CalDAVUsername` (or reuse CardDAV username). Add `caldav_endpoint` and `caldav_username` to `internal/config/config.go` if not present, with default `https://caldav.fastmail.com/dav/`.

Implement all 6 subcommands:
- `calendar list` — `--from`, `--to`, `--calendar` flags
- `calendar show <ID>`
- `calendar create` — `--summary`, `--start`, `--end`, `--location`, `--description`, `--calendar`, `--all-day`
- `calendar update <ID>` — same flags, all optional
- `calendar delete <ID>` — `--force`
- `calendar calendars` — list calendar containers

**Step 3: Run tests, commit**

```bash
git add cli/calendar.go cli/calendar_test.go cli/root.go internal/config/config.go
git commit -m "feat(cli): add calendar commands (list/show/create/update/delete/calendars)"
```

---

### Task 15: Refactor cli/mcp.go to use RegisterMailTools + wire all tools

**Files:**
- Modify: `cli/mcp.go`
- Test: `cli/mcp_test.go`

**Step 1: Write test**

Verify that the MCP server registers all expected tools (mail, masked-email, contacts, calendar, mailbox).

**Step 2: Implement**

Replace the ~180 lines of duplicated inline registration in `cli/mcp.go` with a call to `mcp.RegisterMailTools()`:

```go
func runMCPServer(_ *cobra.Command) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ... signal handling ...

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connecting to Fastmail: %w", err)
	}

	// Create optional DAV clients
	contactsClient, _ := createContactsClient()  // nil if not configured
	calendarAdapter := createCalendarAdapter()    // nil if not configured

	server := mcp.NewServer("fastmail-cli", Version)

	mcp.RegisterMailTools(server, mcp.ToolsConfig{
		Client:   client,
		Contacts: contactsClient,
		Calendar: calendarAdapter,
	})

	return server.Run(ctx, os.Stdin, os.Stdout)
}
```

**Step 3: Run tests**

Run: `go test -race ./cli/ ./mcp/`

**Step 4: Commit**

```bash
git add cli/mcp.go cli/mcp_test.go
git commit -m "refactor(cli): replace inline MCP registration with RegisterMailTools"
```

---

### Task 16: New calendar MCP tools (update, delete, calendars)

**Files:**
- Modify: `mcp/tools_mail.go` (add `calendar_update`, `calendar_delete`, `calendar_list` tools in `registerCalendarTools`)
- Modify: `mcp/tools_mail.go` (extend `CalendarAdapter` with new function pointers)
- Test: `mcp/tools_mail_test.go`

**Step 1: Extend CalendarAdapter**

```go
type CalendarAdapter struct {
	ListEventsFunc    func(ctx context.Context, calendarID string, start, end time.Time) ([]fastmail.Event, error)
	GetEventFunc      func(ctx context.Context, eventID string) (*fastmail.Event, error)
	CreateEventFunc   func(ctx context.Context, event *fastmail.Event) error
	UpdateEventFunc   func(ctx context.Context, event *fastmail.Event) error
	DeleteEventFunc   func(ctx context.Context, eventID string) error
	ListCalendarsFunc func(ctx context.Context) ([]fastmail.Calendar, error)
}
```

**Step 2: Register new tools**

Add `calendar_get`, `calendar_update`, `calendar_delete`, `calendar_list` tools.

**Step 3: Run tests, commit**

```bash
git add mcp/tools_mail.go mcp/tools_mail_test.go
git commit -m "feat(mcp): add calendar_get/update/delete/list tools"
```

---

## Phase 4 — Identity, Vacation, Thread

### Task 17: Fix Identity.ReplyTo/BCC types

**Files:**
- Modify: `internal/jmap/submission.go` (lines 177-178)
- Test: `internal/jmap/submission_test.go` (if exists, or create)

**Step 1: Write test**

Test that Identity with ReplyTo as `[]EmailAddress` deserializes correctly.

**Step 2: Fix types**

Change in `internal/jmap/submission.go`:

```go
// Before:
ReplyTo       string `json:"replyTo,omitempty"`
BCC           string `json:"bcc,omitempty"`

// After:
ReplyTo       []EmailAddress `json:"replyTo,omitempty"`
BCC           []EmailAddress `json:"bcc,omitempty"`
```

**Step 3: Check for compilation breakage**

Run: `go build ./...`

The only use of `Identity.ReplyTo` is in `pkg/fastmail/mail.go:buildEmailForReply` where it reads `original.ReplyTo` — which is `jmap.Email.ReplyTo` (already `[]EmailAddress`), not `jmap.Identity.ReplyTo`. So this change should be safe.

**Step 4: Run tests, commit**

```bash
git add internal/jmap/submission.go
git commit -m "fix(jmap): correct Identity.ReplyTo and BCC types to []EmailAddress per RFC 8621"
```

---

### Task 18: IdentityService + CLI + MCP

**Files:**
- Create: `pkg/fastmail/identity_service.go`
- Test: `pkg/fastmail/identity_service_test.go`
- Modify: `pkg/fastmail/client.go` (add `Identity()` accessor)
- Create: `cli/identity.go`
- Test: `cli/identity_test.go`
- Modify: `cli/root.go` (register identity command)
- Modify: `mcp/tools_mail.go` (add `identity_list` tool in `RegisterMailTools`)

**Step 1: Implement IdentityService**

Reuses existing `IdentityGetBuilder` from `internal/jmap/submission.go`. Single method: `List(ctx) ([]Identity, error)`.

Domain type needed in `pkg/fastmail/`:

```go
type Identity struct {
	ID            string
	Name          string
	Email         string
	ReplyTo       []EmailAddress
	BCC           []EmailAddress
	TextSignature string
	HTMLSignature string
	MayDelete     bool
}
```

**Step 2: Implement CLI**

`cli/identity.go` with single `identity list` command. Text output: table with ID, Name, Email, MayDelete.

**Step 3: Implement MCP tool**

Add `identity_list` tool in a new `registerIdentityTools` function, wired into `RegisterMailTools`.

**Step 4: Run tests, commit**

```bash
git add pkg/fastmail/identity_service.go pkg/fastmail/identity_service_test.go pkg/fastmail/client.go \
  cli/identity.go cli/identity_test.go cli/root.go mcp/tools_mail.go
git commit -m "feat: add identity list command, service, and MCP tool"
```

---

### Task 19: VacationResponse types + builder

**Files:**
- Modify: `internal/jmap/session.go` (add `CapVacationResponse` constant)
- Create: `internal/jmap/vacation.go`
- Test: `internal/jmap/vacation_test.go`

**Step 1: Add capability constant**

Add to `internal/jmap/session.go`:

```go
CapVacationResponse = "urn:ietf:params:jmap:vacationresponse"
```

**Step 2: Create types and builders**

Create `internal/jmap/vacation.go`:

```go
package jmap

type VacationResponse struct {
	ID           string `json:"id"`
	IsEnabled    bool   `json:"isEnabled"`
	FromDate     string `json:"fromDate,omitempty"`
	ToDate       string `json:"toDate,omitempty"`
	Subject      string `json:"subject,omitempty"`
	TextBody     string `json:"textBody,omitempty"`
	HTMLBody     string `json:"htmlBody,omitempty"`
}

type VacationGetBuilder struct { accountID string }
type VacationSetBuilder struct {
	accountID string
	update    map[string]any
}

// Builders follow same pattern as MailboxGetBuilder / EmailSetBuilder
```

**Step 3: Run tests, commit**

```bash
git add internal/jmap/session.go internal/jmap/vacation.go internal/jmap/vacation_test.go
git commit -m "feat(jmap): add VacationResponse types, builders, and CapVacationResponse"
```

---

### Task 20: VacationService + CLI + MCP

**Files:**
- Create: `pkg/fastmail/vacation_service.go`
- Test: `pkg/fastmail/vacation_service_test.go`
- Modify: `pkg/fastmail/client.go` (add `Vacation()` accessor)
- Create: `cli/vacation.go`
- Test: `cli/vacation_test.go`
- Modify: `cli/root.go` (register vacation command)
- Modify: `mcp/tools_mail.go` (add vacation tools)

**Step 1: Implement VacationService**

Methods: `Get(ctx) (*Vacation, error)` and `Set(ctx, opts SetVacationOptions) error`.

VacationResponse/get is a singleton — use `ids: null` and get the single item.

**Step 2: Implement CLI**

`vacation show` and `vacation set` commands with flags per design doc.

**Step 3: Implement MCP tools**

`vacation_get` and `vacation_set` tools.

**Step 4: Run tests, commit**

```bash
git add pkg/fastmail/vacation_service.go pkg/fastmail/vacation_service_test.go pkg/fastmail/client.go \
  cli/vacation.go cli/vacation_test.go cli/root.go mcp/tools_mail.go
git commit -m "feat: add vacation show/set command, service, and MCP tools"
```

---

### Task 21: Thread types + builder

**Files:**
- Create: `internal/jmap/thread.go`
- Test: `internal/jmap/thread_test.go`

**Step 1: Create types and builder**

```go
package jmap

type Thread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

type ThreadGetBuilder struct {
	accountID string
	ids       []string
}

type ThreadGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Thread `json:"list"`
	NotFound  []string `json:"notFound"`
}
```

**Step 2: Run tests, commit**

```bash
git add internal/jmap/thread.go internal/jmap/thread_test.go
git commit -m "feat(jmap): add Thread types and ThreadGetBuilder"
```

---

### Task 22: ThreadService + CLI + MCP

**Files:**
- Create: `pkg/fastmail/thread_service.go`
- Test: `pkg/fastmail/thread_service_test.go`
- Modify: `pkg/fastmail/client.go` (add `Thread()` accessor)
- Create: `cli/thread.go`
- Test: `cli/thread_test.go`
- Modify: `cli/root.go` (register thread command)
- Modify: `mcp/tools_mail.go` (add `thread_get` tool)

**Step 1: Implement ThreadService**

`Get(ctx, id) ([]Email, error)` — two JMAP calls: `Thread/get` for `emailIds`, then `Email/get` for summaries.

**Step 2: Implement CLI**

`thread show <ID>` — displays numbered list of emails in thread.

**Step 3: Implement MCP tool**

`thread_get` tool.

**Step 4: Run tests, commit**

```bash
git add pkg/fastmail/thread_service.go pkg/fastmail/thread_service_test.go pkg/fastmail/client.go \
  cli/thread.go cli/thread_test.go cli/root.go mcp/tools_mail.go
git commit -m "feat: add thread show command, service, and MCP tool"
```

---

## Phase 5 — MCP Catch-up

### Task 23: contacts_update and contacts_delete MCP tools

**Files:**
- Modify: `mcp/tools_mail.go` (add tools in `registerContactTools`)
- Test: `mcp/tools_mail_test.go`

**Step 1: Add tools**

`contacts_update` and `contacts_delete` tools — delegate to `cfg.Contacts.Update()` and `cfg.Contacts.Delete()`.

**Step 2: Run tests, commit**

```bash
git add mcp/tools_mail.go mcp/tools_mail_test.go
git commit -m "feat(mcp): add contacts_update and contacts_delete tools"
```

---

### Task 24: New MCP resources

**Files:**
- Modify: `mcp/resources.go` (add new resources)
- Test: `mcp/resources_test.go`

**Step 1: Add resources**

Per design doc, add:
- `fastmail://mailboxes` — returns mailbox list
- `fastmail://identities` — returns identity list
- `fastmail://vacation` — returns vacation status
- `fastmail://calendars` — returns calendar list

Follow existing resource registration pattern.

**Step 2: Run tests, commit**

```bash
git add mcp/resources.go mcp/resources_test.go
git commit -m "feat(mcp): add mailboxes, identities, vacation, and calendars resources"
```

---

### Task 25: Final integration test + docs update

**Step 1: Run full test suite**

Run: `go test -race -cover ./...`

Verify all tests pass and coverage is reasonable.

**Step 2: Run linter**

Run: `task lint`

Fix any issues.

**Step 3: Update CLI docs**

Update `docs/site/cli/` pages to cover new commands. Minimum: add `mailbox.md`, `calendar.md`, `identity.md`, `vacation.md`, `thread.md`.

**Step 4: Commit**

```bash
git add docs/
git commit -m "docs: add CLI reference pages for new commands"
```

---

## Summary

| Phase | Tasks | New Commands | New MCP Tools |
|-------|-------|-------------|---------------|
| 1 — Mailbox + Mail Read | 1-10 | mailbox list/create/rename/delete, mail show, mail search | mailbox_list/create/rename/delete |
| 2 — Mail Write | 11-13 | mail move, mail delete, mail flag | mail_flag |
| 3 — Calendar CLI | 14-16 | calendar list/show/create/update/delete/calendars | calendar_get/update/delete/list |
| 4 — Identity/Vacation/Thread | 17-22 | identity list, vacation show/set, thread show | identity_list, vacation_get/set, thread_get |
| 5 — MCP Catch-up | 23-25 | — | contacts_update/delete + 4 resources |

**Total: 25 tasks, 19 new CLI commands, 15 new MCP tools, 4 new MCP resources.**
