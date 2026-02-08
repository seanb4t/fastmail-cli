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
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
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
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

// UploadResponse represents the server response after a successful blob upload.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-6.1
type UploadResponse struct {
	AccountID string `json:"accountId"`
	BlobID    string `json:"blobId"`
	Type      string `json:"type"`
	Size      uint64 `json:"size"`
}

// DownloadBlob downloads a blob from the server using the session's download URL template.
// The caller is responsible for closing the returned ReadCloser.
func (c *Client) DownloadBlob(ctx context.Context, accountID, blobID, name string) (io.ReadCloser, error) {
	session, err := c.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	url := buildDownloadURL(session.DownloadURL, accountID, blobID, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return resp.Body, nil
}

// UploadBlob uploads binary data to the server using the session's upload URL template.
func (c *Client) UploadBlob(ctx context.Context, accountID string, data io.Reader, contentType string) (*UploadResponse, error) {
	session, err := c.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	url := buildUploadURL(session.UploadURL, accountID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, data)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("decoding upload response: %w", err)
	}

	return &uploadResp, nil
}

// buildDownloadURL replaces placeholders in the JMAP download URL template.
func buildDownloadURL(template, accountID, blobID, name string) string {
	r := strings.NewReplacer(
		"{accountId}", accountID,
		"{blobId}", blobID,
		"{name}", name,
		"{type}", "application/octet-stream",
	)
	return r.Replace(template)
}

// buildUploadURL replaces placeholders in the JMAP upload URL template.
func buildUploadURL(template, accountID string) string {
	return strings.Replace(template, "{accountId}", accountID, 1)
}
