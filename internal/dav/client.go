package dav

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

// Fastmail DAV endpoints.
const (
	FastmailCardDAVEndpoint = "https://carddav.fastmail.com/"
	FastmailCalDAVEndpoint  = "https://caldav.fastmail.com/"
)

// CardDAVClient wraps the carddav.Client with Fastmail-specific configuration.
type CardDAVClient struct {
	client    *carddav.Client
	principal string
}

// CalDAVClient wraps the caldav.Client with Fastmail-specific configuration.
type CalDAVClient struct {
	client    *caldav.Client
	principal string
}

// NewCardDAVClient creates a new CardDAV client for Fastmail.
// The user parameter should be the Fastmail email address.
// The token parameter should be an API token with contacts access.
func NewCardDAVClient(endpoint, user, token string) (*CardDAVClient, error) {
	if endpoint == "" {
		endpoint = FastmailCardDAVEndpoint
	}

	httpClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, user, token)

	client, err := carddav.NewClient(httpClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("creating carddav client: %w", err)
	}

	return &CardDAVClient{
		client:    client,
		principal: buildPrincipal(user),
	}, nil
}

// NewCalDAVClient creates a new CalDAV client for Fastmail.
// The user parameter should be the Fastmail email address.
// The token parameter should be an API token with calendar access.
func NewCalDAVClient(endpoint, user, token string) (*CalDAVClient, error) {
	if endpoint == "" {
		endpoint = FastmailCalDAVEndpoint
	}

	httpClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, user, token)

	client, err := caldav.NewClient(httpClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("creating caldav client: %w", err)
	}

	return &CalDAVClient{
		client:    client,
		principal: buildPrincipal(user),
	}, nil
}

// buildPrincipal constructs the principal path for a Fastmail user.
func buildPrincipal(user string) string {
	return fmt.Sprintf("/dav/principals/user/%s/", user)
}

// FindAddressBookHomeSet discovers the address book home set for the user.
func (c *CardDAVClient) FindAddressBookHomeSet(ctx context.Context) (string, error) {
	homeSet, err := c.client.FindAddressBookHomeSet(ctx, c.principal)
	if err != nil {
		return "", fmt.Errorf("finding address book home set: %w", err)
	}
	return homeSet, nil
}

// FindAddressBooks returns all address books for the user.
func (c *CardDAVClient) FindAddressBooks(ctx context.Context) ([]carddav.AddressBook, error) {
	homeSet, err := c.FindAddressBookHomeSet(ctx)
	if err != nil {
		return nil, err
	}

	books, err := c.client.FindAddressBooks(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding address books: %w", err)
	}
	return books, nil
}

// Client returns the underlying carddav.Client for advanced operations.
func (c *CardDAVClient) Client() *carddav.Client {
	return c.client
}

// FindCalendarHomeSet discovers the calendar home set for the user.
func (c *CalDAVClient) FindCalendarHomeSet(ctx context.Context) (string, error) {
	homeSet, err := c.client.FindCalendarHomeSet(ctx, c.principal)
	if err != nil {
		return "", fmt.Errorf("finding calendar home set: %w", err)
	}
	return homeSet, nil
}

// FindCalendars returns all calendars for the user.
func (c *CalDAVClient) FindCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	homeSet, err := c.FindCalendarHomeSet(ctx)
	if err != nil {
		return nil, err
	}

	calendars, err := c.client.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding calendars: %w", err)
	}
	return calendars, nil
}

// Client returns the underlying caldav.Client for advanced operations.
func (c *CalDAVClient) Client() *caldav.Client {
	return c.client
}
