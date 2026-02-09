//go:build integration

package integration

func init() {
	RegisterMethodHandler("Identity/get", handleIdentityGetExt)
	RegisterMethodHandler("Identity/set", handleIdentitySet)
}

// identityData holds mock identity state within DomainData.
type identityData struct {
	Identities []mockIdentity
}

type mockIdentity struct {
	ID            string
	Name          string
	Email         string
	TextSignature string
}

const identityDataKey = "identity"

func getIdentityData(w *World) *identityData {
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	if v, ok := w.DomainData[identityDataKey]; ok {
		return v.(*identityData)
	}
	data := &identityData{}
	w.DomainData[identityDataKey] = data
	return data
}

// handleIdentityGetExt returns identity data from DomainData when configured,
// otherwise falls back to the default hardcoded identityGetResponse for
// scenarios that only need a basic identity (e.g., mail send).
func handleIdentityGetExt(w *World, _ map[string]any) map[string]any {
	// If no identity data has been set up, fall back to the default response
	// so that existing tests (e.g., mail send) continue to work.
	if w.DomainData == nil {
		return identityGetResponse()
	}
	if _, ok := w.DomainData[identityDataKey]; !ok {
		return identityGetResponse()
	}

	data := getIdentityData(w)
	list := make([]map[string]any, len(data.Identities))
	for i, id := range data.Identities {
		list[i] = map[string]any{
			"id":            id.ID,
			"name":          id.Name,
			"email":         id.Email,
			"textSignature": id.TextSignature,
		}
	}
	return map[string]any{
		"accountId": "acc1",
		"state":     "i1",
		"list":      list,
		"notFound":  []any{},
	}
}

func handleIdentitySet(w *World, args map[string]any) map[string]any {
	data := getIdentityData(w)

	resp := map[string]any{
		"accountId": "acc1",
		"oldState":  "i1",
		"newState":  "i2",
	}

	if update, ok := args["update"].(map[string]any); ok {
		updated := make(map[string]any)
		for id, patchRaw := range update {
			patch, ok := patchRaw.(map[string]any)
			if !ok {
				continue
			}
			// Apply the patch to the in-memory identity
			for i := range data.Identities {
				if data.Identities[i].ID == id {
					if v, exists := patch["name"]; exists {
						data.Identities[i].Name, _ = v.(string)
					}
					if v, exists := patch["textSignature"]; exists {
						data.Identities[i].TextSignature, _ = v.(string)
					}
					break
				}
			}
			updated[id] = nil
		}
		resp["updated"] = updated
	}

	return resp
}
