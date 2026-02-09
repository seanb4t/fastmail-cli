package search

import (
	"strings"
	"time"
)

// Parse converts a search query string into a Filter.
//
// Supported syntax:
//
//	from:alice          → FilterCondition{From: "alice"}
//	to:bob              → FilterCondition{To: "bob"}
//	cc:carol            → FilterCondition{Cc: "carol"}
//	bcc:dave            → FilterCondition{Bcc: "dave"}
//	subject:"meeting"   → FilterCondition{Subject: "meeting"}
//	body:agenda         → FilterCondition{Body: "agenda"}
//	before:2024-06-01   → FilterCondition{Before: ...}
//	after:2024-01-01    → FilterCondition{After: ...}
//	has:attachment       → FilterCondition{HasAttachment: true}
//	is:unread            → FilterCondition{NotKeyword: "$seen"}
//	is:flagged           → FilterCondition{HasKeyword: "$flagged"}
//	is:read              → FilterCondition{HasKeyword: "$seen"}
//	in:Sent / folder:Sent → FilterCondition{InMailbox: "Sent"}
//	-from:spam           → NOT(FilterCondition{From: "spam"})
//	A OR B               → OR compound filter
//	bare words           → FilterCondition{Text: "bare words"}
//
// Multiple conditions are ANDed by default. Folder names remain as strings;
// the service layer resolves them to IDs.
func Parse(query string) Filter {
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return NewConditionFilter(&FilterCondition{})
	}
	return buildFilter(tokens)
}

// token represents a parsed token from the query string.
type token struct {
	negated bool
	field   string // empty for bare words and OR operator
	value   string
	isOR    bool // true for the "OR" operator token
}

// tokenize splits a query string into tokens.
func tokenize(query string) []token {
	var tokens []token
	i := 0
	runes := []rune(query)

	for i < len(runes) {
		// Skip whitespace
		if runes[i] == ' ' || runes[i] == '\t' {
			i++
			continue
		}

		tok := scanToken(runes, &i)
		tokens = append(tokens, tok)
	}

	return tokens
}

// scanToken reads one token starting at runes[*pos] and advances *pos.
func scanToken(runes []rune, pos *int) token {
	i := *pos
	negated := false

	// Check for negation prefix
	if runes[i] == '-' && i+1 < len(runes) && runes[i+1] != ' ' {
		negated = true
		i++
	}

	// Find the token boundary
	start := i
	var field, value string

	if runes[i] == '"' {
		// Quoted bare text
		value = scanQuoted(runes, &i)
		*pos = i
		return token{negated: negated, value: value}
	}

	// Scan to whitespace or end
	for i < len(runes) && runes[i] != ' ' && runes[i] != '\t' {
		if runes[i] == ':' && field == "" {
			field = string(runes[start:i])
			i++
			// Value may be quoted
			if i < len(runes) && runes[i] == '"' {
				value = scanQuoted(runes, &i)
			} else {
				vStart := i
				for i < len(runes) && runes[i] != ' ' && runes[i] != '\t' {
					i++
				}
				value = string(runes[vStart:i])
			}
			*pos = i
			return token{negated: negated, field: field, value: value}
		}
		i++
	}

	// Bare word (no colon found)
	word := string(runes[start:i])
	*pos = i

	// Check for OR operator
	if !negated && word == "OR" {
		return token{isOR: true}
	}

	return token{negated: negated, value: word}
}

// scanQuoted reads a quoted string starting at runes[*pos] (which must be '"')
// and advances *pos past the closing quote.
func scanQuoted(runes []rune, pos *int) string {
	i := *pos + 1 // skip opening quote
	var sb strings.Builder
	for i < len(runes) {
		if runes[i] == '"' {
			i++ // skip closing quote
			break
		}
		sb.WriteRune(runes[i])
		i++
	}
	*pos = i
	return sb.String()
}

