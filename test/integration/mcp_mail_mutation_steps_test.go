//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPMailMutationSteps(sc *godog.ScenarioContext) {
	// Custom step that converts comma-separated "to" values into a string slice
	sc.Step(`^I call the "([^"]*)" tool with args:$`, iCallToolWithSliceArgs)
}

// iCallToolWithSliceArgs is like iCallToolWith but converts known array fields
// (to, cc, bcc) from comma-separated strings into []string slices, which is
// required by getStringSliceArg in the tool handlers.
func iCallToolWithSliceArgs(ctx context.Context, toolName string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{Client: client}
	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	args := tableToArgsWithSlices(table)
	result, err := handler(ctx, args)
	w.ToolResult = result
	w.ToolError = err
	return ctx, nil
}

// sliceArgKeys lists arg keys that should be parsed as string slices.
var sliceArgKeys = map[string]bool{
	"to":  true,
	"cc":  true,
	"bcc": true,
}

// tableToArgsWithSlices works like tableToArgs but converts known array fields
// into []string slices. Comma-separated values are split into multiple entries.
func tableToArgsWithSlices(table *godog.Table) map[string]any {
	args := make(map[string]any)
	if len(table.Rows) < 2 {
		return args
	}

	if len(table.Rows[0].Cells) == 2 &&
		table.Rows[0].Cells[0].Value == "key" &&
		table.Rows[0].Cells[1].Value == "value" {
		for _, row := range table.Rows[1:] {
			key := row.Cells[0].Value
			value := row.Cells[1].Value

			if sliceArgKeys[key] {
				parts := strings.Split(value, ",")
				trimmed := make([]string, len(parts))
				for i, p := range parts {
					trimmed[i] = strings.TrimSpace(p)
				}
				args[key] = trimmed
				continue
			}

			args[key] = parseArgValue(value)
		}
	}
	return args
}

// parseArgValue attempts to parse a value as int (float64) or bool, falling back to string.
func parseArgValue(value string) any {
	// Try bool
	switch strings.ToLower(value) {
	case "true":
		return true
	case "false":
		return false
	}
	// Try int
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err == nil {
		return float64(n)
	}
	return value
}
