package jmap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Client is a JMAP protocol client.
type Client struct {
	endpoint    string
	accessToken string
	httpClient  *http.Client

	// session caching
	mu      sync.RWMutex
	session *Session
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client for the JMAP client.
// This is useful for testing with mock transports.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new JMAP client for the given endpoint and access token.
func NewClient(endpoint, accessToken string, opts ...ClientOption) *Client {
	c := &Client{
		endpoint:    endpoint,
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Authenticate fetches a new session from the JMAP server.
// This always makes a network request, unlike Session which caches.
func (c *Client) Authenticate(ctx context.Context) (*Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decoding session: %w", err)
	}

	// Cache the session
	c.mu.Lock()
	c.session = &session
	c.mu.Unlock()

	return &session, nil
}

// Session returns the cached session, fetching it if necessary.
func (c *Client) Session(ctx context.Context) (*Session, error) {
	c.mu.RLock()
	if c.session != nil {
		session := c.session
		c.mu.RUnlock()
		return session, nil
	}
	c.mu.RUnlock()

	return c.Authenticate(ctx)
}

// Call executes a JMAP API request and returns the response.
// It uses the session's API URL for the request endpoint.
func (c *Client) Call(ctx context.Context, request *Request) (*Response, error) {
	session, err := c.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, session.APIURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}
