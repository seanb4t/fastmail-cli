Feature: MCP filter tools
  As an AI agent using MCP
  I want to use filter tools to manage Sieve scripts
  So that I can help users automate email processing

  Background:
    Given a connected MCP tools config

  Scenario: List filter scripts
    Given the following sieve scripts exist:
      | id      | name         | isActive |
      | sieve-1 | Spam Filter  | true     |
      | sieve-2 | Auto Archive | false    |
    When I call the "filter_list" tool
    Then the tool call should succeed
    And the tool result should contain 2 items
    And the tool result item 1 field "id" should equal "sieve-1"
    And the tool result item 1 field "name" should equal "Spam Filter"
    And the tool result item 1 field "is_active" should equal "true"
    And the tool result item 2 field "id" should equal "sieve-2"
    And the tool result item 2 field "name" should equal "Auto Archive"
    And the tool result item 2 field "is_active" should equal "false"

  Scenario: Get a specific filter script
    Given the following sieve scripts exist:
      | id      | name        | isActive |
      | sieve-1 | Spam Filter | true     |
    When I call the "filter_get" tool with:
      | key | value   |
      | id  | sieve-1 |
    Then the tool call should succeed
    And the tool result field "id" should equal "sieve-1"
    And the tool result field "name" should equal "Spam Filter"
    And the tool result field "is_active" should be true

  Scenario: Create a filter script
    When I call the "filter_create" tool with:
      | key    | value                                    |
      | name   | New Filter                               |
      | script | require "fileinto"; fileinto "Archive";  |
    Then the tool call should succeed
    And the tool result field "name" should equal "New Filter"
    And the tool result field "status" should equal "created"

  Scenario: Activate a filter script
    Given the following sieve scripts exist:
      | id      | name        | isActive |
      | sieve-1 | Spam Filter | false    |
    When I call the "filter_activate" tool with:
      | key | value   |
      | id  | sieve-1 |
    Then the tool call should succeed
    And the tool result field "id" should equal "sieve-1"
    And the tool result field "status" should equal "activated"

  Scenario: Deactivate a filter script
    Given the following sieve scripts exist:
      | id      | name        | isActive |
      | sieve-1 | Spam Filter | true     |
    When I call the "filter_deactivate" tool with:
      | key | value   |
      | id  | sieve-1 |
    Then the tool call should succeed
    And the tool result field "id" should equal "sieve-1"
    And the tool result field "status" should equal "deactivated"

  Scenario: Validate a valid sieve script
    When I call the "filter_validate" tool with:
      | key    | value                                    |
      | script | require "fileinto"; fileinto "Archive";  |
    Then the tool call should succeed
    And the tool result field "is_valid" should be true

  Scenario: Delete a filter script
    Given the following sieve scripts exist:
      | id      | name        | isActive |
      | sieve-1 | Spam Filter | false    |
    When I call the "filter_delete" tool with:
      | key | value   |
      | id  | sieve-1 |
    Then the tool call should succeed
    And the tool result field "id" should equal "sieve-1"
    And the tool result field "status" should equal "deleted"

  Scenario: Get filter fails without id
    When I call the "filter_get" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: Create filter fails without name
    When I call the "filter_create" tool with:
      | key    | value                                   |
      | script | require "fileinto"; fileinto "Archive"; |
    Then the tool call should fail
    And the tool error should contain "name is required"

  Scenario: Create filter fails without script
    When I call the "filter_create" tool with:
      | key  | value      |
      | name | New Filter |
    Then the tool call should fail
    And the tool error should contain "script is required"

  Scenario: Activate filter fails without id
    When I call the "filter_activate" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: Deactivate filter fails without id
    When I call the "filter_deactivate" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: Validate filter fails without script
    When I call the "filter_validate" tool
    Then the tool call should fail
    And the tool error should contain "script is required"

  Scenario: Delete filter fails without id
    When I call the "filter_delete" tool
    Then the tool call should fail
    And the tool error should contain "id is required"
