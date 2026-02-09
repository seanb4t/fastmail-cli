Feature: Identity management
  As a Fastmail user
  I want to manage my sender identities
  So that I can control how my emails appear to recipients

  Background:
    Given a connected Fastmail client

  Scenario: List identities
    Given the following identities exist:
      | id        | name       | email              |
      | identity1 | Test User  | test@example.com   |
      | identity2 | Work Alias | work@example.com   |
    When I list all identities
    Then I should receive 2 identities
    And identity 1 should have name "Test User"
    And identity 1 should have email "test@example.com"
    And identity 2 should have name "Work Alias"
    And identity 2 should have email "work@example.com"

  Scenario: Update identity display name
    Given the following identities exist:
      | id        | name      | email            |
      | identity1 | Test User | test@example.com |
    When I update identity "identity1" with name "New Display Name"
    Then the identity update should succeed
    When I list all identities
    Then I should receive 1 identity
    And identity 1 should have name "New Display Name"

  Scenario: Update identity signature
    Given the following identities exist:
      | id        | name      | email            |
      | identity1 | Test User | test@example.com |
    When I update identity "identity1" with signature "-- Sent from CLI"
    Then the identity update should succeed
    When I list all identities
    Then I should receive 1 identity
    And identity 1 should have signature "-- Sent from CLI"
