package mcp

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// registerFilterTools registers sieve filter tools on the server.
func registerFilterTools(s *Server, cfg ToolsConfig) {
	// filter_list - List filter scripts
	s.RegisterTool(
		NewTool("filter_list", "List all Sieve filter scripts"),
		makeFilterListHandler(cfg),
	)

	// filter_get - Get filter script by ID
	s.RegisterTool(
		NewTool("filter_get", "Get a Sieve filter script by ID with its content").
			WithProperty("id", "string", "The filter script ID").
			WithRequired("id"),
		makeFilterGetHandler(cfg),
	)

	// filter_create - Create filter script
	s.RegisterTool(
		NewTool("filter_create", "Create a new Sieve filter script").
			WithProperty("name", "string", "Name for the filter script").
			WithProperty("script", "string", "Sieve script content").
			WithProperty("activate", "boolean", "Activate the script on creation (default: false)").
			WithRequired("name", "script"),
		makeFilterCreateHandler(cfg),
	)

	// filter_activate - Activate a filter script
	s.RegisterTool(
		NewTool("filter_activate", "Activate a Sieve filter script").
			WithProperty("id", "string", "The filter script ID").
			WithRequired("id"),
		makeFilterActivateHandler(cfg),
	)

	// filter_deactivate - Deactivate a filter script
	s.RegisterTool(
		NewTool("filter_deactivate", "Deactivate a Sieve filter script").
			WithProperty("id", "string", "The filter script ID").
			WithRequired("id"),
		makeFilterDeactivateHandler(cfg),
	)

	// filter_validate - Validate a sieve script
	s.RegisterTool(
		NewTool("filter_validate", "Validate Sieve script syntax without storing it").
			WithProperty("script", "string", "Sieve script content to validate").
			WithRequired("script"),
		makeFilterValidateHandler(cfg),
	)

	// filter_delete - Delete a filter script
	s.RegisterTool(
		NewTool("filter_delete", "Delete a Sieve filter script by ID").
			WithProperty("id", "string", "The filter script ID").
			WithRequired("id"),
		makeFilterDeleteHandler(cfg),
	)
}

func makeFilterListHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		scripts, err := cfg.Client.Sieve().List(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "listing filter scripts")
		}

		result := make([]map[string]any, len(scripts))
		for i, s := range scripts {
			result[i] = map[string]any{
				"id":         s.ID,
				"name":       s.Name,
				"is_active":  s.IsActive,
				"blob_id":    s.BlobID,
				"created_at": s.CreatedAt,
				"updated_at": s.UpdatedAt,
			}
		}
		return result, nil
	}
}

func makeFilterGetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		script, err := cfg.Client.Sieve().Get(ctx, id)
		if err != nil {
			return nil, oops.Wrapf(err, "getting filter script")
		}

		return map[string]any{
			"id":         script.ID,
			"name":       script.Name,
			"is_active":  script.IsActive,
			"blob_id":    script.BlobID,
			"script":     script.Script,
			"created_at": script.CreatedAt,
			"updated_at": script.UpdatedAt,
		}, nil
	}
}

func makeFilterCreateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		name := getStringArg(args, "name", "")
		if name == "" {
			return nil, oops.Errorf("name is required")
		}
		script := getStringArg(args, "script", "")
		if script == "" {
			return nil, oops.Errorf("script is required")
		}
		activate := getBoolArg(args, "activate", false)

		opts := fastmail.CreateSieveScriptOptions{
			Name:     name,
			Script:   script,
			Activate: activate,
		}

		created, err := cfg.Client.Sieve().Create(ctx, opts)
		if err != nil {
			return nil, oops.Wrapf(err, "creating filter script")
		}

		return map[string]any{
			"id":        created.ID,
			"name":      created.Name,
			"is_active": created.IsActive,
			"status":    "created",
		}, nil
	}
}

func makeFilterActivateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Client.Sieve().Activate(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "activating filter script")
		}

		return map[string]any{
			"id":     id,
			"status": "activated",
		}, nil
	}
}

func makeFilterDeactivateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Client.Sieve().Deactivate(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "deactivating filter script")
		}

		return map[string]any{
			"id":     id,
			"status": "deactivated",
		}, nil
	}
}

func makeFilterValidateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		script := getStringArg(args, "script", "")
		if script == "" {
			return nil, oops.Errorf("script is required")
		}

		result, err := cfg.Client.Sieve().Validate(ctx, script)
		if err != nil {
			return nil, oops.Wrapf(err, "validating filter script")
		}

		response := map[string]any{
			"is_valid": result.IsValid,
		}
		if !result.IsValid {
			response["error_type"] = result.ErrorType
			response["description"] = result.Description
		}
		return response, nil
	}
}

func makeFilterDeleteHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Client.Sieve().Delete(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "deleting filter script")
		}

		return map[string]any{
			"id":     id,
			"status": "deleted",
		}, nil
	}
}
