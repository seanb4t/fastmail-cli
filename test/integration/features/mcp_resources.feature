Feature: MCP resource reading
  As an AI agent using MCP
  I want to read Fastmail resources via URI patterns
  So that I can access mailbox data in a read-only fashion

  Background:
    Given a connected Fastmail client

  Scenario: Read inbox resource with emails
    Given the following emails exist:
      | id   | subject          | preview              | read  | flagged |
      | e-1  | Weekly Report    | Summary of the week  | true  | false   |
      | e-2  | Meeting Invite   | Join us on Friday    | false | true    |
    When I read the MCP resource "fastmail://inbox"
    Then the resource read should succeed
    And the resource content should contain "Weekly Report"
    And the resource content should contain "Meeting Invite"

  Scenario: Read single email resource
    Given the following emails exist:
      | id   | subject         | preview                 | read  | flagged |
      | e-1  | Important Note  | Please review this doc  | false | false   |
    When I read the MCP resource "fastmail://mail/e-1"
    Then the resource read should succeed
    And the resource content should contain "Important Note"
    And the resource content should contain "Please review this doc"

  Scenario: Read masked emails resource
    Given the following masked emails exist:
      | id   | email                   | state   | forDomain    | description      |
      | me-1 | abc123@fastmail.com     | enabled | example.com  | Example signup   |
      | me-2 | def456@fastmail.com     | enabled | shopping.com | Shopping account |
    When I read the MCP resource "fastmail://masked-emails"
    Then the resource read should succeed
    And the resource content should contain "abc123@fastmail.com"
    And the resource content should contain "def456@fastmail.com"

  Scenario: Read single masked email resource
    Given the following masked emails exist:
      | id   | email                   | state   | forDomain   | description    |
      | me-1 | abc123@fastmail.com     | enabled | example.com | Example signup |
    When I read the MCP resource "fastmail://masked-email/me-1"
    Then the resource read should succeed
    And the resource content should contain "abc123@fastmail.com"
    And the resource content should contain "example.com"

  Scenario: Read contacts resource returns CardDAV message
    When I read the MCP resource "fastmail://contacts"
    Then the resource read should succeed
    And the resource content should contain "CardDAV"

  Scenario: Read calendar resource returns CalDAV message
    When I read the MCP resource "fastmail://calendar/today"
    Then the resource read should succeed
    And the resource content should contain "CalDAV"

  Scenario: Read unknown resource returns error
    When I read the MCP resource "fastmail://nonexistent"
    Then the resource read should fail
    And the resource error should contain "not found"
