Feature: Mail thread retrieval
  As a Fastmail user
  I want to view all emails in a conversation thread
  So that I can follow the full discussion

  Background:
    Given a connected Fastmail client

  Scenario: Get thread with multiple emails
    Given a thread "thread-1" with the following emails:
      | id   | subject          | from_email        | receivedAt               |
      | e-t1 | Original Message | alice@example.com | 2024-01-15T10:00:00Z     |
      | e-t2 | Re: Original     | bob@example.com   | 2024-01-15T11:00:00Z     |
      | e-t3 | Re: Re: Original | alice@example.com | 2024-01-15T12:00:00Z     |
    When I get thread "thread-1"
    Then I should receive 3 thread emails
    And thread email 1 should have subject "Original Message"
    And thread email 2 should have subject "Re: Original"
    And thread email 3 should have subject "Re: Re: Original"

  Scenario: Get thread with single email
    Given a thread "thread-solo" with the following emails:
      | id     | subject    | from_email        | receivedAt           |
      | e-solo | Solo Email | alice@example.com | 2024-01-15T10:00:00Z |
    When I get thread "thread-solo"
    Then I should receive 1 thread email
    And thread email 1 should have subject "Solo Email"
