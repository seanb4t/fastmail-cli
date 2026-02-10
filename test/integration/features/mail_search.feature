Feature: Mail search
  As a Fastmail user
  I want to search my emails
  So that I can find specific messages

  Background:
    Given a connected Fastmail client

  Scenario: Simple text search
    Given the following emails exist:
      | id          | subject       | preview                       | read | flagged |
      | email-match | Meeting Notes | Notes from today's meeting    | true | false   |
    When I search for "meeting notes" with limit 10
    Then I should receive 1 emails
    And email 1 should have subject "Meeting Notes"

  Scenario: Search with no results
    When I search for "nonexistent query" with limit 10
    Then I should receive 0 emails

  Scenario: Advanced search by sender
    Given the following emails exist:
      | id      | subject    | preview      | read  | flagged |
      | email-a | From Alice | Hello there  | false | false   |
    When I search with filter from "alice@example.com" with limit 10
    Then I should receive 1 emails
    And email 1 should have subject "From Alice"

  Scenario: Search with snippets
    Given the following emails exist:
      | id      | subject       | preview                    | read | flagged |
      | email-1 | Meeting Notes | Notes from today's meeting | true | false   |
    When I search for "meeting notes" with snippets and limit 10
    Then I should receive 1 search results with snippets
    And search result 1 should have a preview snippet

  Scenario: Advanced search with snippets
    Given the following emails exist:
      | id      | subject       | preview                    | read | flagged |
      | email-1 | Meeting Notes | Notes from today's meeting | true | false   |
    When I search advanced with snippets for from "alice@example.com" with limit 10
    Then I should receive 1 search results with snippets
    And search result 1 should have a preview snippet
