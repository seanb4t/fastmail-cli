Feature: MCP mail mutation tools
  As an AI agent using MCP
  I want to use mail mutation tools to manage emails
  So that I can help users send, reply, move, and delete emails

  Background:
    Given a connected MCP tools config

  # mail_send

  Scenario: MCP mail_send tool sends an email
    When I call the "mail_send" tool with args:
      | key     | value                |
      | to      | recipient@test.com   |
      | subject | Test Subject         |
      | body    | Hello from the test  |
    Then the tool call should succeed
    And the tool result field "status" should equal "sent"

  Scenario: MCP mail_send tool fails without to
    When I call the "mail_send" tool with:
      | key     | value        |
      | subject | Test Subject |
      | body    | Hello        |
    Then the tool call should fail
    And the tool error should contain "to is required"

  Scenario: MCP mail_send tool fails without subject
    When I call the "mail_send" tool with args:
      | key | value              |
      | to  | recipient@test.com |
      | body| Hello              |
    Then the tool call should fail
    And the tool error should contain "subject is required"

  Scenario: MCP mail_send tool fails without body
    When I call the "mail_send" tool with args:
      | key     | value              |
      | to      | recipient@test.com |
      | subject | Test Subject       |
    Then the tool call should fail
    And the tool error should contain "body is required"

  # mail_reply

  Scenario: MCP mail_reply tool replies to an email
    Given an email exists for reply:
      | id       | subject       | from_name | from_email        | messageId        |
      | reply-e1 | Original Subj | Alice     | alice@example.com | <msg1@test.com>  |
    When I call the "mail_reply" tool with:
      | key      | value        |
      | email_id | reply-e1     |
      | body     | Thanks Alice |
    Then the tool call should succeed
    And the tool result field "status" should equal "sent"

  Scenario: MCP mail_reply tool with reply_all
    Given an email exists for reply with cc:
      | id       | subject | from_name | from_email        | to_email          | cc_email        | messageId       |
      | reply-e2 | Thread  | Bob       | bob@example.com   | test@example.com  | cc@example.com  | <msg2@test.com> |
    When I call the "mail_reply" tool with:
      | key       | value        |
      | email_id  | reply-e2     |
      | body      | Reply to all |
      | reply_all | true         |
    Then the tool call should succeed
    And the tool result field "status" should equal "sent"

  Scenario: MCP mail_reply tool fails without email_id
    When I call the "mail_reply" tool with:
      | key  | value |
      | body | Hello |
    Then the tool call should fail
    And the tool error should contain "email_id is required"

  Scenario: MCP mail_reply tool fails without body
    When I call the "mail_reply" tool with:
      | key      | value    |
      | email_id | reply-e1 |
    Then the tool call should fail
    And the tool error should contain "body is required"

  # mail_move

  Scenario: MCP mail_move tool moves an email
    Given the following emails exist:
      | id      | subject    | preview       | read | flagged |
      | move-e1 | Move Email | Move this one | true | false   |
    When I call the "mail_move" tool with:
      | key    | value   |
      | id     | move-e1 |
      | folder | Archive |
    Then the tool call should succeed
    And the tool result field "id" should equal "move-e1"
    And the tool result field "folder" should equal "Archive"
    And the tool result field "status" should equal "moved"

  Scenario: MCP mail_move tool fails without id
    When I call the "mail_move" tool with:
      | key    | value   |
      | folder | Archive |
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP mail_move tool fails without folder
    When I call the "mail_move" tool with:
      | key | value   |
      | id  | move-e1 |
    Then the tool call should fail
    And the tool error should contain "folder is required"

  # mail_delete

  Scenario: MCP mail_delete tool deletes an email
    Given the following emails exist:
      | id     | subject      | preview         | read | flagged |
      | del-e1 | Delete Email | Delete this one | true | false   |
    When I call the "mail_delete" tool with:
      | key | value  |
      | id  | del-e1 |
    Then the tool call should succeed
    And the tool result field "id" should equal "del-e1"
    And the tool result field "status" should equal "deleted"

  Scenario: MCP mail_delete tool fails without id
    When I call the "mail_delete" tool
    Then the tool call should fail
    And the tool error should contain "id is required"
