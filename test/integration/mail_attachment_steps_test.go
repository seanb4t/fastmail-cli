//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// attachmentState holds per-scenario state for attachment tests.
type attachmentState struct {
	ResultAttachments []fastmail.Attachment
	DownloadedContent string
	UploadedBlobID    string
	UploadedBlobSize  uint64
}

func getAttachmentState(w *World) *attachmentState {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	if s, ok := w.DomainData["attachments"].(*attachmentState); ok {
		return s
	}
	s := &attachmentState{}
	w.DomainData["attachments"] = s
	return s
}

func registerMailAttachmentSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^an email "([^"]*)" with the following attachments:$`, anEmailWithAttachments)
	sc.Step(`^an email "([^"]*)" with no attachments$`, anEmailWithNoAttachments)
	sc.Step(`^a downloadable blob "([^"]*)" with content "([^"]*)"$`, aDownloadableBlob)
	sc.Step(`^upload support is available$`, uploadSupportIsAvailable)

	// When steps
	sc.Step(`^I list attachments for email "([^"]*)"$`, iListAttachmentsForEmail)
	sc.Step(`^I download blob "([^"]*)" with name "([^"]*)"$`, iDownloadBlob)
	sc.Step(`^I upload a blob with content "([^"]*)" and type "([^"]*)"$`, iUploadBlob)

	// Then steps
	sc.Step(`^I should receive (\d+) attachments?$`, iShouldReceiveNAttachments)
	sc.Step(`^attachment (\d+) should have name "([^"]*)"$`, attachmentShouldHaveName)
	sc.Step(`^attachment (\d+) should have type "([^"]*)"$`, attachmentShouldHaveType)
	sc.Step(`^attachment (\d+) should have size (\d+)$`, attachmentShouldHaveSize)
	sc.Step(`^attachment (\d+) should have disposition "([^"]*)"$`, attachmentShouldHaveDisposition)
	sc.Step(`^the download should succeed$`, theDownloadShouldSucceed)
	sc.Step(`^the downloaded content should be "([^"]*)"$`, theDownloadedContentShouldBe)
	sc.Step(`^the upload should succeed$`, theUploadShouldSucceed)
	sc.Step(`^the uploaded blob ID should not be empty$`, theUploadedBlobIDShouldNotBeEmpty)
	sc.Step(`^the uploaded blob size should be (\d+)$`, theUploadedBlobSizeShouldBe)
}

// Given steps

func anEmailWithAttachments(ctx context.Context, emailID string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	attachments := make([]MockAttachment, 0, len(table.Rows)-1)
	for _, row := range table.Rows[1:] {
		size, _ := strconv.ParseUint(cellValue(row, headers, "size"), 10, 64)
		attachments = append(attachments, MockAttachment{
			BlobID:      cellValue(row, headers, "blobId"),
			Name:        cellValue(row, headers, "name"),
			Type:        cellValue(row, headers, "type"),
			Size:        size,
			Disposition: cellValue(row, headers, "disposition"),
		})
	}

	w.Emails = append(w.Emails, MockEmail{
		ID:          emailID,
		Subject:     "Email with attachments",
		Preview:     "Has attachments",
		Attachments: attachments,
	})
	return ctx, nil
}

func anEmailWithNoAttachments(ctx context.Context, emailID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	w.Emails = append(w.Emails, MockEmail{
		ID:      emailID,
		Subject: "Email without attachments",
		Preview: "No attachments",
	})
	return ctx, nil
}

func aDownloadableBlob(ctx context.Context, blobID, content string) (context.Context, error) {
	w := WorldFromContext(ctx)
	setupBlobServer(w, map[string]string{blobID: content})
	return ctx, nil
}

func uploadSupportIsAvailable(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	setupBlobServer(w, nil)
	return ctx, nil
}

// setupBlobServer creates a mock blob server for download/upload and reconfigures
// the session server to return URLs pointing to it.
func setupBlobServer(w *World, blobs map[string]string) {
	if w.BlobServer != nil {
		return // already set up
	}

	w.BlobServer = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/download/"):
			handleBlobDownload(rw, r, blobs)
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/upload/"):
			handleBlobUpload(rw, r)
		default:
			http.NotFound(rw, r)
		}
	}))

	// Reconfigure the session server to point download/upload URLs to the blob server
	downloadURL := w.BlobServer.URL + "/download/{accountId}/{blobId}/{name}"
	uploadURL := w.BlobServer.URL + "/upload/{accountId}/"

	w.SessionServer.Close()
	w.SessionServer = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(SessionResponseWithBlob(w.APIServer.URL, downloadURL, uploadURL)))
	}))
}

func handleBlobDownload(rw http.ResponseWriter, r *http.Request, blobs map[string]string) {
	// Path format: /download/{accountId}/{blobId}/{name}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/download/"), "/")
	if len(parts) < 2 { //nolint:mnd // min path segments: accountId + blobId
		http.Error(rw, "invalid download path", http.StatusBadRequest)
		return
	}
	blobID := parts[1]

	content, ok := blobs[blobID]
	if !ok {
		http.NotFound(rw, r)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	_, _ = rw.Write([]byte(content))
}

func handleBlobUpload(rw http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "failed to read body", http.StatusBadRequest)
		return
	}

	resp := map[string]any{
		"accountId": "acc1",
		"blobId":    "blob-uploaded-1",
		"type":      r.Header.Get("Content-Type"),
		"size":      len(body),
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(resp)
}

