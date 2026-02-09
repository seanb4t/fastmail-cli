Feature: MCP mail tools
  As an AI agent using MCP
  I want to use mail tools to interact with Fastmail
  So that I can help users manage their email

  Background:
    Given a connected MCP tools config

  Scenario: MCP mail_list tool
    Given the inbox contains the following emails:
      | id     | subject       | preview           | read  | flagged |
      | email1 | Hello World   | Test email        | true  | false   |
      | email2 | Another Email | More content here | false | true    |
    When I call the "mail_list" tool with:
      | key    | value |
      | folder | Inbox |
      | limit  | 10    |
    Then the tool call should succeed
    And the tool result should contain 2 items

  Scenario: MCP mail_get tool
    Given the following emails exist:
      | id        | subject      | preview     | read | flagged |
      | email-get | Test Subject | Test preview| true | true    |
    When I call the "mail_get" tool with:
      | key | value     |
      | id  | email-get |
    Then the tool call should succeed

  Scenario: MCP mail_search tool
    Given the following emails exist:
      | id          | subject       | preview                    | read | flagged |
      | email-match | Meeting Notes | Notes from today's meeting | true | false   |
    When I call the "mail_search" tool with:
      | key   | value         |
      | query | meeting notes |
      | limit | 10            |
    Then the tool call should succeed
    And the tool result should contain 1 items

  Scenario: MCP mail_flag tool
    Given the following emails exist:
      | id       | subject | preview | read | flagged |
      | email-f1 | Flag Me | ...     | true | false   |
    When I call the "mail_flag" tool with keywords:
      | id       | email-f1 |
      | $flagged | true     |
    Then the tool call should succeed
