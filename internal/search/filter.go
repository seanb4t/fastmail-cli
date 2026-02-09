// Package search provides structured email search query parsing and filter
// construction for JMAP (RFC 8621) email queries.
package search

import "time"

// Operator defines the boolean logic for combining filters.
type Operator string

const (
	// OpAND combines filters with logical AND (all must match).
	OpAND Operator = "AND"
	// OpOR combines filters with logical OR (any must match).
	OpOR Operator = "OR"
	// OpNOT negates the contained filters.
	OpNOT Operator = "NOT"
)

// FilterCondition represents a single JMAP email filter condition.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-4.4.1
type FilterCondition struct {
	InMailbox      string
	InMailboxOther []string
	From           string
	To             string
	Cc             string
	Bcc            string
	Subject        string
	Body           string
	Text           string
	Before         *time.Time
	After          *time.Time
	MinSize        uint64
	MaxSize        uint64
	HasAttachment  *bool
	HasKeyword     string
	NotKeyword     string
	Header         [2]string // [name, value]
}

// CompoundFilter combines multiple filters with a boolean operator.
type CompoundFilter struct {
	Operator Operator
	Filters  []Filter
}

// Filter is a discriminated union: either a single condition or a compound filter.
type Filter struct {
	Condition *FilterCondition
	Compound  *CompoundFilter
}

// NewConditionFilter wraps a FilterCondition in a Filter.
func NewConditionFilter(c *FilterCondition) Filter {
	return Filter{Condition: c}
}

// NewCompoundFilter wraps a CompoundFilter in a Filter.
func NewCompoundFilter(op Operator, filters ...Filter) Filter {
	return Filter{
		Compound: &CompoundFilter{
			Operator: op,
			Filters:  filters,
		},
	}
}

// ToJMAP converts a Filter to the JMAP filter object format.
func (f Filter) ToJMAP() map[string]any {
	if f.Condition != nil {
		return conditionToJMAP(f.Condition)
	}
	if f.Compound != nil {
		return compoundToJMAP(f.Compound)
	}
	return map[string]any{}
}

func conditionToJMAP(c *FilterCondition) map[string]any {
	m := make(map[string]any)

	if c.InMailbox != "" {
		m["inMailbox"] = c.InMailbox
	}
	if len(c.InMailboxOther) > 0 {
		m["inMailboxOtherThan"] = c.InMailboxOther
	}
	if c.From != "" {
		m["from"] = c.From
	}
	if c.To != "" {
		m["to"] = c.To
	}
	if c.Cc != "" {
		m["cc"] = c.Cc
	}
	if c.Bcc != "" {
		m["bcc"] = c.Bcc
	}
	if c.Subject != "" {
		m["subject"] = c.Subject
	}
	if c.Body != "" {
		m["body"] = c.Body
	}
	if c.Text != "" {
		m["text"] = c.Text
	}
	if c.Before != nil {
		m["before"] = c.Before.UTC().Format(time.RFC3339)
	}
	if c.After != nil {
		m["after"] = c.After.UTC().Format(time.RFC3339)
	}
	if c.MinSize > 0 {
		m["minSize"] = c.MinSize
	}
	if c.MaxSize > 0 {
		m["maxSize"] = c.MaxSize
	}
	if c.HasAttachment != nil {
		m["hasAttachment"] = *c.HasAttachment
	}
	if c.HasKeyword != "" {
		m["hasKeyword"] = c.HasKeyword
	}
	if c.NotKeyword != "" {
		m["notKeyword"] = c.NotKeyword
	}
	if c.Header[0] != "" {
		m["header"] = []string{c.Header[0], c.Header[1]}
	}

	return m
}

func compoundToJMAP(cf *CompoundFilter) map[string]any {
	m := make(map[string]any)
	m["operator"] = string(cf.Operator)

	conditions := make([]map[string]any, len(cf.Filters))
	for i, f := range cf.Filters {
		conditions[i] = f.ToJMAP()
	}
	m["conditions"] = conditions

	return m
}
