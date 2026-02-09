Feature: Account quota
  As a Fastmail user
  I want to check my storage quota
  So that I can see how much space I have used

  Background:
    Given a connected Fastmail client

  Scenario: Get quota info
    Given the account has storage quota of 1073741824 bytes used and 5368709120 bytes limit
    When I get the account quota
    Then the quota should show 1073741824 bytes used
    And the quota should show 5368709120 bytes limit
    And the quota used percentage should be approximately 20.0
