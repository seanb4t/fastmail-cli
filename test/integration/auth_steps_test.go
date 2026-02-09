//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/internal/auth"
)

// authResult holds captured state from auth operations.
type authResult struct {
	store       *auth.Store
	configPath  string
	tmpDir      string
	token       string
	hasToken    bool
	err         error
	oldEnvToken string
}

func getAuthResult(w *World) *authResult {
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	r, ok := w.DomainData["auth"].(*authResult)
	if !ok {
		r = &authResult{}
		w.DomainData["auth"] = r
	}
	return r
}

func registerAuthSteps(sc *godog.ScenarioContext) {
	// Background / Given steps
	sc.Step(`^a fresh auth store$`, aFreshAuthStore)

	// When steps
	sc.Step(`^I store the token "([^"]*)"$`, iStoreTheToken)
	sc.Step(`^I retrieve the stored token$`, iRetrieveTheStoredToken)
	sc.Step(`^I check if a token exists$`, iCheckIfATokenExists)
	sc.Step(`^I delete the stored token$`, iDeleteTheStoredToken)
	sc.Step(`^I create a new store with the same config path$`, iCreateANewStoreWithTheSameConfigPath)

	// Then steps
	sc.Step(`^the token operation should succeed$`, theTokenOperationShouldSucceed)
	sc.Step(`^the token should equal "([^"]*)"$`, theTokenShouldEqual)
	sc.Step(`^the token should exist$`, theTokenShouldExist)
	sc.Step(`^the token should not exist$`, theTokenShouldNotExist)
	sc.Step(`^the token retrieval should fail$`, theTokenRetrievalShouldFail)
	sc.Step(`^the token error should contain "([^"]*)"$`, theTokenErrorShouldContain)

	// Cleanup after each scenario
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w := WorldFromContext(ctx)
		r := getAuthResult(w)
		if r.tmpDir != "" {
			_ = os.RemoveAll(r.tmpDir)
		}
		if r.oldEnvToken != "" {
			_ = os.Setenv(auth.EnvToken, r.oldEnvToken)
		}
		return ctx, nil
	})
}

// Background / Given steps

func aFreshAuthStore(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)

	// Clear FASTMAIL_TOKEN env var to avoid interference; restore in After hook
	r.oldEnvToken = os.Getenv(auth.EnvToken)
	if r.oldEnvToken != "" {
		if err := os.Unsetenv(auth.EnvToken); err != nil {
			return ctx, fmt.Errorf("unsetting %s: %w", auth.EnvToken, err)
		}
	}

	tmpDir, err := os.MkdirTemp("", "auth-test-*")
	if err != nil {
		return ctx, fmt.Errorf("creating temp dir: %w", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	store := auth.NewStore(configPath)
	store.DisableKeychain()
	store.SetWarningWriter(io.Discard)

	r.store = store
	r.configPath = configPath
	r.tmpDir = tmpDir

	return ctx, nil
}

// When steps

func iStoreTheToken(ctx context.Context, token string) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	r.err = r.store.SetToken(token)
	return ctx, nil
}

func iRetrieveTheStoredToken(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	r.token, r.err = r.store.GetToken()
	return ctx, nil
}

func iCheckIfATokenExists(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	r.hasToken = r.store.HasToken()
	return ctx, nil
}

func iDeleteTheStoredToken(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	r.err = r.store.DeleteToken()
	return ctx, nil
}

func iCreateANewStoreWithTheSameConfigPath(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)

	store := auth.NewStore(r.configPath)
	store.DisableKeychain()
	store.SetWarningWriter(io.Discard)
	r.store = store

	return ctx, nil
}

// Then steps

func theTokenOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	if r.err != nil {
		return fmt.Errorf("expected no error, got: %w", r.err)
	}
	return nil
}

func theTokenShouldEqual(ctx context.Context, expected string) error {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	if r.err != nil {
		return fmt.Errorf("unexpected error retrieving token: %w", r.err)
	}
	if r.token != expected {
		return fmt.Errorf("expected token %q, got %q", expected, r.token)
	}
	return nil
}

func theTokenShouldExist(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	if !r.hasToken {
		return fmt.Errorf("expected token to exist, but it does not")
	}
	return nil
}

func theTokenShouldNotExist(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	if r.hasToken {
		return fmt.Errorf("expected no token, but one exists")
	}
	return nil
}

func theTokenRetrievalShouldFail(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	if r.err == nil {
		return fmt.Errorf("expected error, but got none (token=%q)", r.token)
	}
	return nil
}

func theTokenErrorShouldContain(ctx context.Context, substring string) error {
	w := WorldFromContext(ctx)
	r := getAuthResult(w)
	if r.err == nil {
		return fmt.Errorf("expected error containing %q, but got no error", substring)
	}
	if !strings.Contains(r.err.Error(), substring) {
		return fmt.Errorf("expected error containing %q, got: %s", substring, r.err.Error())
	}
	return nil
}
