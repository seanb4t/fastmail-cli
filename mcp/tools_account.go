package mcp

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// registerAccountTools registers account-related tools on the server.
func registerAccountTools(s *Server, cfg ToolsConfig) {
	// quota_get - Get storage quota
	s.RegisterTool(
		NewTool("quota_get", "Get storage quota usage for the account"),
		makeQuotaGetHandler(cfg),
	)
}

func makeQuotaGetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		if cfg.Client == nil {
			return nil, oops.Errorf("client not configured")
		}

		quota, err := cfg.Client.Quota().Get(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "getting quota")
		}

		return map[string]any{
			"used":          quota.Used,
			"limit":         quota.Limit,
			"used_percent":  quota.UsedPercent,
			"used_display":  fastmail.FormatSize(quota.Used),
			"limit_display": fastmail.FormatSize(quota.Limit),
		}, nil
	}
}
