package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// ThreadService provides thread/conversation operations.
type ThreadService struct {
	client *Client
}

// Get returns the emails in a thread, ordered by receivedAt.
func (s *ThreadService) Get(ctx context.Context, threadID string) ([]Email, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	// Step 1: Thread/get to get emailIds
	threadBuilder := jmap.NewThreadGet(accountID).IDs(threadID)
	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	threadCallID := req.Invoke("Thread/get", threadBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing Thread/get request")
	}

	threadResult, err := resp.GetResult(threadCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting thread result")
	}
	if threadResult.IsError() {
		return nil, oops.Errorf("thread get failed: %s", threadResult.Error())
	}

	var threadResp jmap.ThreadGetResponse
	if err := threadResult.Decode(&threadResp); err != nil {
		return nil, oops.Wrapf(err, "decoding thread response")
	}

	if len(threadResp.NotFound) > 0 || len(threadResp.List) == 0 {
		return nil, oops.Errorf("thread not found: %s", threadID)
	}

	emailIDs := threadResp.List[0].EmailIDs
	if len(emailIDs) == 0 {
		return []Email{}, nil
	}

	// Step 2: Email/get for the email summaries
	getBuilder := jmap.NewEmailGet(accountID).
		IDs(emailIDs...).
		Properties("id", "threadId", "subject", "from", "receivedAt", "preview", "size", "keywords", "mailboxIds")

	req2 := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	getCallID := req2.Invoke("Email/get", getBuilder.Build())

	resp2, err := s.client.jmap.Call(ctx, req2)
	if err != nil {
		return nil, oops.Wrapf(err, "executing Email/get request")
	}

	getResult, err := resp2.GetResult(getCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting email result")
	}
	if getResult.IsError() {
		return nil, oops.Errorf("email get failed: %s", getResult.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := getResult.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding email response")
	}

	return convertEmails(getResp.List), nil
}
