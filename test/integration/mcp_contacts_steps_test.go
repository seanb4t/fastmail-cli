//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPContactsSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I call the "([^"]*)" contacts tool$`, iCallContactsToolNoArgs)
	sc.Step(`^I call the "([^"]*)" contacts tool with:$`, iCallContactsToolWith)
}

func iCallContactsToolNoArgs(ctx context.Context, toolName string) (context.Context, error) {
	return callContactsTool(ctx, toolName, map[string]any{})
}

func iCallContactsToolWith(ctx context.Context, toolName string, table *godog.Table) (context.Context, error) {
	args := tableToArgs(table)
	return callContactsTool(ctx, toolName, args)
}

func callContactsTool(ctx context.Context, toolName string, args map[string]any) (context.Context, error) {
	w := WorldFromContext(ctx)

	// Get current contact data from the scenario (may be empty for error-case scenarios).
	data := getContactData(w)

	// Create mock CardDAV server with current contact data.
	server := newMockCardDAVServer(data.Contacts)
	defer server.Close()

	// Create ContactsClient pointing to mock server.
	contactsClient := fastmail.NewContactsClient(server.URL, "test", "test-token")

	// Create JMAP client too (needed for ToolsConfig).
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{
		Client:   client,
		Contacts: contactsClient,
	}
	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	result, err := handler(ctx, args)
	w.ToolResult = result
	w.ToolError = err
	return ctx, nil
}
