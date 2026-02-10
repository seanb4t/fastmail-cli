Feature: MCP masked email tools
  As an AI agent using MCP
  I want to use masked email tools to interact with Fastmail
  So that I can help users manage their masked email addresses

  Background:
    Given a connected MCP tools config

  Scenario: MCP masked_email_list tool
    Given the following masked emails exist:
      | id   | email              | state   | forDomain   | description    |
      | me1  | abc@fastmail.com   | enabled | example.com | Test alias     |
      | me2  | def@fastmail.com   | pending | other.org   | Another alias  |
    When I call the "masked_email_list" tool
    Then the tool call should succeed
    And the tool result should contain 2 items
    And the tool result item 1 field "id" should equal "me1"
    And the tool result item 1 field "email" should equal "abc@fastmail.com"
    And the tool result item 1 field "for_domain" should equal "example.com"
    And the tool result item 2 field "id" should equal "me2"
    And the tool result item 2 field "state" should equal "pending"

  Scenario: MCP masked_email_get tool
    Given the following masked emails exist:
      | id   | email              | state   | forDomain   | description   | createdBy |
      | me1  | abc@fastmail.com   | enabled | example.com | Test alias    | user      |
    When I call the "masked_email_get" tool with:
      | key | value |
      | id  | me1   |
    Then the tool call should succeed
    And the tool result field "id" should equal "me1"
    And the tool result field "email" should equal "abc@fastmail.com"
    And the tool result field "state" should equal "enabled"
    And the tool result field "for_domain" should equal "example.com"
    And the tool result field "description" should equal "Test alias"

  Scenario: MCP masked_email_get fails without id
    When I call the "masked_email_get" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP masked_email_create tool
    When I call the "masked_email_create" tool with:
      | key         | value       |
      | domain      | example.com |
      | description | Shopping    |
    Then the tool call should succeed
    And the tool result field "for_domain" should equal "example.com"
    And the tool result field "description" should equal "Shopping"

  Scenario: MCP masked_email_enable tool
    Given the following masked emails exist:
      | id   | email              | state    | forDomain   | description |
      | me1  | abc@fastmail.com   | disabled | example.com | Test alias  |
    When I call the "masked_email_enable" tool with:
      | key | value |
      | id  | me1   |
    Then the tool call should succeed
    And the tool result field "id" should equal "me1"
    And the tool result field "status" should equal "enabled"

  Scenario: MCP masked_email_enable fails without id
    When I call the "masked_email_enable" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP masked_email_disable tool
    Given the following masked emails exist:
      | id   | email              | state   | forDomain   | description |
      | me1  | abc@fastmail.com   | enabled | example.com | Test alias  |
    When I call the "masked_email_disable" tool with:
      | key | value |
      | id  | me1   |
    Then the tool call should succeed
    And the tool result field "id" should equal "me1"
    And the tool result field "status" should equal "disabled"

  Scenario: MCP masked_email_disable fails without id
    When I call the "masked_email_disable" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP masked_email_delete tool
    Given the following masked emails exist:
      | id   | email              | state   | forDomain   | description |
      | me1  | abc@fastmail.com   | enabled | example.com | Test alias  |
    When I call the "masked_email_delete" tool with:
      | key | value |
      | id  | me1   |
    Then the tool call should succeed
    And the tool result field "id" should equal "me1"
    And the tool result field "status" should equal "deleted"

  Scenario: MCP masked_email_delete fails without id
    When I call the "masked_email_delete" tool
    Then the tool call should fail
    And the tool error should contain "id is required"
