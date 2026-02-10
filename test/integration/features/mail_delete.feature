Feature: Mail delete
  As a Fastmail user
  I want to delete emails
  So that I can clean up my mailbox

  Background:
    Given a connected Fastmail client

  Scenario: Delete email moves to trash
    Given the following emails exist:
      | id      | subject       | preview    | read | flagged |
      | email-d1 | Delete Me    | preview... | true | false   |
    When I delete email "email-d1"
    Then the email delete should succeed

  Scenario: Delete email in trash permanently destroys
    Given an email "email-d2" exists in trash with subject "Already Trashed"
    When I delete email "email-d2"
    Then the email delete should succeed
