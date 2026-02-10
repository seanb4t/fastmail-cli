Feature: MCP contacts tools
  As an AI agent using MCP
  I want to manage contacts via MCP tools
  So that I can help users with their address book

  Background:
    Given a connected MCP tools config

  Scenario: MCP contacts_list tool
    Given the following contacts exist:
      | id   | name      | email              | phone        |
      | c-1  | Alice Doe | alice@example.com  | 555-0001     |
      | c-2  | Bob Smith | bob@example.com    | 555-0002     |
    When I call the "contacts_list" contacts tool
    Then the tool call should succeed
    And the tool result should contain 2 items

  Scenario: MCP contacts_get tool
    Given the following contacts exist:
      | id   | name      | email              | phone        |
      | c-1  | Alice Doe | alice@example.com  | 555-0001     |
    When I call the "contacts_get" contacts tool with:
      | key | value |
      | id  | c-1   |
    Then the tool call should succeed
    And the tool result field "name" should equal "Alice Doe"

  Scenario: MCP contacts_get fails without id
    When I call the "contacts_get" contacts tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP contacts_create tool
    When I call the "contacts_create" contacts tool with:
      | key   | value             |
      | name  | New Contact       |
      | email | new@example.com   |
    Then the tool call should succeed
    And the tool result field "status" should equal "created"

  Scenario: MCP contacts_create fails without name
    When I call the "contacts_create" contacts tool with:
      | key   | value           |
      | email | no@example.com  |
    Then the tool call should fail
    And the tool error should contain "name is required"

  Scenario: MCP contacts_update tool
    Given the following contacts exist:
      | id   | name      | email              | phone    |
      | c-1  | Alice Doe | alice@example.com  | 555-0001 |
    When I call the "contacts_update" contacts tool with:
      | key  | value       |
      | id   | c-1         |
      | name | Alice Smith |
    Then the tool call should succeed
    And the tool result field "status" should equal "updated"

  Scenario: MCP contacts_update fails without id
    When I call the "contacts_update" contacts tool with:
      | key  | value       |
      | name | Alice Smith |
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP contacts_delete tool
    Given the following contacts exist:
      | id   | name      | email              | phone    |
      | c-1  | Alice Doe | alice@example.com  | 555-0001 |
    When I call the "contacts_delete" contacts tool with:
      | key | value |
      | id  | c-1   |
    Then the tool call should succeed
    And the tool result field "status" should equal "deleted"

  Scenario: MCP contacts_delete fails without id
    When I call the "contacts_delete" contacts tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP contacts_search tool
    Given the following contacts exist:
      | id   | name      | email              | phone    |
      | c-1  | Alice Doe | alice@example.com  | 555-0001 |
      | c-2  | Bob Smith | bob@example.com    | 555-0002 |
    When I call the "contacts_search" contacts tool with:
      | key   | value |
      | query | Alice |
    Then the tool call should succeed
    And the tool result should contain 1 items

  Scenario: MCP contacts_search fails without query
    When I call the "contacts_search" contacts tool
    Then the tool call should fail
    And the tool error should contain "query is required"
