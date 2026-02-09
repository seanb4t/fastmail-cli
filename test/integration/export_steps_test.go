//go:build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

const (
	exportOutputKey = "export_output"
	exportErrorKey  = "export_error"
	exportDirKey    = "export_dir"
)

func registerExportSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I export emails as "([^"]*)"$`, iExportEmailsAs)
	sc.Step(`^the export should succeed$`, theExportShouldSucceed)
	sc.Step(`^the export output should contain "([^"]*)"$`, theExportOutputShouldContain)
	sc.Step(`^the export output should have (\d+) lines?$`, theExportOutputShouldHaveNLines)
	sc.Step(`^the maildir should have (\d+) files?$`, theMaildirShouldHaveNFiles)

	// Clean up maildir temp directories after each scenario.
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w := WorldFromContext(ctx)
		if w.DomainData == nil {
			return ctx, nil
		}
		if dir, ok := w.DomainData[exportDirKey].(string); ok && dir != "" {
			_ = os.RemoveAll(dir)
		}
		return ctx, nil
	})
}

// ensureExportDomainData lazily initializes DomainData on the World.
func ensureExportDomainData(w *World) {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
}

// mockEmailsToFastmail converts the World's MockEmail slice to fastmail.Email slice.
func mockEmailsToFastmail(mocks []MockEmail) []fastmail.Email {
	emails := make([]fastmail.Email, len(mocks))
	for i, m := range mocks {
		receivedAt, _ := time.Parse(time.RFC3339, m.ReceivedAt)
		if receivedAt.IsZero() {
			receivedAt = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		}

		var keywords []string
		for k, v := range m.Keywords {
			if v {
				keywords = append(keywords, k)
			}
		}

		emails[i] = fastmail.Email{
			ID:         m.ID,
			Subject:    m.Subject,
			Preview:    m.Preview,
			ReceivedAt: receivedAt,
			Keywords:   keywords,
		}
	}
	return emails
}

// When steps

func iExportEmailsAs(ctx context.Context, format string) (context.Context, error) {
	w := WorldFromContext(ctx)
	ensureExportDomainData(w)

	emails := mockEmailsToFastmail(w.Emails)

	switch format {
	case "jsonl":
		var buf bytes.Buffer
		err := fastmail.ExportJSONL(&buf, emails)
		w.DomainData[exportOutputKey] = buf.String()
		w.DomainData[exportErrorKey] = err

	case "mbox":
		var buf bytes.Buffer
		err := fastmail.ExportMbox(&buf, emails)
		w.DomainData[exportOutputKey] = buf.String()
		w.DomainData[exportErrorKey] = err

	case "maildir":
		dir, err := os.MkdirTemp("", "fastmail-test-maildir-*")
		if err != nil {
			return ctx, fmt.Errorf("creating temp dir: %w", err)
		}
		w.DomainData[exportDirKey] = dir
		exportErr := fastmail.ExportMaildir(dir, emails)
		w.DomainData[exportErrorKey] = exportErr

	default:
		return ctx, fmt.Errorf("unknown export format: %s", format)
	}

	return ctx, nil
}

// Then steps

func theExportShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	ensureExportDomainData(w)

	if err, ok := w.DomainData[exportErrorKey].(error); ok && err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	return nil
}

func theExportOutputShouldContain(ctx context.Context, expected string) error {
	w := WorldFromContext(ctx)
	ensureExportDomainData(w)

	output, _ := w.DomainData[exportOutputKey].(string)
	if !strings.Contains(output, expected) {
		return fmt.Errorf("export output does not contain %q; got:\n%s", expected, output)
	}
	return nil
}

func theExportOutputShouldHaveNLines(ctx context.Context, expected int) error {
	w := WorldFromContext(ctx)
	ensureExportDomainData(w)

	output, _ := w.DomainData[exportOutputKey].(string)

	// Count non-empty lines.
	var count int
	if output != "" {
		for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
	}

	if count != expected {
		return fmt.Errorf("expected %d non-empty lines, got %d; output:\n%s", expected, count, output)
	}
	return nil
}

func theMaildirShouldHaveNFiles(ctx context.Context, expected int) error {
	w := WorldFromContext(ctx)
	ensureExportDomainData(w)

	dir, _ := w.DomainData[exportDirKey].(string)
	if dir == "" {
		return fmt.Errorf("no maildir directory set (was maildir export run?)")
	}

	// Verify Maildir subdirectory structure exists.
	for _, subdir := range []string{"cur", "new", "tmp"} {
		info, err := os.Stat(filepath.Join(dir, subdir))
		if err != nil {
			return fmt.Errorf("maildir subdir %q missing: %w", subdir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("maildir %q is not a directory", subdir)
		}
	}

	// Count files in cur/ (where ExportMaildir writes them).
	entries, err := os.ReadDir(filepath.Join(dir, "cur"))
	if err != nil {
		return fmt.Errorf("reading cur/ directory: %w", err)
	}

	var fileCount int
	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
		}
	}

	if fileCount != expected {
		return fmt.Errorf("expected %d files in maildir cur/, got %d", expected, fileCount)
	}
	return nil
}
