//go:build integration

package integration

func init() {
	RegisterMethodHandler("VacationResponse/get", handleVacationGet)
	RegisterMethodHandler("VacationResponse/set", handleVacationSet)
}

// vacationState holds the mock vacation response state within DomainData.
type vacationState struct {
	IsEnabled bool
	Subject   string
	TextBody  string
}

const vacationDataKey = "vacation"

func getVacationState(w *World) *vacationState {
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	if v, ok := w.DomainData[vacationDataKey]; ok {
		return v.(*vacationState)
	}
	state := &vacationState{}
	w.DomainData[vacationDataKey] = state
	return state
}

func handleVacationGet(w *World, _ map[string]any) map[string]any {
	state := getVacationState(w)
	vr := map[string]any{
		"id":        "singleton",
		"isEnabled": state.IsEnabled,
		"subject":   state.Subject,
		"textBody":  state.TextBody,
	}
	return map[string]any{
		"accountId": "acc1",
		"state":     "vr1",
		"list":      []map[string]any{vr},
		"notFound":  []any{},
	}
}

func handleVacationSet(w *World, args map[string]any) map[string]any {
	state := getVacationState(w)

	resp := map[string]any{
		"accountId": "acc1",
		"oldState":  "vr1",
		"newState":  "vr2",
	}

	if update, ok := args["update"].(map[string]any); ok {
		updated := make(map[string]any)
		for id, patchRaw := range update {
			patch, ok := patchRaw.(map[string]any)
			if !ok {
				continue
			}
			if v, exists := patch["isEnabled"]; exists {
				state.IsEnabled, _ = v.(bool)
			}
			if v, exists := patch["subject"]; exists {
				state.Subject, _ = v.(string)
			}
			if v, exists := patch["textBody"]; exists {
				state.TextBody, _ = v.(string)
			}
			updated[id] = nil
		}
		resp["updated"] = updated
	}

	return resp
}
