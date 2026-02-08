package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/oops"
)

// registerVacationTools registers vacation/out-of-office tools on the server.
func registerVacationTools(s *Server, cfg ToolsConfig) {
	// vacation_status - Get current vacation response status
	s.RegisterTool(
		NewTool("vacation_status", "Get the current vacation/out-of-office auto-reply status"),
		makeVacationStatusHandler(cfg),
	)

	// vacation_set - Set vacation response
	s.RegisterTool(
		NewTool("vacation_set", "Enable or disable the vacation/out-of-office auto-reply. When enabled=true, subject and body are required.").
			WithProperty("enabled", "boolean", "Whether the vacation response is enabled").
			WithProperty("subject", "string", "Auto-reply subject line (required if enabled=true)").
			WithProperty("body", "string", "Auto-reply message body (required if enabled=true)").
			WithProperty("from_date", "string", "Start date in RFC3339 format (optional)").
			WithProperty("to_date", "string", "End date in RFC3339 format (optional)").
			WithRequired("enabled"),
		makeVacationSetHandler(cfg),
	)
}

func makeVacationStatusHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		status, err := cfg.Client.Vacation().GetStatus(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "getting vacation status")
		}

		result := map[string]any{
			"is_enabled": status.IsEnabled,
			"subject":    status.Subject,
			"text_body":  status.TextBody,
		}
		if status.FromDate != nil {
			result["from_date"] = status.FromDate.Format(time.RFC3339)
		}
		if status.ToDate != nil {
			result["to_date"] = status.ToDate.Format(time.RFC3339)
		}

		return result, nil
	}
}

func makeVacationSetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		enabled, err := parseVacationEnabled(args)
		if err != nil {
			return nil, err
		}

		if enabled {
			return handleVacationEnable(ctx, cfg, args)
		}

		if err := cfg.Client.Vacation().Disable(ctx); err != nil {
			return nil, oops.Wrapf(err, "disabling vacation response")
		}

		return map[string]any{
			"status":  "disabled",
			"message": "Vacation response disabled",
		}, nil
	}
}

func parseVacationEnabled(args map[string]any) (bool, error) {
	enabledRaw, ok := args["enabled"]
	if !ok {
		return false, oops.Errorf("enabled is required")
	}
	enabled, ok := enabledRaw.(bool)
	if !ok {
		return false, oops.Errorf("enabled must be a boolean")
	}
	return enabled, nil
}

func handleVacationEnable(ctx context.Context, cfg ToolsConfig, args map[string]any) (any, error) {
	subject := getStringArg(args, "subject", "")
	body := getStringArg(args, "body", "")

	if subject == "" || body == "" {
		return nil, oops.Errorf("subject and body are required when enabling vacation response")
	}

	fromTime, toTime, err := parseVacationDateRange(args)
	if err != nil {
		return nil, err
	}

	if err := cfg.Client.Vacation().Enable(ctx, subject, body, fromTime, toTime); err != nil {
		return nil, oops.Wrapf(err, "enabling vacation response")
	}

	return map[string]any{
		"status":  "enabled",
		"message": fmt.Sprintf("Vacation response enabled with subject: %s", subject),
	}, nil
}

func parseVacationDateRange(args map[string]any) (*time.Time, *time.Time, error) {
	var fromTime, toTime *time.Time
	if fromStr := getStringArg(args, "from_date", ""); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return nil, nil, oops.Wrapf(err, "parsing from_date")
		}
		fromTime = &t
	}
	if toStr := getStringArg(args, "to_date", ""); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return nil, nil, oops.Wrapf(err, "parsing to_date")
		}
		toTime = &t
	}

	if fromTime != nil && toTime != nil && fromTime.After(*toTime) {
		return nil, nil, oops.Errorf("from_date must be before to_date")
	}

	return fromTime, toTime, nil
}
