Feature: Authentication token management
  As a Fastmail CLI user
  I want to manage my API tokens
  So that I can authenticate with the Fastmail API

  Background:
    Given a fresh auth store

  Scenario: Store and retrieve a token
    When I store the token "fmu1-test-token"
    Then the token operation should succeed
    When I retrieve the stored token
    Then the token should equal "fmu1-test-token"

  Scenario: Check token exists after storing
    When I store the token "fmu1-test-token"
    And I check if a token exists
    Then the token should exist

  Scenario: No token initially
    When I check if a token exists
    Then the token should not exist

  Scenario: Delete a stored token
    Given I store the token "fmu1-test-token"
    When I delete the stored token
    Then the token operation should succeed
    When I check if a token exists
    Then the token should not exist

  Scenario: Get token when none exists
    When I retrieve the stored token
    Then the token retrieval should fail
    And the token error should contain "no token found"

  Scenario: Token persists across store instances
    When I store the token "fmu1-persist-token"
    And I create a new store with the same config path
    And I retrieve the stored token
    Then the token should equal "fmu1-persist-token"
