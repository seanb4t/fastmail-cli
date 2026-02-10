Feature: Mail reply
  As a Fastmail user
  I want to reply to emails
  So that I can respond to conversations

  Background:
    Given a connected Fastmail client

  Scenario: Reply to an email
    Given an email exists for reply:
      | id   | subject    | from_email        | from_name | messageId           |
      | e-r1 | Hello      | alice@example.com | Alice     | <msg1@example.com>  |
    When I reply to email "e-r1" with body "Thanks for your message"
    Then the reply should succeed
    And the reply email ID should not be empty

  Scenario: Reply all to an email
    Given an email exists for reply with cc:
      | id   | subject    | from_email        | from_name | to_email          | cc_email         | messageId           |
      | e-r2 | Team Chat  | alice@example.com | Alice     | test@example.com  | bob@example.com  | <msg2@example.com>  |
    When I reply all to email "e-r2" with body "Sounds good to me"
    Then the reply should succeed
    And the reply email ID should not be empty
