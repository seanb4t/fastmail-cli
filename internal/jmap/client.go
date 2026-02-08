package jmap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/samber/oops"
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

// HTTPError represents a non-200 HTTP response.
type HTTPError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unexpected status: %d", e.StatusCode)
	}
	return fmt.Sprintf("unexpected status %d: %s", e.StatusCode, e.Body)
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
		return nil, oops.Wrapf(err, "creating request")
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, oops.Wrapf(err, "decoding session")
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
		return nil, oops.Wrapf(err, "getting session")
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, oops.Wrapf(err, "marshaling request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, session.APIURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, oops.Wrapf(err, "creating request")
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	return &response, nil
}

// DownloadBlob downloads a blob by ID using the session's download URL template.
func (c *Client) DownloadBlob(ctx context.Context, accountID, blobID string) (io.ReadCloser, error) {
	session, err := c.Session(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "getting session")
	}

	url := session.DownloadURL
	url = strings.ReplaceAll(url, "{accountId}", accountID)
	url = strings.ReplaceAll(url, "{blobId}", blobID)
	url = strings.ReplaceAll(url, "{name}", "raw")
	url = strings.ReplaceAll(url, "{type}", "application/octet-stream")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, oops.Wrapf(err, "creating request")
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, oops.Wrapf(err, "downloading blob")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return resp.Body, nil
}
