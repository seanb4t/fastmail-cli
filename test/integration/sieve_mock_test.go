//go:build integration

package integration

func init() {
	RegisterMethodHandler("SieveScript/get", handleSieveScriptGet)
	RegisterMethodHandler("SieveScript/set", handleSieveScriptSet)
	RegisterMethodHandler("SieveScript/validate", handleSieveScriptValidate)
}

func handleSieveScriptGet(w *World, args map[string]any) map[string]any {
	scripts := getSieveScripts(w)

	// Filter by IDs if specified
	if ids, ok := args["ids"].([]any); ok {
		idSet := make(map[string]bool, len(ids))
		for _, id := range ids {
			idSet[id.(string)] = true
		}
		var filtered []MockSieveScript
		for _, s := range scripts {
			if idSet[s.ID] {
				filtered = append(filtered, s)
			}
		}
		scripts = filtered
	}

	list := make([]map[string]any, len(scripts))
	for i, s := range scripts {
		list[i] = map[string]any{
			"id":       s.ID,
			"name":     s.Name,
			"blobId":   "blob-" + s.ID,
			"script":   s.Script,
			"isActive": s.IsActive,
		}
	}

	return map[string]any{
		"accountId": "acc1",
		"state":     "sieve1",
		"list":      list,
		"notFound":  []any{},
	}
}

func handleSieveScriptSet(_ *World, args map[string]any) map[string]any {
	resp := map[string]any{
		"accountId": "acc1",
		"oldState":  "sieve1",
		"newState":  "sieve2",
	}

	// Handle create
	if create, ok := args["create"].(map[string]any); ok {
		created := make(map[string]any)
		for clientID, data := range create {
			props := map[string]any{
				"id": "sieve-created-" + clientID,
			}
			if m, ok := data.(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					props["name"] = name
				}
				if script, ok := m["script"].(string); ok {
					props["script"] = script
				}
				if isActive, ok := m["isActive"].(bool); ok {
					props["isActive"] = isActive
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

func handleSieveScriptValidate(_ *World, _ map[string]any) map[string]any {
	// Return a successful validation (no error)
	return map[string]any{
		"accountId": "acc1",
	}
}

// MockSieveScript holds sieve script test data.
type MockSieveScript struct {
	ID       string
	Name     string
	Script   string
	IsActive bool
}

// getSieveScripts retrieves mock sieve scripts from the World's DomainData.
func getSieveScripts(w *World) []MockSieveScript {
	if w.DomainData == nil {
		return nil
	}
	scripts, _ := w.DomainData["sieve_scripts"].([]MockSieveScript)
	return scripts
}
