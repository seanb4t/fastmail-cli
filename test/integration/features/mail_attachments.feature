Feature: Mail attachment operations
  As a Fastmail user
  I want to manage email attachments
  So that I can view, download, and upload file attachments

  Background:
    Given a connected Fastmail client

  Scenario: List attachments on an email with attachments
    Given an email "e-att-1" with the following attachments:
      | blobId     | name          | type            | size  | disposition |
      | blob-pdf   | report.pdf    | application/pdf | 51200 | attachment  |
      | blob-img   | photo.jpg     | image/jpeg      | 30720 | inline      |
    When I list attachments for email "e-att-1"
    Then I should receive 2 attachments
    And attachment 1 should have name "report.pdf"
    And attachment 1 should have type "application/pdf"
    And attachment 1 should have size 51200
    And attachment 1 should have disposition "attachment"
    And attachment 2 should have name "photo.jpg"
    And attachment 2 should have type "image/jpeg"

  Scenario: List attachments on an email with no attachments
    Given an email "e-att-2" with no attachments
    When I list attachments for email "e-att-2"
    Then I should receive 0 attachments

  Scenario: Download an attachment blob
    Given a downloadable blob "blob-dl-1" with content "Hello, this is file content."
    When I download blob "blob-dl-1" with name "test.txt"
    Then the download should succeed
    And the downloaded content should be "Hello, this is file content."

  Scenario: Upload a blob
    Given upload support is available
    When I upload a blob with content "Upload test data" and type "text/plain"
    Then the upload should succeed
    And the uploaded blob ID should not be empty
    And the uploaded blob size should be 16
