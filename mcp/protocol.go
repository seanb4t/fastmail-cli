// Package mcp provides MCP (Model Context Protocol) server implementation
// for exposing Fastmail functionality to AI agents.
package mcp

import (
	"encoding/json"
)

// JSONRPCVersion is the JSON-RPC 2.0 protocol version.
const JSONRPCVersion = "2.0"

// MCP method names.
const (
	// MethodInitialize is the discovery method for initialization.
	MethodInitialize = "initialize"

	// MethodToolsList lists available tools.
	MethodToolsList = "tools/list"
	// MethodToolsCall executes a tool.
	MethodToolsCall = "tools/call"

	// MethodResourcesList lists available resources.
	MethodResourcesList = "resources/list"
	// MethodResourcesRead reads a resource.
	MethodResourcesRead = "resources/read"
	// MethodResourcesTemplates lists resource templates.
	MethodResourcesTemplates = "resources/templates"

	// MethodPromptsList lists available prompts.
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet gets a prompt.
	MethodPromptsGet = "prompts/get"

	// MethodNotificationsInitialized is sent when initialization is complete.
	MethodNotificationsInitialized = "notifications/initialized"
)

// Request represents a JSON-RPC 2.0 request.
// See: https://www.jsonrpc.org/specification#request_object
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RequestID represents a JSON-RPC request ID which can be a string or integer.
type RequestID struct {
	value any
}

// NewStringID creates a RequestID from a string.
func NewStringID(s string) RequestID {
	return RequestID{value: s}
}

// NewIntID creates a RequestID from an integer.
func NewIntID(n int) RequestID {
	return RequestID{value: n}
}

// IsNull returns true if the request ID is null (notification).
func (id RequestID) IsNull() bool {
	return id.value == nil
}

// String returns the ID as a string if it's a string, empty otherwise.
func (id RequestID) String() string {
	if s, ok := id.value.(string); ok {
		return s
	}
	return ""
}

// Int returns the ID as an int if it's an int, 0 otherwise.
func (id RequestID) Int() int {
	if n, ok := id.value.(int); ok {
		return n
	}
	return 0
}

// MarshalJSON implements json.Marshaler for RequestID.
func (id RequestID) MarshalJSON() ([]byte, error) {
	if id.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(id.value)
}

// UnmarshalJSON implements json.Unmarshaler for RequestID.
func (id *RequestID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		id.value = nil
		return nil
	}

	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		id.value = s
		return nil
	}

	// Try integer
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		id.value = n
		return nil
	}

	// Keep as raw value
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	id.value = v
	return nil
}

// ParseRequest parses a JSON-RPC request from JSON bytes.
func ParseRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// IsNotification returns true if this is a notification (no ID).
func (r *Request) IsNotification() bool {
	return r.ID.IsNull()
}

// DecodeParams unmarshals the request params into the given value.
func (r *Request) DecodeParams(v any) error {
	if r.Params == nil {
		return nil
	}
	return json.Unmarshal(r.Params, v)
}

// Response represents a JSON-RPC 2.0 response.
// See: https://www.jsonrpc.org/specification#response_object
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents a JSON-RPC 2.0 error object.
// See: https://www.jsonrpc.org/specification#error_object
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)

// Error implements the error interface.
func (e *ResponseError) Error() string {
	return e.Message
}

// NewResponse creates a successful response with the given result.
func NewResponse(id RequestID, result any) (*Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  resultJSON,
	}, nil
}

// NewErrorResponse creates an error response.
func NewErrorResponse(id RequestID, code int, message string) *Response {
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &ResponseError{
			Code:    code,
			Message: message,
		},
	}
}

// IsError returns true if this response contains an error.
func (r *Response) IsError() bool {
	return r.Error != nil
}

// DecodeResult unmarshals the response result into the given value.
func (r *Response) DecodeResult(v any) error {
	if r.Result == nil {
		return nil
	}
	return json.Unmarshal(r.Result, v)
}

// Tool represents an MCP tool definition.
// See: https://spec.modelcontextprotocol.io/specification/server/tools/
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema represents the JSON Schema for a tool's input.
type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema represents a single property in a JSON Schema.
type PropertySchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// NewTool creates a new tool definition.
func NewTool(name, description string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		InputSchema: InputSchema{
			Type:       "object",
			Properties: make(map[string]PropertySchema),
		},
	}
}

// WithProperty adds a property to the tool's input schema.
func (t *Tool) WithProperty(name, propType, description string) *Tool {
	t.InputSchema.Properties[name] = PropertySchema{
		Type:        propType,
		Description: description,
	}
	return t
}

// WithRequired marks properties as required.
func (t *Tool) WithRequired(names ...string) *Tool {
	t.InputSchema.Required = append(t.InputSchema.Required, names...)
	return t
}

// Resource represents an MCP resource definition.
// See: https://spec.modelcontextprotocol.io/specification/server/resources/
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// NewResource creates a new resource definition.
func NewResource(uri, name string) *Resource {
	return &Resource{
		URI:  uri,
		Name: name,
	}
}

// WithDescription sets the resource description.
func (r *Resource) WithDescription(description string) *Resource {
	r.Description = description
	return r
}

// WithMimeType sets the resource MIME type.
func (r *Resource) WithMimeType(mimeType string) *Resource {
	r.MimeType = mimeType
	return r
}

// ResourceContent represents the content returned when reading a resource.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// Prompt represents an MCP prompt definition.
// See: https://spec.modelcontextprotocol.io/specification/server/prompts/
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents an argument in a prompt definition.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}
