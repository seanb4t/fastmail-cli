Feature: Calendar event CRUD operations
  As a Fastmail user
  I want to get, create, update, and delete calendar events
  So that I can manage my schedule

  Background:
    Given the following calendars exist:
      | id                       | name     | description       |
      | /user/calendars/cal1     | Personal | Personal calendar |
    And the following events exist in calendar "/user/calendars/cal1":
      | uid   | summary        | start                 | end                   | location  | description  |
      | evt-1 | Team Meeting   | 2025-01-15T10:00:00Z  | 2025-01-15T11:00:00Z  | Room 42   | Weekly sync  |
      | evt-2 | Lunch with Bob | 2025-01-15T12:00:00Z  | 2025-01-15T13:00:00Z  | Cafe      | Catch up     |

  Scenario: Get a single event by ID
    When I get event "/user/calendars/cal1/evt-1.ics"
    Then the operation should succeed
    And the event should have summary "Team Meeting"
    And the event should have location "Room 42"
    And the event should have description "Weekly sync"

  Scenario: Create a new event
    When I create an event in calendar "/user/calendars/cal1" with:
      | uid   | summary      | start                 | end                   | location    | description       |
      | evt-3 | Code Review  | 2025-01-16T14:00:00Z  | 2025-01-16T15:00:00Z  | Conference  | Review sprint PR  |
    Then the operation should succeed

  Scenario: Update an existing event
    When I update event "/user/calendars/cal1/evt-1.ics" in calendar "/user/calendars/cal1" with summary "Updated Meeting"
    Then the operation should succeed

  Scenario: Delete an event
    When I delete event "/user/calendars/cal1/evt-1.ics"
    Then the operation should succeed
