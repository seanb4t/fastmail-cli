package mcp

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// registerIdentityTools registers identity management tools on the server.
func registerIdentityTools(s *Server, cfg ToolsConfig) {
	// identity_list - List all sender identities
	s.RegisterTool(
		NewTool("identity_list", "List all sender identities"),
		makeIdentityListHandler(cfg),
	)

	// identity_set - Update a sender identity
	s.RegisterTool(
		NewTool("identity_set", "Update a sender identity").
			WithProperty("id", "string", "The identity ID").
			WithProperty("name", "string", "Display name (optional)").
			WithProperty("reply_to", "string", "Reply-to email address (optional)").
			WithProperty("text_signature", "string", "Text signature (optional)").
			WithProperty("html_signature", "string", "HTML signature (optional)").
			WithRequired("id"),
		makeIdentitySetHandler(cfg),
	)
}

func makeIdentityListHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		identities, err := cfg.Client.Identity().List(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "listing identities")
		}

		result := make([]map[string]any, len(identities))
		for i, id := range identities {
			entry := map[string]any{
				"id":             id.ID,
				"name":           id.Name,
				"email":          id.Email,
				"text_signature": id.TextSignature,
				"html_signature": id.HTMLSignature,
				"may_delete":     id.MayDelete,
			}
			if len(id.ReplyTo) > 0 {
				replyTo := make([]map[string]string, len(id.ReplyTo))
				for j, addr := range id.ReplyTo {
					replyTo[j] = map[string]string{"name": addr.Name, "email": addr.Email}
				}
				entry["reply_to"] = replyTo
			}
			if len(id.Bcc) > 0 {
				bcc := make([]map[string]string, len(id.Bcc))
				for j, addr := range id.Bcc {
					bcc[j] = map[string]string{"name": addr.Name, "email": addr.Email}
				}
				entry["bcc"] = bcc
			}
			result[i] = entry
		}
		return result, nil
	}
}

func makeIdentitySetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		opts := fastmail.UpdateIdentityOptions{}
		hasChange := false

		if name := getStringArg(args, "name", ""); name != "" {
			opts.Name = &name
			hasChange = true
		}
		if replyTo := getStringArg(args, "reply_to", ""); replyTo != "" {
			opts.ReplyTo = []fastmail.EmailAddress{{Email: replyTo}}
			hasChange = true
		}
		if sig := getStringArg(args, "text_signature", ""); sig != "" {
			opts.TextSignature = &sig
			hasChange = true
		}
		if htmlSig := getStringArg(args, "html_signature", ""); htmlSig != "" {
			opts.HTMLSignature = &htmlSig
			hasChange = true
		}

		if !hasChange {
			return nil, oops.Errorf("at least one of name, reply_to, text_signature, or html_signature must be specified")
		}

		if err := cfg.Client.Identity().Update(ctx, id, opts); err != nil {
			return nil, oops.Wrapf(err, "updating identity")
		}

		return map[string]any{
			"id":     id,
			"status": "updated",
		}, nil
	}
}
