Feature: Mail import
  As a Fastmail user
  I want to import email messages
  So that I can add messages from external sources

  Background:
    Given a connected Fastmail client

  Scenario: Import an email message
    Given a blob upload server is configured
    When I import an email with content "From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Imported\r\n\r\nBody" to folder "Inbox"
    Then the import should succeed
    And the imported email ID should not be empty
