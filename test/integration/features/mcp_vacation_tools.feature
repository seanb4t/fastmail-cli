Feature: MCP vacation tools
  As an AI agent using MCP
  I want to use vacation tools to manage Fastmail out-of-office replies
  So that I can help users configure their vacation auto-responder

  Background:
    Given a connected MCP tools config

  Scenario: Get vacation status when disabled
    Given the vacation response is disabled
    When I call the "vacation_status" tool
    Then the tool call should succeed
    And the tool result field "is_enabled" should be false
    And the tool result field "subject" should equal ""
    And the tool result field "text_body" should equal ""

  Scenario: Get vacation status when enabled
    Given the vacation response is enabled with subject "Out of Office" and body "I am away until Monday."
    When I call the "vacation_status" tool
    Then the tool call should succeed
    And the tool result field "is_enabled" should be true
    And the tool result field "subject" should equal "Out of Office"
    And the tool result field "text_body" should equal "I am away until Monday."

  Scenario: Enable vacation response via vacation_set
    Given the vacation response is disabled
    When I call the "vacation_set" tool with:
      | key     | value                   |
      | enabled | true                    |
      | subject | On Holiday              |
      | body    | Back in two weeks.      |
    Then the tool call should succeed
    And the tool result field "status" should equal "enabled"

  Scenario: Disable vacation response via vacation_set
    Given the vacation response is enabled with subject "On Holiday" and body "Back in two weeks."
    When I call the "vacation_set" tool with:
      | key     | value |
      | enabled | false |
    Then the tool call should succeed
    And the tool result field "status" should equal "disabled"

  Scenario: Enable vacation response without subject and body should fail
    Given the vacation response is disabled
    When I call the "vacation_set" tool with:
      | key     | value |
      | enabled | true  |
    Then the tool call should fail
    And the tool error should contain "subject and body are required"
