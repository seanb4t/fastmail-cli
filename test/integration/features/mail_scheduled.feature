Feature: Scheduled email management
  As a Fastmail user
  I want to manage scheduled emails
  So that I can view and cancel pending sends

  Background:
    Given a connected Fastmail client

  Scenario: List scheduled emails
    Given the following scheduled submissions exist:
      | id    | emailId  | sendAt                   | undoStatus |
      | sub-1 | email-s1 | 2024-06-15T14:00:00Z     | pending    |
      | sub-2 | email-s2 | 2024-06-16T10:00:00Z     | pending    |
    And the following emails exist:
      | id       | subject       | preview | read | flagged |
      | email-s1 | Meeting Later | ...     | true | false   |
      | email-s2 | Reminder      | ...     | true | false   |
    When I list scheduled emails with limit 10
    Then I should receive 2 scheduled emails
    And scheduled email 1 should have subject "Meeting Later"
    And scheduled email 1 should have status "pending"

  Scenario: Cancel a scheduled email
    Given the following scheduled submissions exist:
      | id    | emailId  | sendAt               | undoStatus |
      | sub-1 | email-c1 | 2024-06-15T14:00:00Z | pending    |
    When I cancel scheduled email "sub-1"
    Then the cancel operation should succeed
