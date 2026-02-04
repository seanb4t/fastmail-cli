package jmap

import (
	"encoding/json"
	"fmt"
)

// Response represents a JMAP API response.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-3.4
type Response struct {
	SessionState    string         `json:"sessionState"`
	MethodResponses []MethodResult `json:"methodResponses"`
}

// MethodResult represents a single method response in a JMAP response.
// Like Invocation, it serializes as a 3-element array: [methodName, arguments, callId].
type MethodResult struct {
	Name   string
	Args   json.RawMessage
	CallID string
}

// MarshalJSON implements json.Marshaler for MethodResult.
func (r MethodResult) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{r.Name, r.Args, r.CallID})
}

// UnmarshalJSON implements json.Unmarshaler for MethodResult.
func (r *MethodResult) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 3 {
		return &json.UnmarshalTypeError{Value: "array", Type: nil}
	}

	if err := json.Unmarshal(arr[0], &r.Name); err != nil {
		return err
	}
	// Keep Args as raw JSON for later decoding
	r.Args = arr[1]
	if err := json.Unmarshal(arr[2], &r.CallID); err != nil {
		return err
	}

	return nil
}

// IsError returns true if this method result is an error response.
func (r *MethodResult) IsError() bool {
	return r.Name == "error"
}

// MethodError represents a JMAP method-level error.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-3.6.2
type MethodError struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// Error implements the error interface.
func (e *MethodError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("jmap: %s: %s", e.Type, e.Description)
	}
	return fmt.Sprintf("jmap: %s", e.Type)
}

// Error returns the error details if this is an error response.
// Returns nil if this is not an error response.
func (r *MethodResult) Error() *MethodError {
	if !r.IsError() {
		return nil
	}

	var err MethodError
	if unmarshalErr := json.Unmarshal(r.Args, &err); unmarshalErr != nil {
		return &MethodError{Type: "unknownError", Description: "failed to parse error response"}
	}
	return &err
}

// Decode unmarshals the method result arguments into the given value.
func (r *MethodResult) Decode(v any) error {
	return json.Unmarshal(r.Args, v)
}

// GetResult returns the method result for the given call ID.
// Returns an error if no result exists for that call ID.
func (r *Response) GetResult(callID string) (*MethodResult, error) {
	for i := range r.MethodResponses {
		if r.MethodResponses[i].CallID == callID {
			return &r.MethodResponses[i], nil
		}
	}
	return nil, fmt.Errorf("no result for call ID %q", callID)
}
