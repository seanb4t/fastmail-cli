Feature: Masked email management
  As a Fastmail user
  I want to manage my masked email addresses
  So that I can protect my real email from spam

  Background:
    Given a connected Fastmail client

  Scenario: List masked emails
    Given the following masked emails exist:
      | id   | email                    | state   | forDomain    | description       |
      | me-1 | abc123@fastmail.com      | enabled | example.com  | Example signup    |
      | me-2 | def456@fastmail.com      | disabled| shopping.com | Shopping account  |
    When I list masked emails
    Then I should receive 2 masked emails
    And masked email 1 should have email "abc123@fastmail.com"
    And masked email 1 should have state "enabled"
    And masked email 1 should have forDomain "example.com"
    And masked email 2 should have email "def456@fastmail.com"
    And masked email 2 should have state "disabled"

  Scenario: Create a masked email
    When I create a masked email for domain "newsite.com" with description "New site signup"
    Then the masked email creation should succeed
    And the created masked email ID should not be empty

  Scenario: Disable a masked email
    Given the following masked emails exist:
      | id   | email                    | state   | forDomain   | description     |
      | me-1 | abc123@fastmail.com      | enabled | example.com | Example signup  |
    When I disable masked email "me-1"
    Then the masked email operation should succeed

  Scenario: Enable a masked email
    Given the following masked emails exist:
      | id   | email                    | state    | forDomain   | description     |
      | me-1 | abc123@fastmail.com      | disabled | example.com | Example signup  |
    When I enable masked email "me-1"
    Then the masked email operation should succeed

  Scenario: Delete a masked email
    Given the following masked emails exist:
      | id   | email                    | state   | forDomain   | description     |
      | me-1 | abc123@fastmail.com      | enabled | example.com | Example signup  |
    When I delete masked email "me-1"
    Then the masked email operation should succeed
