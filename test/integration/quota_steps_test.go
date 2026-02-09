//go:build integration

package integration

import (
	"context"
	"fmt"
	"math"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// quotaResult stores the result from a quota Get call.
type quotaResult struct {
	Info *fastmail.QuotaInfo
	Err  error
}

func registerQuotaSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the account has storage quota of (\d+) bytes used and (\d+) bytes limit$`, theAccountHasStorageQuota)
	sc.Step(`^I get the account quota$`, iGetTheAccountQuota)
	sc.Step(`^the quota should show (\d+) bytes used$`, theQuotaShouldShowBytesUsed)
	sc.Step(`^the quota should show (\d+) bytes limit$`, theQuotaShouldShowBytesLimit)
	sc.Step(`^the quota used percentage should be approximately ([0-9.]+)$`, theQuotaUsedPercentageShouldBe)
}

func theAccountHasStorageQuota(ctx context.Context, used, limit uint64) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	w.DomainData["quota"] = map[string]any{
		"used":  used,
		"limit": limit,
	}
	return ctx, nil
}

func iGetTheAccountQuota(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	info, err := client.Quota().Get(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	w.DomainData["quotaResult"] = &quotaResult{Info: info, Err: err}
	return ctx, nil
}

func theQuotaShouldShowBytesUsed(ctx context.Context, expected uint64) error {
	w := WorldFromContext(ctx)
	qr, ok := w.DomainData["quotaResult"].(*quotaResult)
	if !ok {
		return fmt.Errorf("no quota result found")
	}
	if qr.Err != nil {
		return fmt.Errorf("unexpected error: %w", qr.Err)
	}
	if qr.Info == nil {
		return fmt.Errorf("quota info is nil")
	}
	if qr.Info.Used != expected {
		return fmt.Errorf("expected used %d, got %d", expected, qr.Info.Used)
	}
	return nil
}

func theQuotaShouldShowBytesLimit(ctx context.Context, expected uint64) error {
	w := WorldFromContext(ctx)
	qr, ok := w.DomainData["quotaResult"].(*quotaResult)
	if !ok {
		return fmt.Errorf("no quota result found")
	}
	if qr.Err != nil {
		return fmt.Errorf("unexpected error: %w", qr.Err)
	}
	if qr.Info == nil {
		return fmt.Errorf("quota info is nil")
	}
	if qr.Info.Limit != expected {
		return fmt.Errorf("expected limit %d, got %d", expected, qr.Info.Limit)
	}
	return nil
}

func theQuotaUsedPercentageShouldBe(ctx context.Context, expected float64) error {
	w := WorldFromContext(ctx)
	qr, ok := w.DomainData["quotaResult"].(*quotaResult)
	if !ok {
		return fmt.Errorf("no quota result found")
	}
	if qr.Err != nil {
		return fmt.Errorf("unexpected error: %w", qr.Err)
	}
	if qr.Info == nil {
		return fmt.Errorf("quota info is nil")
	}
	if math.Abs(qr.Info.UsedPercent-expected) > 0.1 {
		return fmt.Errorf("expected percentage ~%.1f, got %.1f", expected, qr.Info.UsedPercent)
	}
	return nil
}