// When steps

func iListAttachmentsForEmail(ctx context.Context, emailID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	state := getAttachmentState(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	attachments, err := client.Mail().Attachments(ctx, emailID)
	w.ResultError = err
	state.ResultAttachments = attachments
	return ctx, nil
}

func iDownloadBlob(ctx context.Context, blobID, name string) (context.Context, error) {
	w := WorldFromContext(ctx)
	state := getAttachmentState(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	reader, err := client.Mail().DownloadAttachment(ctx, blobID, name)
	w.ResultError = err

	if err == nil && reader != nil {
		data, readErr := io.ReadAll(reader)
		_ = reader.Close()
		if readErr != nil {
			w.ResultError = readErr
		} else {
			state.DownloadedContent = string(data)
		}
	}

	return ctx, nil
}

func iUploadBlob(ctx context.Context, content, contentType string) (context.Context, error) {
	w := WorldFromContext(ctx)
	state := getAttachmentState(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	blobID, size, err := client.Mail().UploadBlob(ctx, bytes.NewReader([]byte(content)), contentType)
	w.ResultError = err
	state.UploadedBlobID = blobID
	state.UploadedBlobSize = size

	return ctx, nil
}

// Then steps

func iShouldReceiveNAttachments(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("unexpected error: %w", w.ResultError)
	}
	state := getAttachmentState(w)
	if len(state.ResultAttachments) != count {
		return fmt.Errorf("expected %d attachments, got %d", count, len(state.ResultAttachments))
	}
	return nil
}

func attachmentShouldHaveName(ctx context.Context, index int, name string) error {
	state := getAttachmentState(WorldFromContext(ctx))
	i := index - 1
	if i < 0 || i >= len(state.ResultAttachments) {
		return fmt.Errorf("attachment index %d out of range (have %d)", index, len(state.ResultAttachments))
	}
	if state.ResultAttachments[i].Name != name {
		return fmt.Errorf("expected name %q, got %q", name, state.ResultAttachments[i].Name)
	}
	return nil
}

func attachmentShouldHaveType(ctx context.Context, index int, typ string) error {
	state := getAttachmentState(WorldFromContext(ctx))
	i := index - 1
	if i < 0 || i >= len(state.ResultAttachments) {
		return fmt.Errorf("attachment index %d out of range (have %d)", index, len(state.ResultAttachments))
	}
	if state.ResultAttachments[i].Type != typ {
		return fmt.Errorf("expected type %q, got %q", typ, state.ResultAttachments[i].Type)
	}
	return nil
}

func attachmentShouldHaveSize(ctx context.Context, index, size int) error {
	state := getAttachmentState(WorldFromContext(ctx))
	i := index - 1
	if i < 0 || i >= len(state.ResultAttachments) {
		return fmt.Errorf("attachment index %d out of range (have %d)", index, len(state.ResultAttachments))
	}
	if state.ResultAttachments[i].Size != uint64(size) { //nolint:gosec // test data, no overflow risk
		return fmt.Errorf("expected size %d, got %d", size, state.ResultAttachments[i].Size)
	}
	return nil
}

func attachmentShouldHaveDisposition(ctx context.Context, index int, disposition string) error {
	state := getAttachmentState(WorldFromContext(ctx))
	i := index - 1
	if i < 0 || i >= len(state.ResultAttachments) {
		return fmt.Errorf("attachment index %d out of range (have %d)", index, len(state.ResultAttachments))
	}
	if state.ResultAttachments[i].Disposition != disposition {
		return fmt.Errorf("expected disposition %q, got %q", disposition, state.ResultAttachments[i].Disposition)
	}
	return nil
}

func theDownloadShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("download failed: %w", w.ResultError)
	}
	return nil
}

func theDownloadedContentShouldBe(ctx context.Context, expected string) error {
	state := getAttachmentState(WorldFromContext(ctx))
	if state.DownloadedContent != expected {
		return fmt.Errorf("expected content %q, got %q", expected, state.DownloadedContent)
	}
	return nil
}

func theUploadShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("upload failed: %w", w.ResultError)
	}
	return nil
}

func theUploadedBlobIDShouldNotBeEmpty(ctx context.Context) error {
	state := getAttachmentState(WorldFromContext(ctx))
	if state.UploadedBlobID == "" {
		return fmt.Errorf("uploaded blob ID is empty")
	}
	return nil
}

func theUploadedBlobSizeShouldBe(ctx context.Context, size int) error {
	state := getAttachmentState(WorldFromContext(ctx))
	if state.UploadedBlobSize != uint64(size) { //nolint:gosec // test data, no overflow risk
		return fmt.Errorf("expected uploaded size %d, got %d", size, state.UploadedBlobSize)
	}
	return nil
}
