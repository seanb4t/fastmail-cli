package fastmail

import (
	"context"
	"fmt"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// QuotaInfo represents storage quota information.
type QuotaInfo struct {
	Used        uint64  `json:"used"`
	Limit       uint64  `json:"limit"`
	UsedPercent float64 `json:"used_percent"`
}

// FormatSize returns a human-readable size string (e.g., "2.1 GB").
func FormatSize(bytes uint64) string {
	const (
		_          = iota
		kb float64 = 1024
		mb         = kb * 1024
		gb         = mb * 1024
		tb         = gb * 1024
	)

	b := float64(bytes)

	switch {
	case b >= tb:
		return fmt.Sprintf("%.1f TB", b/tb)
	case b >= gb:
		return fmt.Sprintf("%.1f GB", b/gb)
	case b >= mb:
		return fmt.Sprintf("%.1f MB", b/mb)
	case b >= kb:
		return fmt.Sprintf("%.1f KB", b/kb)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// QuotaService provides quota/storage operations.
type QuotaService struct {
	client *Client
}

// Get fetches the storage quota for the account.
// It returns the "octets" (bytes) quota.
func (s *QuotaService) Get(ctx context.Context) (*QuotaInfo, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	builder := jmap.NewQuotaGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapQuota)
	callID := req.Invoke("Quota/get", builder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("quota get failed: %s", result.Error())
	}

	var quotaResp jmap.QuotaGetResponse
	if err := result.Decode(&quotaResp); err != nil {
		return nil, oops.Wrapf(err, "decoding quota response")
	}

	// Find the octets (bytes) quota
	for _, q := range quotaResp.List {
		if q.ResourceType == "octets" {
			var usedPercent float64
			if q.HardLimit > 0 {
				usedPercent = float64(q.Used) / float64(q.HardLimit) * 100
			}

			return &QuotaInfo{
				Used:        q.Used,
				Limit:       q.HardLimit,
				UsedPercent: usedPercent,
			}, nil
		}
	}

	return nil, oops.Errorf("no storage quota found")
}
