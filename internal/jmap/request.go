package jmap

import (
	"encoding/json"
	"strconv"
)

// Request represents a JMAP API request.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-3.3
type Request struct {
	Using       []string     `json:"using"`
	MethodCalls []Invocation `json:"methodCalls"`

	nextCallID int
}

// Invocation represents a single JMAP method call.
// It serializes as a 3-element array: [methodName, arguments, callId].
type Invocation struct {
	Name   string
	Args   map[string]any
	CallID string
}

// MarshalJSON implements json.Marshaler for Invocation.
// JMAP invocations serialize as [name, args, callId] arrays.
func (i Invocation) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{i.Name, i.Args, i.CallID})
}

// UnmarshalJSON implements json.Unmarshaler for Invocation.
func (i *Invocation) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 3 {
		return &json.UnmarshalTypeError{Value: "array", Type: nil}
	}

	if err := json.Unmarshal(arr[0], &i.Name); err != nil {
		return err
	}
	if err := json.Unmarshal(arr[1], &i.Args); err != nil {
		return err
	}
	if err := json.Unmarshal(arr[2], &i.CallID); err != nil {
		return err
	}

	return nil
}

// ResultRef represents a JMAP result reference for method chaining.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-3.7
type ResultRef struct {
	ResultOf string `json:"resultOf"`
	Name     string `json:"name"`
	Path     string `json:"path"`
}

// ResultReference creates a result reference for chaining method calls.
// Use this when a method argument should come from a previous method's result.
//
// Example:
//
//	queryID := req.Invoke("Email/query", ...)
//	req.Invoke("Email/get", map[string]any{
//	    "#ids": ResultReference(queryID, "Email/query", "/ids"),
//	})
func ResultReference(callID, methodName, path string) ResultRef {
	return ResultRef{
		ResultOf: callID,
		Name:     methodName,
		Path:     path,
	}
}

// NewRequest creates a new empty JMAP request.
func NewRequest() *Request {
	return &Request{
		Using:       []string{},
		MethodCalls: []Invocation{},
	}
}

// WithCapabilities adds capability URIs to the request.
// Returns the request for method chaining.
func (r *Request) WithCapabilities(capabilities ...string) *Request {
	r.Using = append(r.Using, capabilities...)
	return r
}

// Invoke adds a method call to the request.
// Returns the call ID which can be used for result references.
func (r *Request) Invoke(method string, args map[string]any) string {
	callID := strconv.Itoa(r.nextCallID)
	r.nextCallID++

	r.MethodCalls = append(r.MethodCalls, Invocation{
		Name:   method,
		Args:   args,
		CallID: callID,
	})

	return callID
}
