Feature: MCP account tools
  As an AI agent using MCP
  I want to use account tools to check Fastmail account info
  So that I can help users understand their account usage

  Background:
    Given a connected MCP tools config

  Scenario: Get storage quota via MCP quota_get tool
    Given the account has storage quota of 1073741824 bytes used and 5368709120 bytes limit
    When I call the "quota_get" tool
    Then the tool call should succeed
    And the tool result field "used" should equal "1073741824"
    And the tool result field "limit" should equal "5368709120"
    And the tool result field "used_percent" should equal "20"
    And the tool result field "used_display" should equal "1.0 GB"
    And the tool result field "limit_display" should equal "5.0 GB"
