//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPResourceSteps(sc *godog.ScenarioContext) {
	// When steps
	sc.Step(`^I read the MCP resource "([^"]*)"$`, iReadTheMCPResource)

	// Then steps
	sc.Step(`^the resource read should succeed$`, theResourceReadShouldSucceed)
	sc.Step(`^the resource read should fail$`, theResourceReadShouldFail)
	sc.Step(`^the resource content should contain "([^"]*)"$`, theResourceContentShouldContain)
	sc.Step(`^the resource error should contain "([^"]*)"$`, theResourceErrorShouldContain)
}

// resourceResult holds captured results from MCP resource When steps.
type resourceResult struct {
	Content *mcp.ResourceContent
	Err     error
}

func getResourceResult(w *World) *resourceResult {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	result, ok := w.DomainData["resource_result"].(*resourceResult)
	if !ok {
		result = &resourceResult{}
		w.DomainData["resource_result"] = result
	}
	return result
}

// When steps

func iReadTheMCPResource(ctx context.Context, uri string) (context.Context, error) {
	w := WorldFromContext(ctx)
	result := getResourceResult(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	registry := mcp.NewResourceRegistry(client)

	content, err := registry.Read(ctx, uri)
	result.Content = content
	result.Err = err
	return ctx, nil
}

// Then steps

func theResourceReadShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getResourceResult(w)
	if result.Err != nil {
		return fmt.Errorf("expected resource read to succeed, got error: %w", result.Err)
	}
	if result.Content == nil {
		return fmt.Errorf("resource content is nil")
	}
	return nil
}

func theResourceReadShouldFail(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getResourceResult(w)
	if result.Err == nil {
		return fmt.Errorf("expected resource read to fail, but it succeeded")
	}
	return nil
}

func theResourceContentShouldContain(ctx context.Context, substring string) error {
	w := WorldFromContext(ctx)
	result := getResourceResult(w)
	if result.Content == nil {
		return fmt.Errorf("resource content is nil")
	}
	if !strings.Contains(result.Content.Text, substring) {
		return fmt.Errorf("expected resource content to contain %q, got:\n%s", substring, result.Content.Text)
	}
	return nil
}

func theResourceErrorShouldContain(ctx context.Context, substring string) error {
	w := WorldFromContext(ctx)
	result := getResourceResult(w)
	if result.Err == nil {
		return fmt.Errorf("expected an error containing %q, but no error occurred", substring)
	}
	if !strings.Contains(result.Err.Error(), substring) {
		return fmt.Errorf("expected error containing %q, got: %s", substring, result.Err.Error())
	}
	return nil
}
