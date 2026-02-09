Feature: Calendar operations
  As a Fastmail user
  I want to manage my calendars and events
  So that I can organize my schedule

  Scenario: List calendars
    Given the following calendars exist:
      | id              | name      | description         |
      | /user/calendars/cal1 | Personal  | Personal calendar   |
      | /user/calendars/cal2 | Work      | Work calendar       |
    When I list calendars
    Then I should receive 2 calendars
    And calendar 1 should have name "Personal"
    And calendar 2 should have name "Work"

  Scenario: List events in a date range
    Given the following calendars exist:
      | id              | name     | description       |
      | /user/calendars/cal1 | Personal | Personal calendar |
    And the following events exist in calendar "/user/calendars/cal1":
      | uid   | summary        | start                 | end                   | location  | description  |
      | evt-1 | Team Meeting   | 2025-01-15T10:00:00Z  | 2025-01-15T11:00:00Z  | Room 42   | Weekly sync  |
      | evt-2 | Lunch with Bob | 2025-01-15T12:00:00Z  | 2025-01-15T13:00:00Z  | Cafe      | Catch up     |
    When I list events in calendar "/user/calendars/cal1" from "2025-01-15T00:00:00Z" to "2025-01-16T00:00:00Z"
    Then I should receive 2 events
    And event 1 should have summary "Team Meeting"
    And event 1 should have location "Room 42"
    And event 2 should have summary "Lunch with Bob"

  Scenario: List events with no results
    Given the following calendars exist:
      | id              | name     | description       |
      | /user/calendars/cal1 | Personal | Personal calendar |
    When I list events in calendar "/user/calendars/cal1" from "2025-01-15T00:00:00Z" to "2025-01-16T00:00:00Z"
    Then I should receive 0 events

  Scenario: Calendar has correct properties
    Given the following calendars exist:
      | id              | name      | description              |
      | /user/calendars/cal1 | Birthdays | Friends and family dates |
    When I list calendars
    Then I should receive 1 calendar
    And calendar 1 should have name "Birthdays"
    And calendar 1 should have description "Friends and family dates"