// buildFilter converts tokens into a Filter, handling OR and AND grouping.
func buildFilter(tokens []token) Filter {
	// Split on OR operators first (lower precedence than AND)
	groups := splitOnOR(tokens)

	if len(groups) == 1 {
		return buildANDGroup(groups[0])
	}

	// Multiple OR groups
	orFilters := make([]Filter, 0, len(groups))
	for _, group := range groups {
		orFilters = append(orFilters, buildANDGroup(group))
	}
	return NewCompoundFilter(OpOR, orFilters...)
}

// splitOnOR divides tokens into groups separated by OR operators.
func splitOnOR(tokens []token) [][]token {
	var groups [][]token
	var current []token

	for _, tok := range tokens {
		if tok.isOR {
			if len(current) > 0 {
				groups = append(groups, current)
				current = nil
			}
			continue
		}
		current = append(current, tok)
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}

	return groups
}

// buildANDGroup converts a group of tokens (no OR operators) into a filter.
func buildANDGroup(tokens []token) Filter {
	var filters []Filter
	var bareWords []string

	for _, tok := range tokens {
		f, drop := tokenToFilter(tok)
		if f != nil {
			filters = append(filters, *f)
		} else if !drop {
			bareWords = append(bareWords, tok.value)
		}
	}

	// Combine bare words into a single text filter
	if len(bareWords) > 0 {
		text := strings.Join(bareWords, " ")
		filters = append(filters, NewConditionFilter(&FilterCondition{Text: text}))
	}

	if len(filters) == 0 {
		return NewConditionFilter(&FilterCondition{})
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return NewCompoundFilter(OpAND, filters...)
}

// tokenToFilter converts a single field:value token to a Filter.
// Returns nil for bare words and unknown fields (accumulated as text).
// Returns a special empty filter pointer for known fields with invalid values (dropped).
func tokenToFilter(tok token) (*Filter, bool) {
	if tok.field == "" {
		return nil, false // bare word
	}

	if !isKnownField(tok.field) {
		return nil, false // unknown field, treat as bare text
	}

	cond := fieldToCondition(tok.field, tok.value)
	if cond == nil {
		return nil, true // known field, invalid value — drop silently
	}

	f := NewConditionFilter(cond)
	if tok.negated {
		f = NewCompoundFilter(OpNOT, f)
	}

	return &f, false
}

// knownFields lists all recognized field prefixes.
var knownFields = map[string]bool{
	"from": true, "to": true, "cc": true, "bcc": true,
	"subject": true, "body": true, "before": true, "after": true,
	"has": true, "is": true, "in": true, "folder": true,
}

// isKnownField reports whether a field name is a recognized search prefix.
func isKnownField(field string) bool {
	return knownFields[strings.ToLower(field)]
}

// fieldToCondition maps a field:value pair to a FilterCondition.
func fieldToCondition(field, value string) *FilterCondition {
	c := &FilterCondition{}

	switch strings.ToLower(field) {
	case "from":
		c.From = value
	case "to":
		c.To = value
	case "cc":
		c.Cc = value
	case "bcc":
		c.Bcc = value
	case "subject":
		c.Subject = value
	case "body":
		c.Body = value
	case "before":
		t, err := parseDate(value)
		if err != nil {
			return nil
		}
		c.Before = &t
	case "after":
		t, err := parseDate(value)
		if err != nil {
			return nil
		}
		c.After = &t
	case "has":
		if strings.EqualFold(value, "attachment") {
			hasAtt := true
			c.HasAttachment = &hasAtt
		} else {
			return nil
		}
	case "is":
		return isCondition(value)
	case "in", "folder":
		c.InMailbox = value
	default:
		return nil
	}

	return c
}

// isCondition handles "is:unread", "is:read", "is:flagged", etc.
func isCondition(value string) *FilterCondition {
	c := &FilterCondition{}
	switch strings.ToLower(value) {
	case "unread":
		c.NotKeyword = "$seen"
	case "read":
		c.HasKeyword = "$seen"
	case "flagged", "starred":
		c.HasKeyword = "$flagged"
	case "unflagged", "unstarred":
		c.NotKeyword = "$flagged"
	case "draft":
		c.HasKeyword = "$draft"
	case "answered":
		c.HasKeyword = "$answered"
	default:
		return nil
	}
	return c
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
