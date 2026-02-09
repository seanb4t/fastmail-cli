//go:build integration

package integration

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPSteps(sc *godog.ScenarioContext) {
	// Background
	sc.Step(`^a connected MCP tools config$`, aConnectedMCPToolsConfig)

	// When steps
	sc.Step(`^I call the "([^"]*)" tool with:$`, iCallToolWith)
	sc.Step(`^I call the "([^"]*)" tool with keywords:$`, iCallToolWithKeywords)

	// Then steps
	sc.Step(`^the tool call should succeed$`, theToolCallShouldSucceed)
	sc.Step(`^the tool result should contain (\d+) items?$`, theToolResultShouldContainNItems)
}

func aConnectedMCPToolsConfig(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w == nil {
		return ctx, fmt.Errorf("world not initialized")
	}
	return ctx, nil
}

func iCallToolWith(ctx context.Context, toolName string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{Client: client}
	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	args := tableToArgs(table)
	result, err := handler(ctx, args)
	w.ToolResult = result
	w.ToolError = err
	return ctx, nil
}

func iCallToolWithKeywords(ctx context.Context, toolName string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{Client: client}
	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Build args with nested keywords object
	args := make(map[string]any)
	keywords := make(map[string]any)
	for _, row := range table.Rows {
		key := row.Cells[0].Value
		value := row.Cells[1].Value
		if key == "id" {
			args["id"] = value
		} else {
			// It's a keyword like "$flagged"
			b, _ := strconv.ParseBool(value)
			keywords[key] = b
		}
	}
	args["keywords"] = keywords

	result, err := handler(ctx, args)
	w.ToolResult = result
	w.ToolError = err
	return ctx, nil
}

func theToolCallShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ToolError != nil {
		return fmt.Errorf("tool call failed: %w", w.ToolError)
	}
	return nil
}

func theToolResultShouldContainNItems(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	if w.ToolError != nil {
		return fmt.Errorf("tool call failed: %w", w.ToolError)
	}

	switch result := w.ToolResult.(type) {
	case []any:
		if len(result) != count {
			return fmt.Errorf("expected %d items, got %d", count, len(result))
		}
	case []*mcp.Email:
		if len(result) != count {
			return fmt.Errorf("expected %d items, got %d", count, len(result))
		}
	default:
		return fmt.Errorf("unexpected result type: %T", w.ToolResult)
	}
	return nil
}

// resolveToolHandler creates a temporary MCP server, registers tools, and extracts the handler.
func resolveToolHandler(toolName string, cfg mcp.ToolsConfig) mcp.ToolHandler {
	s := mcp.NewServer("test", "0.0.0")
	mcp.RegisterMailTools(s, cfg)
	return s.GetToolHandler(toolName)
}

func tableToArgs(table *godog.Table) map[string]any {
	args := make(map[string]any)
	if len(table.Rows) < 2 {
		return args
	}

	// Detect format: key-value pairs (2 columns named "key"/"value")
	// or header-row format
	if len(table.Rows[0].Cells) == 2 &&
		table.Rows[0].Cells[0].Value == "key" &&
		table.Rows[0].Cells[1].Value == "value" {
		for _, row := range table.Rows[1:] {
			key := row.Cells[0].Value
			value := row.Cells[1].Value

			// Try to parse as integer
			if n, err := strconv.Atoi(value); err == nil {
				args[key] = float64(n)
				continue
			}
			// Try to parse as bool
			if b, err := strconv.ParseBool(value); err == nil {
				args[key] = b
				continue
			}
			args[key] = value
		}
	}
	return args
}
