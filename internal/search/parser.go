// Package search provides query string parsing for JMAP Email/query filter conditions.
package search

import "strings"

// Parse converts a human-readable search query into a JMAP FilterCondition.
func Parse(query string) map[string]any {
	tokens := strings.Fields(query)
	if len(tokens) == 0 {
		return map[string]any{}
	}

	var conditions []map[string]any
	var freeText []string

	for _, token := range tokens {
		parts := strings.SplitN(token, ":", 2)
		if len(parts) == 2 {
			prefix, value := parts[0], parts[1]
			if cond := mapToken(prefix, value); cond != nil {
				conditions = append(conditions, cond)
				continue
			}
		}
		freeText = append(freeText, token)
	}

	if len(freeText) > 0 {
		conditions = append(conditions, map[string]any{"text": strings.Join(freeText, " ")})
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	return map[string]any{
		"operator":   "AND",
		"conditions": conditions,
	}
}

func mapToken(prefix, value string) map[string]any {
	switch prefix {
	case "from":
		return map[string]any{"from": value}
	case "to":
		return map[string]any{"to": value}
	case "subject":
		return map[string]any{"subject": value}
	case "before":
		return map[string]any{"before": value}
	case "after":
		return map[string]any{"after": value}
	case "in":
		return map[string]any{"inMailbox": value}
	case "has":
		if value == "attachment" {
			return map[string]any{"hasAttachment": true}
		}
	case "is":
		switch value {
		case "unread":
			return map[string]any{"notKeyword": "$seen"}
		case "flagged":
			return map[string]any{"hasKeyword": "$flagged"}
		}
	}
	return nil
}
