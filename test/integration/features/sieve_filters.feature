Feature: Sieve filters
  As a Fastmail user
  I want to manage sieve filter scripts
  So that I can automate email processing

  Background:
    Given a connected Fastmail client

  Scenario: List Sieve scripts
    Given the following sieve scripts exist:
      | id       | name          | isActive |
      | sieve-1  | Spam Filter   | true     |
      | sieve-2  | Auto Archive  | false    |
    When I list sieve scripts
    Then I should receive 2 sieve scripts
    And sieve script 1 should have name "Spam Filter"
    And sieve script 1 should be active
    And sieve script 2 should have name "Auto Archive"
    And sieve script 2 should not be active

  Scenario: Get a specific Sieve script
    Given the following sieve scripts exist:
      | id       | name          | isActive |
      | sieve-1  | Spam Filter   | true     |
    When I get sieve script "sieve-1"
    Then the sieve script should have name "Spam Filter"
    And the sieve script should be active

  Scenario: Create a Sieve script
    When I create a sieve script with name "New Filter" and script "require \"fileinto\"; if header :contains \"Subject\" \"test\" { fileinto \"Tests\"; }"
    Then the sieve operation should succeed
    And the created sieve script ID should not be empty

  Scenario: Activate a Sieve script
    Given the following sieve scripts exist:
      | id       | name          | isActive |
      | sieve-1  | Spam Filter   | false    |
    When I activate sieve script "sieve-1"
    Then the sieve operation should succeed

  Scenario: Deactivate a Sieve script
    Given the following sieve scripts exist:
      | id       | name          | isActive |
      | sieve-1  | Spam Filter   | true     |
    When I deactivate sieve script "sieve-1"
    Then the sieve operation should succeed

  Scenario: Validate a valid Sieve script
    When I validate sieve script "require \"fileinto\"; if header :contains \"Subject\" \"test\" { fileinto \"Tests\"; }"
    Then the sieve validation should be valid

  Scenario: Delete a Sieve script
    Given the following sieve scripts exist:
      | id       | name          | isActive |
      | sieve-1  | Spam Filter   | false    |
    When I delete sieve script "sieve-1"
    Then the sieve operation should succeed
