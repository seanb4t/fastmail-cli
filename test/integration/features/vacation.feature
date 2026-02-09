Feature: Vacation response
  As a Fastmail user
  I want to manage my vacation auto-reply
  So that people know when I am out of office

  Background:
    Given a connected Fastmail client

  Scenario: Get vacation status when disabled
    Given the vacation response is disabled
    When I get the vacation status
    Then the vacation response should be disabled
    And the vacation subject should be ""
    And the vacation body should be ""

  Scenario: Enable vacation response
    Given the vacation response is disabled
    When I enable the vacation response with subject "Out of Office" and body "I am away until Monday."
    Then the vacation operation should succeed
    When I get the vacation status
    Then the vacation response should be enabled
    And the vacation subject should be "Out of Office"
    And the vacation body should be "I am away until Monday."

  Scenario: Disable vacation response
    Given the vacation response is enabled with subject "On Holiday" and body "Back in two weeks."
    When I disable the vacation response
    Then the vacation operation should succeed
    When I get the vacation status
    Then the vacation response should be disabled
