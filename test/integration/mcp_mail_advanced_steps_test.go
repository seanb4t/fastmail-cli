//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPMailAdvancedSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I call the "([^"]*)" advanced mail tool$`, iCallAdvancedMailToolNoArgs)
	sc.Step(`^I call the "([^"]*)" advanced mail tool with:$`, iCallAdvancedMailToolWith)
}

func iCallAdvancedMailToolNoArgs(ctx context.Context, toolName string) (context.Context, error) {
	return callAdvancedMailTool(ctx, toolName, map[string]any{})
}

func iCallAdvancedMailToolWith(ctx context.Context, toolName string, table *godog.Table) (context.Context, error) {
	args := tableToArgs(table)
	return callAdvancedMailTool(ctx, toolName, args)
}

// callAdvancedMailTool invokes an MCP mail tool using the session server URL.
// For tools that need blob operations (download, import, upload), the session
// server must already be reconfigured to point downloadUrl/uploadUrl at the
// blob server (done by setupBlobServer in the Given step).
func callAdvancedMailTool(ctx context.Context, toolName string, args map[string]any) (context.Context, error) {
	w := WorldFromContext(ctx)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{Client: client}
	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	result, err := handler(ctx, args)
	w.ToolResult = result
	w.ToolError = err
	return ctx, nil
}
