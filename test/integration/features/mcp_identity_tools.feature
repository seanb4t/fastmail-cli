Feature: MCP identity tools
  As an AI agent using MCP
  I want to use identity tools to interact with Fastmail
  So that I can help users manage their sender identities

  Background:
    Given a connected MCP tools config

  Scenario: List identities via MCP tool
    Given the following identities exist:
      | id        | name       | email              |
      | identity1 | Test User  | test@example.com   |
      | identity2 | Work Alias | work@example.com   |
    When I call the "identity_list" tool
    Then the tool call should succeed
    And the tool result should contain 2 items
    And the tool result item 1 field "id" should equal "identity1"
    And the tool result item 1 field "name" should equal "Test User"
    And the tool result item 1 field "email" should equal "test@example.com"
    And the tool result item 2 field "id" should equal "identity2"
    And the tool result item 2 field "name" should equal "Work Alias"
    And the tool result item 2 field "email" should equal "work@example.com"

  Scenario: Update identity name via MCP tool
    Given the following identities exist:
      | id        | name      | email            |
      | identity1 | Test User | test@example.com |
    When I call the "identity_set" tool with:
      | key  | value            |
      | id   | identity1        |
      | name | New Display Name |
    Then the tool call should succeed
    And the tool result field "id" should equal "identity1"
    And the tool result field "status" should equal "updated"

  Scenario: Update identity fails without id
    When I call the "identity_set" tool with:
      | key  | value            |
      | name | New Display Name |
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: Update identity fails without any fields to update
    When I call the "identity_set" tool with:
      | key | value     |
      | id  | identity1 |
    Then the tool call should fail
    And the tool error should contain "at least one of"
