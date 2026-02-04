package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// ProtocolVersion is the MCP protocol version supported by this server.
const ProtocolVersion = "2024-11-05"

// ServerInfo contains information about the MCP server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities declares what the server supports.
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

// ToolsCapability declares tool-related capabilities.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability declares resource-related capabilities.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// InitializeResult is returned from the initialize method.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ToolHandler is a function that executes a tool with the given arguments.
type ToolHandler func(ctx context.Context, args map[string]any) (any, error)

// ResourceReader is a function that reads a resource by URI.
type ResourceReader func(ctx context.Context, uri string) (*ResourceContent, error)

// Server implements an MCP server with stdio transport.
type Server struct {
	info      ServerInfo
	tools     map[string]*registeredTool
	resources map[string]*Resource
	readers   map[string]ResourceReader
	mu        sync.RWMutex

	// I/O for stdio transport
	input  io.Reader
	output io.Writer
}

// registeredTool holds a tool definition and its handler.
type registeredTool struct {
	tool    *Tool
	handler ToolHandler
}

// NewServer creates a new MCP server.
func NewServer(name, version string) *Server {
	return &Server{
		info: ServerInfo{
			Name:    name,
			Version: version,
		},
		tools:     make(map[string]*registeredTool),
		resources: make(map[string]*Resource),
		readers:   make(map[string]ResourceReader),
	}
}

// RegisterTool adds a tool to the server.
func (s *Server) RegisterTool(tool *Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = &registeredTool{
		tool:    tool,
		handler: handler,
	}
}

// RegisterResource adds a resource to the server.
func (s *Server) RegisterResource(resource *Resource, reader ResourceReader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[resource.URI] = resource
	s.readers[resource.URI] = reader
}

// Run starts the server's main loop, reading from stdin and writing to stdout.
func (s *Server) Run(ctx context.Context, input io.Reader, output io.Writer) error {
	s.input = input
	s.output = output

	scanner := bufio.NewScanner(input)
	// Increase buffer size for large messages
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading input: %w", err)
			}
			// EOF - clean shutdown
			return nil
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp := s.handleMessage(ctx, line)
		if resp != nil {
			if err := s.writeResponse(resp); err != nil {
				return fmt.Errorf("writing response: %w", err)
			}
		}
	}
}

// handleMessage processes a single JSON-RPC message.
func (s *Server) handleMessage(ctx context.Context, data []byte) *Response {
	req, err := ParseRequest(data)
	if err != nil {
		return NewErrorResponse(RequestID{}, ErrorCodeParseError, "Parse error: "+err.Error())
	}

	// Notifications don't get responses
	if req.IsNotification() {
		s.handleNotification(req)
		return nil
	}

	return s.handleRequest(ctx, req)
}

// handleNotification processes a notification (no response expected).
func (s *Server) handleNotification(req *Request) {
	// Handle known notifications.
	if req.Method == MethodNotificationsInitialized {
		// Client acknowledges initialization - nothing to do.
		return
	}
	// Unknown notifications are silently ignored per MCP spec.
}

// handleRequest processes a request and returns a response.
func (s *Server) handleRequest(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case MethodToolsList:
		return s.handleToolsList(req)
	case MethodToolsCall:
		return s.handleToolsCall(ctx, req)
	case MethodResourcesList:
		return s.handleResourcesList(req)
	case MethodResourcesRead:
		return s.handleResourcesRead(ctx, req)
	default:
		return NewErrorResponse(req.ID, ErrorCodeMethodNotFound, "Method not found: "+req.Method)
	}
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(req *Request) *Response {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools:     &ToolsCapability{},
			Resources: &ResourcesCapability{},
		},
		ServerInfo: s.info,
	}

	resp, err := NewResponse(req.ID, result)
	if err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInternalError, "Failed to create response: "+err.Error())
	}
	return resp
}

// ToolsListResult is the result of tools/list.
type ToolsListResult struct {
	Tools []*Tool `json:"tools"`
}

// handleToolsList returns the list of registered tools.
func (s *Server) handleToolsList(req *Request) *Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]*Tool, 0, len(s.tools))
	for _, rt := range s.tools {
		tools = append(tools, rt.tool)
	}

	result := ToolsListResult{Tools: tools}
	resp, err := NewResponse(req.ID, result)
	if err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInternalError, "Failed to create response: "+err.Error())
	}
	return resp
}

// ToolsCallParams are the parameters for tools/call.
type ToolsCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolsCallResult is the result of tools/call.
type ToolsCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content block in a tool result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// handleToolsCall executes a tool.
func (s *Server) handleToolsCall(ctx context.Context, req *Request) *Response {
	var params ToolsCallParams
	if err := req.DecodeParams(&params); err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid params: "+err.Error())
	}

	s.mu.RLock()
	rt, ok := s.tools[params.Name]
	s.mu.RUnlock()

	if !ok {
		return NewErrorResponse(req.ID, ErrorCodeInvalidParams, "Unknown tool: "+params.Name)
	}

	result, err := rt.handler(ctx, params.Arguments)
	if err != nil {
		// Return error as tool result, not JSON-RPC error
		callResult := ToolsCallResult{
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			IsError: true,
		}
		resp, _ := NewResponse(req.ID, callResult)
		return resp
	}

	// Convert result to text content
	var text string
	switch v := result.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			text = fmt.Sprintf("%v", v)
		} else {
			text = string(data)
		}
	}

	callResult := ToolsCallResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
	resp, err := NewResponse(req.ID, callResult)
	if err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInternalError, "Failed to create response: "+err.Error())
	}
	return resp
}

// ResourcesListResult is the result of resources/list.
type ResourcesListResult struct {
	Resources []*Resource `json:"resources"`
}

// handleResourcesList returns the list of registered resources.
func (s *Server) handleResourcesList(req *Request) *Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]*Resource, 0, len(s.resources))
	for _, r := range s.resources {
		resources = append(resources, r)
	}

	result := ResourcesListResult{Resources: resources}
	resp, err := NewResponse(req.ID, result)
	if err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInternalError, "Failed to create response: "+err.Error())
	}
	return resp
}

// ResourcesReadParams are the parameters for resources/read.
type ResourcesReadParams struct {
	URI string `json:"uri"`
}

// ResourcesReadResult is the result of resources/read.
type ResourcesReadResult struct {
	Contents []*ResourceContent `json:"contents"`
}

// handleResourcesRead reads a resource.
func (s *Server) handleResourcesRead(ctx context.Context, req *Request) *Response {
	var params ResourcesReadParams
	if err := req.DecodeParams(&params); err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid params: "+err.Error())
	}

	s.mu.RLock()
	reader, ok := s.readers[params.URI]
	s.mu.RUnlock()

	if !ok {
		return NewErrorResponse(req.ID, ErrorCodeInvalidParams, "Unknown resource: "+params.URI)
	}

	content, err := reader(ctx, params.URI)
	if err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInternalError, "Failed to read resource: "+err.Error())
	}

	result := ResourcesReadResult{
		Contents: []*ResourceContent{content},
	}
	resp, err := NewResponse(req.ID, result)
	if err != nil {
		return NewErrorResponse(req.ID, ErrorCodeInternalError, "Failed to create response: "+err.Error())
	}
	return resp
}

// writeResponse writes a JSON-RPC response to output.
func (s *Server) writeResponse(resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshaling response: %w", err)
	}

	// Write response followed by newline (MCP stdio protocol)
	if _, err := fmt.Fprintf(s.output, "%s\n", data); err != nil {
		return fmt.Errorf("writing response: %w", err)
	}

	return nil
}
