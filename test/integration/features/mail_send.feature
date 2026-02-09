Feature: Mail send
  As a Fastmail user
  I want to send emails
  So that I can communicate with others

  Background:
    Given a connected Fastmail client

  Scenario: Send a simple email
    When I send an email to "recipient@example.com" with subject "Test Subject" and body "Hello there"
    Then the send should succeed
    And the sent email ID should not be empty

  Scenario: Send a scheduled email
    When I send a scheduled email to "recipient@example.com" with subject "Later" and body "Delayed" at "2024-06-15T14:00:00Z"
    Then the send should succeed
    And the sent email ID should not be empty

  Scenario: Flag an email
    Given the following emails exist:
      | id       | subject  | preview | read | flagged |
      | email-f1 | Flag Me  | ...     | true | false   |
    When I set keyword "$flagged" to true on email "email-f1"
    Then the flag operation should succeed

  Scenario: Move an email to archive
    Given the following emails exist:
      | id       | subject | preview | read | flagged |
      | email-m1 | Move Me | ...     | true | false   |
    When I move email "email-m1" to "Archive"
    Then the move operation should succeed
