Feature: Mail listing
  As a Fastmail user
  I want to list emails from my mailboxes
  So that I can see what messages I have

  Background:
    Given a connected Fastmail client

  Scenario: List inbox emails
    Given the inbox contains the following emails:
      | id     | subject       | preview              | read  | flagged |
      | email1 | Hello World   | This is a test email | true  | false   |
      | email2 | Another Email | More content here    | false | true    |
    When I list emails from "Inbox" with limit 10
    Then I should receive 2 emails
    And email 1 should have subject "Hello World"
    And email 1 should be read
    And email 2 should have subject "Another Email"
    And email 2 should be flagged

  Scenario: List with limit
    Given the inbox contains the following emails:
      | id     | subject   | preview | read  | flagged |
      | email1 | First     | ...     | true  | false   |
      | email2 | Second    | ...     | true  | false   |
      | email3 | Third     | ...     | true  | false   |
    When I list emails from "Inbox" with limit 2
    Then I should receive 2 emails

  Scenario: List empty folder
    Given the folder "Archive" contains no emails
    When I list emails from "Archive" with limit 10
    Then I should receive 0 emails
