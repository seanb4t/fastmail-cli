//go:build integration

package integration

// MockMaskedEmail holds masked email test data for mock server responses.
type MockMaskedEmail struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	State         string `json:"state"`
	ForDomain     string `json:"forDomain"`
	Description   string `json:"description"`
	CreatedBy     string `json:"createdBy"`
	CreatedAt     string `json:"createdAt"`
	LastMessageAt string `json:"lastMessageAt"`
}

// maskedEmailDomainData holds typed state for masked email scenarios.
type maskedEmailDomainData struct {
	MaskedEmails []MockMaskedEmail
}

func init() {
	RegisterMethodHandler("MaskedEmail/get", handleMaskedEmailGet)
	RegisterMethodHandler("MaskedEmail/set", handleMaskedEmailSet)
}

func getMaskedEmailData(w *World) *maskedEmailDomainData {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	data, ok := w.DomainData["masked-email"].(*maskedEmailDomainData)
	if !ok {
		data = &maskedEmailDomainData{}
		w.DomainData["masked-email"] = data
	}
	return data
}

func handleMaskedEmailGet(w *World, args map[string]any) map[string]any {
	data := getMaskedEmailData(w)

	// Check if specific IDs were requested
	var requestedIDs map[string]bool
	if ids, ok := args["ids"].([]any); ok {
		requestedIDs = make(map[string]bool, len(ids))
		for _, id := range ids {
			requestedIDs[id.(string)] = true
		}
	}

	var list []map[string]any
	for _, me := range data.MaskedEmails {
		if requestedIDs != nil && !requestedIDs[me.ID] {
			continue
		}
		list = append(list, maskedEmailToMap(me))
	}

	if list == nil {
		list = []map[string]any{}
	}

	return map[string]any{
		"accountId": "acc1",
		"state":     "me1",
		"list":      list,
		"notFound":  []any{},
	}
}

//nolint:gocognit,gocyclo // test helper with set operation dispatch
func handleMaskedEmailSet(w *World, args map[string]any) map[string]any {
	resp := map[string]any{
		"accountId": "acc1",
		"oldState":  "me1",
		"newState":  "me2",
	}

	// Handle create
	if create, ok := args["create"].(map[string]any); ok {
		created := make(map[string]any)
		for clientID, rawProps := range create {
			props, _ := rawProps.(map[string]any)
			forDomain, _ := props["forDomain"].(string)
			description, _ := props["description"].(string)

			created[clientID] = map[string]any{
				"id":          "masked-created-" + clientID,
				"email":       "generated-" + clientID + "@fastmail.com",
				"state":       "enabled",
				"forDomain":   forDomain,
				"description": description,
				"createdBy":   "test",
				"createdAt":   "2024-01-15T10:00:00Z",
			}
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

func maskedEmailToMap(me MockMaskedEmail) map[string]any {
	state := me.State
	if state == "" {
		state = "enabled"
	}
	createdAt := me.CreatedAt
	if createdAt == "" {
		createdAt = "2024-01-15T10:00:00Z"
	}

	return map[string]any{
		"id":            me.ID,
		"email":         me.Email,
		"state":         state,
		"forDomain":     me.ForDomain,
		"description":   me.Description,
		"createdBy":     me.CreatedBy,
		"createdAt":     createdAt,
		"lastMessageAt": me.LastMessageAt,
	}
}
