//go:build integration

package integration

func init() {
	RegisterMethodHandler("Mailbox/set", handleMailboxSet)
}

func handleMailboxSet(_ *World, args map[string]any) map[string]any {
	resp := map[string]any{
		"accountId": "acc1",
		"oldState":  "m1",
		"newState":  "m2",
	}

	// Handle create
	if create, ok := args["create"].(map[string]any); ok {
		created := make(map[string]any)
		for clientID, data := range create {
			props := map[string]any{
				"id": "mb-created-" + clientID,
			}
			if m, ok := data.(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					props["name"] = name
				}
				if parentID, ok := m["parentId"].(string); ok {
					props["parentId"] = parentID
				}
			}
			created[clientID] = props
		}
		resp["created"] = created
	}

	// Handle update
	if update, ok := args["update"].(map[string]any); ok {
		updated := make(map[string]any)
		for id := range update {
			updated[id] = nil
		}
		resp["updated"] = updated
	}

	// Handle destroy
	if destroy, ok := args["destroy"].([]any); ok {
		resp["destroyed"] = destroy
	}

	return resp
}
