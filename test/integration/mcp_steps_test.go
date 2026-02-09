//go:build integration

package integration

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPSteps(sc *godog.ScenarioContext) {
	// Background
	sc.Step(`^a connected MCP tools config$`, aConnectedMCPToolsConfig)

	// When steps
	sc.Step(`^I call the "([^"]*)" tool$`, iCallToolNoArgs)
	sc.Step(`^I call the "([^"]*)" tool with:$`, iCallToolWith)
	sc.Step(`^I call the "([^"]*)" tool with keywords:$`, iCallToolWithKeywords)

	// Then steps
	sc.Step(`^the tool call should succeed$`, theToolCallShouldSucceed)
	sc.Step(`^the tool call should fail$`, theToolCallShouldFail)
	sc.Step(`^the tool error should contain "([^"]*)"$`, theToolErrorShouldContain)
	sc.Step(`^the tool result should contain (\d+) items?$`, theToolResultShouldContainNItems)
	sc.Step(`^the tool result field "([^"]*)" should equal "([^"]*)"$`, theToolResultFieldShouldEqual)
	sc.Step(`^the tool result field "([^"]*)" should be true$`, theToolResultFieldShouldBeTrue)
	sc.Step(`^the tool result field "([^"]*)" should be false$`, theToolResultFieldShouldBeFalse)
	sc.Step(`^the tool result item (\d+) field "([^"]*)" should equal "([^"]*)"$`, theToolResultItemFieldShouldEqual)
}

func aConnectedMCPToolsConfig(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w == nil {
		return ctx, fmt.Errorf("world not initialized")
	}
	return ctx, nil
}

func iCallToolNoArgs(ctx context.Context, toolName string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{Client: client}
	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	result, err := handler(ctx, map[string]any{})
	w.ToolResult = result
	w.ToolError = err
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

func theToolCallShouldFail(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ToolError == nil {
		return fmt.Errorf("expected tool call to fail, but it succeeded")
	}
	return nil
}

func theToolErrorShouldContain(ctx context.Context, substring string) error {
	w := WorldFromContext(ctx)
	if w.ToolError == nil {
		return fmt.Errorf("expected an error containing %q, but no error occurred", substring)
	}
	if !strings.Contains(w.ToolError.Error(), substring) {
		return fmt.Errorf("expected error containing %q, got: %s", substring, w.ToolError.Error())
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
	case []map[string]any:
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

func theToolResultFieldShouldEqual(ctx context.Context, field, expected string) error {
	w := WorldFromContext(ctx)
	result, ok := w.ToolResult.(map[string]any)
	if !ok {
		return fmt.Errorf("tool result is not a map, got %T", w.ToolResult)
	}
	val, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in tool result (keys: %v)", field, mapKeys(result))
	}
	actual := fmt.Sprintf("%v", val)
	if actual != expected {
		return fmt.Errorf("field %q: expected %q, got %q", field, expected, actual)
	}
	return nil
}

func theToolResultFieldShouldBeTrue(ctx context.Context, field string) error {
	w := WorldFromContext(ctx)
	result, ok := w.ToolResult.(map[string]any)
	if !ok {
		return fmt.Errorf("tool result is not a map, got %T", w.ToolResult)
	}
	val, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in tool result", field)
	}
	b, ok := val.(bool)
	if !ok || !b {
		return fmt.Errorf("field %q: expected true, got %v", field, val)
	}
	return nil
}

func theToolResultFieldShouldBeFalse(ctx context.Context, field string) error {
	w := WorldFromContext(ctx)
	result, ok := w.ToolResult.(map[string]any)
	if !ok {
		return fmt.Errorf("tool result is not a map, got %T", w.ToolResult)
	}
	val, exists := result[field]
	if !exists {
		return fmt.Errorf("field %q not found in tool result", field)
	}
	b, ok := val.(bool)
	if !ok || b {
		return fmt.Errorf("field %q: expected false, got %v", field, val)
	}
	return nil
}

func theToolResultItemFieldShouldEqual(ctx context.Context, index int, field, expected string) error {
	w := WorldFromContext(ctx)
	items, err := toolResultAsSlice(w.ToolResult)
	if err != nil {
		return err
	}
	if index < 1 || index > len(items) {
		return fmt.Errorf("item index %d out of range (have %d items)", index, len(items))
	}
	item := items[index-1]
	val, exists := item[field]
	if !exists {
		return fmt.Errorf("field %q not found in item %d (keys: %v)", field, index, mapKeys(item))
	}
	actual := fmt.Sprintf("%v", val)
	if actual != expected {
		return fmt.Errorf("item %d field %q: expected %q, got %q", index, field, expected, actual)
	}
	return nil
}

func toolResultAsSlice(result any) ([]map[string]any, error) {
	switch r := result.(type) {
	case []map[string]any:
		return r, nil
	case []any:
		items := make([]map[string]any, len(r))
		for i, item := range r {
			m, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("item %d is not a map, got %T", i, item)
			}
			items[i] = m
		}
		return items, nil
	default:
		return nil, fmt.Errorf("result is not a slice, got %T", result)
	}
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
