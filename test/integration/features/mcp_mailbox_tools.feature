Feature: MCP mailbox tools
  As an AI agent using MCP
  I want to use mailbox tools to interact with Fastmail
  So that I can help users manage their mailbox folders

  Background:
    Given a connected MCP tools config

  Scenario: MCP mailbox_list tool returns default mailboxes
    When I call the "mailbox_list" tool
    Then the tool call should succeed
    And the tool result should contain 5 items
    And the tool result item 1 field "id" should equal "mb-inbox"
    And the tool result item 1 field "name" should equal "Inbox"
    And the tool result item 1 field "role" should equal "inbox"
    And the tool result item 2 field "id" should equal "mb-sent"
    And the tool result item 2 field "name" should equal "Sent"
    And the tool result item 3 field "id" should equal "mb-trash"
    And the tool result item 3 field "name" should equal "Trash"
    And the tool result item 4 field "id" should equal "mb-archive"
    And the tool result item 4 field "name" should equal "Archive"
    And the tool result item 5 field "id" should equal "mb-drafts"
    And the tool result item 5 field "name" should equal "Drafts"

  Scenario: MCP mailbox_create tool creates a new mailbox
    When I call the "mailbox_create" tool with:
      | key  | value    |
      | name | Projects |
    Then the tool call should succeed
    And the tool result field "name" should equal "Projects"
    And the tool result field "status" should equal "created"

  Scenario: MCP mailbox_create tool with parent_id
    When I call the "mailbox_create" tool with:
      | key       | value     |
      | name      | Work      |
      | parent_id | mb-inbox  |
    Then the tool call should succeed
    And the tool result field "name" should equal "Work"
    And the tool result field "status" should equal "created"

  Scenario: MCP mailbox_create tool fails without name
    When I call the "mailbox_create" tool
    Then the tool call should fail
    And the tool error should contain "name is required"

  Scenario: MCP mailbox_rename tool renames a mailbox
    When I call the "mailbox_rename" tool with:
      | key  | value      |
      | id   | mb-archive |
      | name | Old Mail   |
    Then the tool call should succeed
    And the tool result field "id" should equal "mb-archive"
    And the tool result field "name" should equal "Old Mail"
    And the tool result field "status" should equal "renamed"

  Scenario: MCP mailbox_rename tool fails without id
    When I call the "mailbox_rename" tool with:
      | key  | value    |
      | name | New Name |
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP mailbox_rename tool fails without name
    When I call the "mailbox_rename" tool with:
      | key | value      |
      | id  | mb-archive |
    Then the tool call should fail
    And the tool error should contain "name is required"

  Scenario: MCP mailbox_delete tool deletes a mailbox
    When I call the "mailbox_delete" tool with:
      | key | value    |
      | id  | mb-trash |
    Then the tool call should succeed
    And the tool result field "id" should equal "mb-trash"
    And the tool result field "status" should equal "deleted"

  Scenario: MCP mailbox_delete tool fails without id
    When I call the "mailbox_delete" tool
    Then the tool call should fail
    And the tool error should contain "id is required"
