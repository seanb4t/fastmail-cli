Feature: Email export
  As a Fastmail user
  I want to export my emails in different formats
  So that I can back up or migrate my data

  Background:
    Given a connected Fastmail client

  Scenario: Export emails as JSONL
    Given the following emails exist:
      | id  | subject         | preview        |
      | e-1 | Hello World     | Preview text 1 |
      | e-2 | Important Email | Preview text 2 |
    When I export emails as "jsonl"
    Then the export should succeed
    And the export output should have 2 lines
    And the export output should contain "Hello World"
    And the export output should contain "Important Email"

  Scenario: Export emails as mbox
    Given the following emails exist:
      | id  | subject         | preview        |
      | e-1 | Hello World     | Preview text 1 |
      | e-2 | Important Email | Preview text 2 |
    When I export emails as "mbox"
    Then the export should succeed
    And the export output should contain "From "
    And the export output should contain "Subject: Hello World"
    And the export output should contain "Subject: Important Email"

  Scenario: Export emails as maildir
    Given the following emails exist:
      | id  | subject         | preview        |
      | e-1 | Hello World     | Preview text 1 |
      | e-2 | Important Email | Preview text 2 |
    When I export emails as "maildir"
    Then the export should succeed
    And the maildir should have 2 files

  Scenario: Export empty email list as JSONL
    When I export emails as "jsonl"
    Then the export should succeed
    And the export output should have 0 lines

  Scenario: Export JSONL contains email fields
    Given the following emails exist:
      | id  | subject           | preview            |
      | e-1 | Specific Subject  | Specific preview   |
    When I export emails as "jsonl"
    Then the export should succeed
    And the export output should have 1 lines
    And the export output should contain "Specific Subject"
