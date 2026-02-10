Feature: MCP tool error handling
  As an AI agent using MCP
  I want MCP tools to return clear errors for invalid input
  So that I can provide helpful feedback to users

  Background:
    Given a connected MCP tools config

  # --- mail_get ---

  Scenario: mail_get fails without id
    When I call the "mail_get" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  # --- mail_search ---

  Scenario: mail_search fails without query or filters
    When I call the "mail_search" tool
    Then the tool call should fail
    And the tool error should contain "query or at least one filter is required"

  # --- mail_flag ---

  Scenario: mail_flag fails without id
    When I call the "mail_flag" tool with keywords:
      | $flagged | true |
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: mail_flag fails without keywords
    When I call the "mail_flag" tool with:
      | key | value    |
      | id  | email-f1 |
    Then the tool call should fail
    And the tool error should contain "keywords is required"

  # --- mail_scheduled_cancel ---

  Scenario: mail_scheduled_cancel fails without submission_id
    When I call the "mail_scheduled_cancel" tool
    Then the tool call should fail
    And the tool error should contain "submission_id is required"
