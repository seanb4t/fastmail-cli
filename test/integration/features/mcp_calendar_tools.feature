Feature: MCP calendar tools
  As an AI agent using MCP
  I want to manage calendars and events via MCP tools
  So that I can help users with their calendar

  Background:
    Given a connected MCP tools config

  Scenario: MCP calendar_list tool
    Given the following MCP calendars exist:
      | id    | name     | color   | description     | is_default | read_only |
      | cal-1 | Personal | #FF0000 | My calendar     | true       | false     |
      | cal-2 | Work     | #0000FF | Work calendar   | false      | false     |
    When I call the "calendar_list" calendar tool
    Then the tool call should succeed
    And the tool result should contain 2 items
    And the tool result item 1 field "name" should equal "Personal"
    And the tool result item 2 field "name" should equal "Work"

  Scenario: MCP calendar_events tool
    Given the following MCP calendar events exist:
      | id   | calendar_id | summary       | start                    | end                      |
      | ev-1 | cal-1       | Team Standup  | 2024-06-01T09:00:00Z     | 2024-06-01T09:30:00Z     |
      | ev-2 | cal-1       | Lunch         | 2024-06-01T12:00:00Z     | 2024-06-01T13:00:00Z     |
    When I call the "calendar_events" calendar tool with:
      | key         | value                    |
      | start       | 2024-06-01T00:00:00Z     |
      | end         | 2024-06-02T00:00:00Z     |
    Then the tool call should succeed
    And the tool result should contain 2 items

  Scenario: MCP calendar_events with calendar_id filter
    Given the following MCP calendar events exist:
      | id   | calendar_id | summary       | start                    | end                      |
      | ev-1 | cal-1       | Personal Mtg  | 2024-06-01T09:00:00Z     | 2024-06-01T09:30:00Z     |
      | ev-2 | cal-2       | Work Mtg      | 2024-06-01T10:00:00Z     | 2024-06-01T10:30:00Z     |
    When I call the "calendar_events" calendar tool with:
      | key         | value                    |
      | calendar_id | cal-1                    |
      | start       | 2024-06-01T00:00:00Z     |
      | end         | 2024-06-02T00:00:00Z     |
    Then the tool call should succeed
    And the tool result should contain 1 items

  Scenario: MCP calendar_events fails without start
    When I call the "calendar_events" calendar tool with:
      | key | value                    |
      | end | 2024-06-02T00:00:00Z     |
    Then the tool call should fail
    And the tool error should contain "start is required"

  Scenario: MCP calendar_events fails without end
    When I call the "calendar_events" calendar tool with:
      | key   | value                    |
      | start | 2024-06-01T00:00:00Z     |
    Then the tool call should fail
    And the tool error should contain "end is required"

  Scenario: MCP calendar_create tool
    When I call the "calendar_create" calendar tool with:
      | key         | value                    |
      | calendar_id | cal-1                    |
      | summary     | New Meeting              |
      | start       | 2024-06-15T14:00:00Z     |
      | end         | 2024-06-15T15:00:00Z     |
    Then the tool call should succeed
    And the tool result field "status" should equal "created"
    And the tool result field "summary" should equal "New Meeting"
    And the tool result field "calendar_id" should equal "cal-1"

  Scenario: MCP calendar_create with optional fields
    When I call the "calendar_create" calendar tool with:
      | key         | value                    |
      | calendar_id | cal-1                    |
      | summary     | Team Offsite             |
      | description | Annual team offsite      |
      | location    | Conference Room B        |
      | start       | 2024-06-20T09:00:00Z     |
      | end         | 2024-06-20T17:00:00Z     |
    Then the tool call should succeed
    And the tool result field "status" should equal "created"
    And the tool result field "summary" should equal "Team Offsite"

  Scenario: MCP calendar_create fails without calendar_id
    When I call the "calendar_create" calendar tool with:
      | key     | value                    |
      | summary | No Calendar              |
      | start   | 2024-06-15T14:00:00Z     |
      | end     | 2024-06-15T15:00:00Z     |
    Then the tool call should fail
    And the tool error should contain "calendar_id is required"

  Scenario: MCP calendar_create fails without summary
    When I call the "calendar_create" calendar tool with:
      | key         | value                    |
      | calendar_id | cal-1                    |
      | start       | 2024-06-15T14:00:00Z     |
      | end         | 2024-06-15T15:00:00Z     |
    Then the tool call should fail
    And the tool error should contain "summary is required"

  Scenario: MCP calendar_get tool
    Given the following MCP calendar events exist:
      | id   | calendar_id | summary      | description  | location       | start                    | end                      |
      | ev-1 | cal-1       | Team Standup | Daily standup| Room 101       | 2024-06-01T09:00:00Z     | 2024-06-01T09:30:00Z     |
    When I call the "calendar_get" calendar tool with:
      | key | value |
      | id  | ev-1  |
    Then the tool call should succeed
    And the tool result field "id" should equal "ev-1"
    And the tool result field "summary" should equal "Team Standup"
    And the tool result field "description" should equal "Daily standup"
    And the tool result field "location" should equal "Room 101"

  Scenario: MCP calendar_get fails without id
    When I call the "calendar_get" calendar tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP calendar_get fails for nonexistent event
    When I call the "calendar_get" calendar tool with:
      | key | value       |
      | id  | nonexistent |
    Then the tool call should fail
    And the tool error should contain "event not found"

  Scenario: MCP calendar_update tool
    Given the following MCP calendar events exist:
      | id   | calendar_id | summary       | start                    | end                      |
      | ev-1 | cal-1       | Old Title     | 2024-06-01T09:00:00Z     | 2024-06-01T09:30:00Z     |
    When I call the "calendar_update" calendar tool with:
      | key     | value        |
      | id      | ev-1         |
      | summary | New Title    |
    Then the tool call should succeed
    And the tool result field "status" should equal "updated"
    And the tool result field "summary" should equal "New Title"

  Scenario: MCP calendar_update fails without id
    When I call the "calendar_update" calendar tool with:
      | key     | value     |
      | summary | No ID     |
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP calendar_delete tool
    Given the following MCP calendar events exist:
      | id   | calendar_id | summary       | start                    | end                      |
      | ev-1 | cal-1       | To Delete     | 2024-06-01T09:00:00Z     | 2024-06-01T09:30:00Z     |
    When I call the "calendar_delete" calendar tool with:
      | key | value |
      | id  | ev-1  |
    Then the tool call should succeed
    And the tool result field "status" should equal "deleted"
    And the tool result field "id" should equal "ev-1"

  Scenario: MCP calendar_delete fails without id
    When I call the "calendar_delete" calendar tool
    Then the tool call should fail
    And the tool error should contain "id is required"
