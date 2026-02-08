// Package fastmail provides a high-level client for Fastmail operations.
package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// Client provides high-level Fastmail operations.
type Client struct {
	jmap      *jmap.Client
	accountID string
}

// NewClient creates a new Fastmail client.
// The endpoint should be the JMAP session URL (e.g., https://api.fastmail.com/jmap/session).
func NewClient(endpoint, accessToken string) *Client {
	return &Client{
		jmap: jmap.NewClient(endpoint, accessToken),
	}
}

// Connect establishes a session and retrieves the account ID.
// Must be called before using other operations.
func (c *Client) Connect(ctx context.Context) error {
	session, err := c.jmap.Authenticate(ctx)
	if err != nil {
		return oops.Wrapf(err, "authenticating")
	}

	c.accountID = session.MailAccountID()
	if c.accountID == "" {
		return oops.Errorf("no mail account found")
	}

	return nil
}

// Mail returns the mail service for email operations.
func (c *Client) Mail() *MailService {
	return &MailService{
		client: c,
	}
}

// Mailbox returns the mailbox service for mailbox management operations.
func (c *Client) Mailbox() *MailboxService {
	return &MailboxService{
		client: c,
	}
}

// MaskedEmail returns the masked email service for masked email operations.
func (c *Client) MaskedEmail() *MaskedEmailService {
	return &MaskedEmailService{
		client: c,
	}
}

// Vacation returns the vacation response service for out-of-office operations.
func (c *Client) Vacation() *VacationService {
	return &VacationService{
		client: c,
	}
}

// Quota returns the quota service for storage quota operations.
func (c *Client) Quota() *QuotaService {
	return &QuotaService{
		client: c,
	}
}

// Identity returns the identity service for sender identity operations.
func (c *Client) Identity() *IdentityService {
	return &IdentityService{
		client: c,
	}
}

// accountID returns the current account ID, fetching session if needed.
func (c *Client) getAccountID(ctx context.Context) (string, error) {
	if c.accountID != "" {
		return c.accountID, nil
	}

	session, err := c.jmap.Session(ctx)
	if err != nil {
		return "", oops.Wrapf(err, "getting session")
	}

	c.accountID = session.MailAccountID()
	if c.accountID == "" {
		return "", oops.Errorf("no mail account found")
	}

	return c.accountID, nil
}
