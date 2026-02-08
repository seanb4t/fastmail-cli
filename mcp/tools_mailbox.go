package mcp

import (
	"context"

	"github.com/samber/oops"
)

func registerMailboxTools(s *Server, cfg ToolsConfig) {
	// mailbox_list - List all mailboxes
	s.RegisterTool(
		NewTool("mailbox_list", "List all email mailboxes/folders"),
		makeMailboxListHandler(cfg),
	)

	// mailbox_create - Create a new mailbox
	s.RegisterTool(
		NewTool("mailbox_create", "Create a new mailbox/folder").
			WithProperty("name", "string", "Folder name").
			WithProperty("parent_id", "string", "Parent mailbox ID (optional)").
			WithRequired("name"),
		makeMailboxCreateHandler(cfg),
	)

	// mailbox_rename - Rename a mailbox
	s.RegisterTool(
		NewTool("mailbox_rename", "Rename a mailbox/folder").
			WithProperty("id", "string", "Mailbox ID").
			WithProperty("name", "string", "New folder name").
			WithRequired("id", "name"),
		makeMailboxRenameHandler(cfg),
	)

	// mailbox_delete - Delete a mailbox
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
				"id":            mb.ID,
				"name":          mb.Name,
				"role":          string(mb.Role),
				"parent_id":     mb.ParentID,
				"total_emails":  mb.TotalEmails,
				"unread_emails": mb.UnreadEmails,
			}
		}

		return result, nil
	}
}

func makeMailboxCreateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		name, err := getStringArg(args, "name", "")
		if err != nil {
			return nil, err
		}
		if name == "" {
			return nil, oops.Errorf("name is required")
		}

		parentID, err := getStringArg(args, "parent_id", "")
		if err != nil {
			return nil, err
		}

		mb, err := cfg.Client.Mailbox().Create(ctx, name, parentID)
		if err != nil {
			return nil, oops.Wrapf(err, "creating mailbox")
		}

		return map[string]any{
			"id":     mb.ID,
			"name":   mb.Name,
			"status": "created",
		}, nil
	}
}

func makeMailboxRenameHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id, err := getStringArg(args, "id", "")
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		name, err := getStringArg(args, "name", "")
		if err != nil {
			return nil, err
		}
		if name == "" {
			return nil, oops.Errorf("name is required")
		}

		if err := cfg.Client.Mailbox().Rename(ctx, id, name); err != nil {
			return nil, oops.Wrapf(err, "renaming mailbox")
		}

		return map[string]any{
			"id":     id,
			"name":   name,
			"status": "renamed",
		}, nil
	}
}

func makeMailboxDeleteHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id, err := getStringArg(args, "id", "")
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Client.Mailbox().Delete(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "deleting mailbox")
		}

		return map[string]any{
			"id":     id,
			"status": "deleted",
		}, nil
	}
}
