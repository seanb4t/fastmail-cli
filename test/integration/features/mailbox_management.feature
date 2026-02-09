Feature: Mailbox management
  As a Fastmail user
  I want to manage my mailboxes
  So that I can organize my email

  Background:
    Given a connected Fastmail client

  Scenario: List default mailboxes
    When I list all mailboxes
    Then I should receive 5 mailboxes
    And the mailbox list should contain "Inbox"
    And the mailbox list should contain "Sent"
    And the mailbox list should contain "Trash"
    And the mailbox list should contain "Archive"
    And the mailbox list should contain "Drafts"

  Scenario: Create a new mailbox
    When I create a mailbox named "Projects"
    Then the mailbox creation should succeed
    And the created mailbox ID should not be empty

  Scenario: Create a child mailbox
    When I create a mailbox named "Work" under parent "mb-inbox"
    Then the mailbox creation should succeed
    And the created mailbox ID should not be empty

  Scenario: Rename a mailbox
    When I rename mailbox "mb-archive" to "Old Mail"
    Then the rename operation should succeed

  Scenario: Delete a mailbox
    When I delete mailbox "mb-trash"
    Then the delete operation should succeed
