Feature: MCP mail scheduled send tools
  As an AI agent using MCP
  I want to list and cancel scheduled email deliveries
  So that I can help users manage their delayed sends

  Background:
    Given a connected MCP tools config

  Scenario: MCP mail_scheduled tool lists scheduled sends
    Given the following scheduled emails exist for MCP:
      | id     | emailId  | sendAt                   | undoStatus | subject        | to                |
      | sub-1  | email-s1 | 2025-06-15T14:00:00Z     | pending    | Weekly Report  | alice@example.com |
      | sub-2  | email-s2 | 2025-06-16T09:00:00Z     | pending    | Follow Up      | bob@example.com   |
    When I call the "mail_scheduled" tool
    Then the tool call should succeed
    And the tool result should contain 2 items
    And the tool result item 1 field "submission_id" should equal "sub-1"
    And the tool result item 1 field "subject" should equal "Weekly Report"
    And the tool result item 1 field "status" should equal "pending"
    And the tool result item 2 field "submission_id" should equal "sub-2"
    And the tool result item 2 field "subject" should equal "Follow Up"

  Scenario: MCP mail_scheduled tool respects limit
    Given the following scheduled emails exist for MCP:
      | id     | emailId  | sendAt                   | undoStatus | subject   | to                |
      | sub-1  | email-s1 | 2025-06-15T14:00:00Z     | pending    | First     | alice@example.com |
      | sub-2  | email-s2 | 2025-06-16T09:00:00Z     | pending    | Second    | bob@example.com   |
      | sub-3  | email-s3 | 2025-06-17T09:00:00Z     | pending    | Third     | carol@example.com |
    When I call the "mail_scheduled" tool with:
      | key   | value |
      | limit | 2     |
    Then the tool call should succeed
    And the tool result should contain 3 items

  Scenario: MCP mail_scheduled tool with no scheduled sends
    When I call the "mail_scheduled" tool
    Then the tool call should succeed
    And the tool result should contain 0 items

  Scenario: MCP mail_scheduled_cancel tool cancels a scheduled send
    Given the following scheduled emails exist for MCP:
      | id     | emailId  | sendAt                   | undoStatus | subject       | to                |
      | sub-1  | email-s1 | 2025-06-15T14:00:00Z     | pending    | Cancel This   | alice@example.com |
    When I call the "mail_scheduled_cancel" tool with:
      | key           | value |
      | submission_id | sub-1 |
    Then the tool call should succeed
    And the tool result field "submission_id" should equal "sub-1"
    And the tool result field "status" should equal "canceled"

  Scenario: MCP mail_scheduled_cancel tool fails without submission_id
    When I call the "mail_scheduled_cancel" tool
    Then the tool call should fail
    And the tool error should contain "submission_id is required"
