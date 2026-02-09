//go:build integration

package integration

func init() {
	RegisterMethodHandler("Quota/get", handleQuotaGet)
}

func handleQuotaGet(w *World, _ map[string]any) map[string]any {
	data, _ := w.DomainData["quota"].(map[string]any)

	used, _ := data["used"].(uint64)
	limit, _ := data["limit"].(uint64)

	return map[string]any{
		"accountId": "acc1",
		"state":     "q1",
		"list": []map[string]any{
			{
				"id":           "quota-octets",
				"resourceType": "octets",
				"used":         used,
				"hardLimit":    limit,
				"scope":        "account",
				"name":         "Storage",
			},
		},
		"notFound": []any{},
	}
}
