package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
	"github.com/spf13/cobra"
)

// newContactsCommand creates the contacts command with subcommands.
func newContactsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Contact operations",
		Long:  "Commands for managing Fastmail contacts via CardDAV.",
	}

	cmd.AddCommand(newContactsListCommand())
	cmd.AddCommand(newContactsShowCommand())
	cmd.AddCommand(newContactsCreateCommand())
	cmd.AddCommand(newContactsUpdateCommand())
	cmd.AddCommand(newContactsDeleteCommand())

	return cmd
}

// newContactsListCommand creates the contacts list command.
func newContactsListCommand() *cobra.Command {
	var searchQuery string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Long:  "List all contacts or search by name/email.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createContactsClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			var contacts []fastmail.Contact
			if searchQuery != "" {
				contacts, err = client.Search(ctx, searchQuery)
			} else {
				contacts, err = client.List(ctx)
			}
			if err != nil {
				return fmt.Errorf("listing contacts: %w", err)
			}

			return outputContacts(cmd, contacts)
		},
	}

	cmd.Flags().StringVar(&searchQuery, "search", "", "search contacts by name or email")

	return cmd
}

// newContactsShowCommand creates the contacts show command.
func newContactsShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show ID",
		Short: "Show contact details",
		Long:  "Display detailed information about a single contact.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := createContactsClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			contact, err := client.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("getting contact: %w", err)
			}

			return outputContact(cmd, contact)
		},
	}

	return cmd
}

// newContactsCreateCommand creates the contacts create command.
func newContactsCreateCommand() *cobra.Command {
	var name, email, phone string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact",
		Long:  "Create a new contact in your address book.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			client, err := createContactsClient()
			if err != nil {
				return err
			}

			contact := &fastmail.Contact{
				Name:  name,
				Email: email,
				Phone: phone,
			}

			ctx := context.Background()
			if err := client.Create(ctx, contact); err != nil {
				return fmt.Errorf("creating contact: %w", err)
			}

			return outputContactCreated(cmd, contact)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "contact name (required)")
	cmd.Flags().StringVar(&email, "email", "", "contact email address")
	cmd.Flags().StringVar(&phone, "phone", "", "contact phone number")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newContactsUpdateCommand creates the contacts update command.
func newContactsUpdateCommand() *cobra.Command {
	var name, email, phone string

	cmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update a contact",
		Long:  "Update an existing contact's information.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := createContactsClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// First get the existing contact
			existing, err := client.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("getting contact: %w", err)
			}

			// Update only provided fields
			if name != "" {
				existing.Name = name
			}
			if email != "" {
				existing.Email = email
			}
			if phone != "" {
				existing.Phone = phone
			}

			if err := client.Update(ctx, existing); err != nil {
				return fmt.Errorf("updating contact: %w", err)
			}

			return outputContactStatus(cmd, id, "Updated")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new contact name")
	cmd.Flags().StringVar(&email, "email", "", "new contact email address")
	cmd.Flags().StringVar(&phone, "phone", "", "new contact phone number")

	return cmd
}

// newContactsDeleteCommand creates the contacts delete command.
func newContactsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a contact",
		Long:  "Permanently delete a contact from your address book.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !force {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete contact %s? Use --force to confirm.\n", id)
				return nil
			}

			client, err := createContactsClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Delete(ctx, id); err != nil {
				return fmt.Errorf("deleting contact: %w", err)
			}

			return outputContactStatus(cmd, id, "Deleted")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}

// createContactsClient creates a contacts client from config.
func createContactsClient() (*fastmail.ContactsClient, error) {
	configPath := GetConfigPath()
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Get token from auth store (used as CardDAV password)
	store := auth.NewStore(configPath)
	store.DisableKeychain()

	token, err := store.GetToken()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w (run 'fastmail auth login' first)", err)
	}

	// CardDAV username can come from config or defaults to the token
	username := cfg.CardDAVUsername
	if username == "" {
		return nil, fmt.Errorf("carddav_username not configured (set FASTMAIL_CARDDAV_USERNAME or add to config)")
	}

	return fastmail.NewContactsClient(cfg.CardDAVEndpoint, username, token), nil
}

// outputContacts writes the contact list to output.
func outputContacts(cmd *cobra.Command, contacts []fastmail.Contact) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(contacts)
	}

	// Simple text output
	for _, c := range contacts {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n",
			c.ID, c.Name, c.Email)
	}

	return nil
}

// outputContact writes a single contact to output.
func outputContact(cmd *cobra.Command, contact *fastmail.Contact) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(contact)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", contact.ID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:    %s\n", contact.Name)
	if contact.Email != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email:   %s\n", contact.Email)
	}
	if contact.Phone != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Phone:   %s\n", contact.Phone)
	}
	if contact.Address != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Address: %s\n", contact.Address)
	}

	return nil
}

// outputContactCreated writes the created contact to output.
func outputContactCreated(cmd *cobra.Command, contact *fastmail.Contact) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(contact)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s (%s)\n", contact.Name, contact.ID)
	return nil
}

// outputContactStatus writes a status update to output.
func outputContactStatus(cmd *cobra.Command, id, status string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": id, "status": status}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", status, id)
	return nil
}
