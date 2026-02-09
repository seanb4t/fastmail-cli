//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero exit status from godog")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w := &World{
			Mailboxes: DefaultMailboxes(),
		}
		NewMockServers(w)
		return ContextWithWorld(ctx, w), nil
	})

	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w := WorldFromContext(ctx)
		if w != nil {
			CloseServers(w)
		}
		return ctx, nil
	})

	registerMailSteps(sc)
	registerMCPSteps(sc)
	registerMailboxSteps(sc)
	registerMaskedEmailSteps(sc)
	registerVacationSteps(sc)
	registerIdentitySteps(sc)
	registerSieveSteps(sc)
	registerQuotaSteps(sc)
	registerMCPResourceSteps(sc)
	registerExportSteps(sc)
	registerAuthSteps(sc)
	registerContactSteps(sc)
	registerCalendarSteps(sc)
}
