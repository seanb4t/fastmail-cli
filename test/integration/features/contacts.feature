Feature: Contact management
  As a Fastmail user
  I want to manage my contacts via CardDAV
  So that I can keep my address book organized

  Background:
    Given a connected Fastmail client

  Scenario: List contacts
    Given the following contacts exist:
      | id | name          | email              | phone    |
      | c1 | Alice Wonder  | alice@example.com  | 555-1234 |
      | c2 | Bob Builder   | bob@example.com    | 555-5678 |
    When I list all contacts
    Then I should receive 2 contacts
    And contact 1 should have name "Alice Wonder"
    And contact 1 should have email "alice@example.com"
    And contact 2 should have name "Bob Builder"
    And contact 2 should have email "bob@example.com"

  Scenario: Search contacts by name
    Given the following contacts exist:
      | id | name            | email                | phone    |
      | c1 | Alice Wonder    | alice@example.com    | 555-1234 |
      | c2 | Bob Builder     | bob@example.com      | 555-5678 |
      | c3 | Charlie Checker | charlie@example.com  | 555-9012 |
    When I search contacts for "Alice"
    Then I should receive 1 contacts
    And contact 1 should have name "Alice Wonder"

  Scenario: Search contacts by email
    Given the following contacts exist:
      | id | name          | email              | phone    |
      | c1 | Alice Wonder  | alice@example.com  | 555-1234 |
      | c2 | Bob Builder   | bob@example.com    | 555-5678 |
    When I search contacts for "bob@"
    Then I should receive 1 contacts
    And contact 1 should have name "Bob Builder"
    And contact 1 should have email "bob@example.com"

  Scenario: Get single contact by ID
    Given the following contacts exist:
      | id | name          | email              | phone    |
      | c1 | Alice Wonder  | alice@example.com  | 555-1234 |
      | c2 | Bob Builder   | bob@example.com    | 555-5678 |
    When I get contact "c1"
    Then the contact should have name "Alice Wonder"
    And the contact should have email "alice@example.com"
    And the contact should have phone "555-1234"

  Scenario: List empty contacts
    Given the following contacts exist:
      | id | name | email | phone |
    When I list all contacts
    Then I should receive 0 contacts

  Scenario: Get nonexistent contact
    Given the following contacts exist:
      | id | name          | email              | phone    |
      | c1 | Alice Wonder  | alice@example.com  | 555-1234 |
    When I get contact "nonexistent"
    Then the contact operation should have failed
