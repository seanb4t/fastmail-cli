Feature: Get single email with full content
  As a Fastmail user
  I want to view a single email with its full body and sender details
  So that I can read the complete message

  Background:
    Given a connected Fastmail client

  Scenario: Get email returns body and sender
    Given the following emails exist:
      | id   | subject       | from_name  | from_email       | body                  |
      | e-1  | Hello World   | Alice      | alice@example.com | This is the full body |
    When I get email "e-1"
    Then the email should have subject "Hello World"
    And the email should have from "alice@example.com"
    And the email should have body containing "full body"
    And the email body should not be empty

  Scenario: List emails includes sender
    Given the following emails exist:
      | id   | subject       | from_name  | from_email       |
      | e-1  | Hello World   | Alice      | alice@example.com |
      | e-2  | Second Email  | Bob        | bob@example.com   |
    When I list emails from "inbox" with limit 10
    Then I should receive 2 emails
    And email 1 should have from "alice@example.com"
    And email 2 should have from "bob@example.com"
