package jmap

import (
	"context"
	"encoding/json"
	"fmt"
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

// NewClient creates a new JMAP client for the given endpoint and access token.
func NewClient(endpoint, accessToken string) *Client {
	return &Client{
		endpoint:    endpoint,
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
}

// Authenticate fetches a new session from the JMAP server.
// This always makes a network request, unlike Session which caches.
func (c *Client) Authenticate(ctx context.Context) (*Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
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
