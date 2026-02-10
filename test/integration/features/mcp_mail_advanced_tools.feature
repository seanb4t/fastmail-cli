Feature: MCP mail advanced tools
  As an AI agent using MCP
  I want to use advanced mail tools for threads, attachments, and blob operations
  So that I can help users manage conversations, downloads, imports, and uploads

  Background:
    Given a connected MCP tools config

  # --- mail_thread ---

  Scenario: MCP mail_thread tool returns thread emails
    Given a thread "thread-adv-1" with the following emails:
      | id      | subject          | from_email        | receivedAt           |
      | e-adv-1 | Thread Start     | alice@example.com | 2024-06-01T09:00:00Z |
      | e-adv-2 | Re: Thread Start | bob@example.com   | 2024-06-01T10:00:00Z |
    When I call the "mail_thread" tool with:
      | key       | value        |
      | thread_id | thread-adv-1 |
    Then the tool call should succeed
    And the tool result should contain 2 items

  Scenario: MCP mail_thread tool fails without thread_id
    When I call the "mail_thread" tool
    Then the tool call should fail
    And the tool error should contain "thread_id is required"

  # --- mail_attachments ---

  Scenario: MCP mail_attachments tool lists email attachments
    Given an email "e-mcp-att" with the following attachments:
      | blobId    | name        | type            | size  | disposition |
      | blob-doc  | report.pdf  | application/pdf | 51200 | attachment  |
      | blob-img  | logo.png    | image/png       | 10240 | inline      |
    When I call the "mail_attachments" tool with:
      | key | value     |
      | id  | e-mcp-att |
    Then the tool call should succeed
    And the tool result should contain 2 items
    And the tool result item 1 field "name" should equal "report.pdf"
    And the tool result item 1 field "type" should equal "application/pdf"
    And the tool result item 2 field "name" should equal "logo.png"

  Scenario: MCP mail_attachments tool returns empty for email without attachments
    Given an email "e-mcp-noatt" with no attachments
    When I call the "mail_attachments" tool with:
      | key | value       |
      | id  | e-mcp-noatt |
    Then the tool call should succeed
    And the tool result should contain 0 items

  Scenario: MCP mail_attachments tool fails without id
    When I call the "mail_attachments" tool
    Then the tool call should fail
    And the tool error should contain "id is required"

  # --- mail_download ---

  Scenario: MCP mail_download tool downloads an attachment
    Given an email "e-mcp-dl" with the following attachments:
      | blobId     | name      | type            | size | disposition |
      | blob-dl-mcp| readme.md | text/markdown   | 42   | attachment  |
    And a downloadable blob "blob-dl-mcp" with content "# Hello World"
    When I call the "mail_download" advanced mail tool with:
      | key     | value       |
      | id      | e-mcp-dl    |
      | blob_id | blob-dl-mcp |
    Then the tool call should succeed
    And the tool result field "blob_id" should equal "blob-dl-mcp"
    And the tool result field "name" should equal "readme.md"
    And the tool result field "encoding" should equal "base64"

  Scenario: MCP mail_download tool fails without id
    When I call the "mail_download" advanced mail tool with:
      | key     | value    |
      | blob_id | some-blob|
    Then the tool call should fail
    And the tool error should contain "id is required"

  Scenario: MCP mail_download tool fails without blob_id
    When I call the "mail_download" advanced mail tool with:
      | key | value |
      | id  | e-123 |
    Then the tool call should fail
    And the tool error should contain "blob_id is required"

  # --- mail_import ---

  Scenario: MCP mail_import tool imports an email
    Given a blob upload server is configured
    When I call the "mail_import" advanced mail tool with:
      | key     | value                                    |
      | content | RnJvbTogdGVzdEBleGFtcGxlLmNvbQ0KDQpIZWxsbw== |
    Then the tool call should succeed
    And the tool result field "status" should equal "imported"
    And the tool result field "id" should equal "imported-email-1"

  Scenario: MCP mail_import tool with options
    Given a blob upload server is configured
    When I call the "mail_import" advanced mail tool with:
      | key     | value                                    |
      | content | RnJvbTogdGVzdEBleGFtcGxlLmNvbQ0KDQpIZWxsbw== |
      | folder  | Archive                                  |
      | seen    | true                                     |
      | flagged | true                                     |
    Then the tool call should succeed
    And the tool result field "status" should equal "imported"

  Scenario: MCP mail_import tool fails without content
    When I call the "mail_import" advanced mail tool
    Then the tool call should fail
    And the tool error should contain "content is required"

  # --- mail_upload ---

  Scenario: MCP mail_upload tool uploads a blob
    Given a blob upload server is configured
    When I call the "mail_upload" advanced mail tool with:
      | key          | value           |
      | content      | SGVsbG8gV29ybGQ= |
      | content_type | text/plain      |
      | filename     | hello.txt       |
    Then the tool call should succeed
    And the tool result field "filename" should equal "hello.txt"

  Scenario: MCP mail_upload tool fails without content
    When I call the "mail_upload" advanced mail tool with:
      | key          | value      |
      | content_type | text/plain |
      | filename     | hello.txt  |
    Then the tool call should fail
    And the tool error should contain "content is required"

  Scenario: MCP mail_upload tool fails without content_type
    When I call the "mail_upload" advanced mail tool with:
      | key      | value           |
      | content  | SGVsbG8gV29ybGQ= |
      | filename | hello.txt       |
    Then the tool call should fail
    And the tool error should contain "content_type is required"

  Scenario: MCP mail_upload tool fails without filename
    When I call the "mail_upload" advanced mail tool with:
      | key          | value           |
      | content      | SGVsbG8gV29ybGQ= |
      | content_type | text/plain      |
    Then the tool call should fail
    And the tool error should contain "filename is required"
