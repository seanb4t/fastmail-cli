package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	subcommands := []string{"list", "create", "rename", "delete"}
	for _, sub := range subcommands {
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

func TestMailboxRename_RequiresName(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "rename", "mb-123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox rename without --name should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
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

	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailboxDelete_NoForceShowsWarning(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "delete", "mb-123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox delete without --force should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Are you sure") {
		t.Errorf("expected confirmation warning, got: %q", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("expected --force hint, got: %q", output)
	}
}

func TestMailboxDelete_ProtectedRole(t *testing.T) {
	// Mock JMAP server that returns mailboxes with roles
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		if methodName == "Mailbox/get" {
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
							{"id": "mb-drafts", "name": "Drafts", "role": "drafts"},
							{"id": "mb-sent", "name": "Sent", "role": "sent"},
							{"id": "mb-trash", "name": "Trash", "role": "trash"},
							{"id": "mb-junk", "name": "Junk", "role": "junk"},
							{"id": "mb-custom", "name": "MyFolder", "role": ""}
						],
						"notFound": []
					}, "0"]
				]
			}`))
			return
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"capabilities": {
				"urn:ietf:params:jmap:core": {},
				"urn:ietf:params:jmap:mail": {}
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
			"apiUrl": "` + apiServer.URL + `",
			"downloadUrl": "https://example.com/download",
			"uploadUrl": "https://example.com/upload",
			"eventSourceUrl": "https://example.com/events",
			"state": "s1"
		}`))
	}))
	defer sessionServer.Close()

	protectedIDs := []struct {
		id   string
		role string
	}{
		{"mb-inbox", "inbox"},
		{"mb-drafts", "drafts"},
		{"mb-sent", "sent"},
		{"mb-trash", "trash"},
		{"mb-junk", "junk"},
	}

	// Create a minimal config file for tests
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(configPath, []byte("endpoint: "+sessionServer.URL+"\n"), 0o600)

	for _, tc := range protectedIDs {
		t.Run(tc.role, func(t *testing.T) {
			t.Setenv("FASTMAIL_TOKEN", "test-token")

			cmd := NewRootCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"mailbox", "delete", tc.id, "--force", "--config", configPath})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error deleting %s mailbox, got nil", tc.role)
			}

			if !strings.Contains(err.Error(), "cannot delete") {
				t.Errorf("expected 'cannot delete' error for %s, got: %v", tc.role, err)
			}
			if !strings.Contains(err.Error(), tc.role) {
				t.Errorf("expected role %q in error message, got: %v", tc.role, err)
			}
		})
	}
}
