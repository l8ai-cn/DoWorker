package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/stretchr/testify/require"
)

func assertConcurrentSingleActiveInstallation(
	t *testing.T,
	sqlDB *sql.DB,
	orchestration *service.InstallationOrchestrationService,
	repository *InstallationRepository,
) {
	t.Helper()
	plans := make([]service.InstallationPlanResult, 2)
	for index := range plans {
		var err error
		plans[index], err = orchestration.CreatePlan(
			context.Background(),
			service.CreateInstallationPlanCommand{
				MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
				ListingVersionID: 61, TargetOrganizationID: 9, ActorUserID: 14,
			},
		)
		require.NoError(t, err)
	}
	type result struct {
		execution service.ApplyExecution
		err       error
	}
	results := make(chan result, len(plans))
	for _, plan := range plans {
		command := service.ApplyInstallationCommand{
			OperationID: plan.OperationID, PlanID: plan.PlanID,
			PlanDigest: plan.PlanDigest, IdempotencyKey: plan.OperationID,
			ActorUserID: 14,
		}
		go func() {
			execution, _, err := repository.BeginApply(context.Background(), command)
			results <- result{execution: execution, err: err}
		}()
	}
	var succeeded *service.ApplyExecution
	var blocked int
	for range plans {
		outcome := <-results
		switch {
		case outcome.err == nil:
			execution := outcome.execution
			succeeded = &execution
		case errors.Is(outcome.err, service.ErrApplicationAlreadyInstalled):
			blocked++
		default:
			require.NoError(t, outcome.err)
		}
	}
	require.NotNil(t, succeeded)
	require.Equal(t, 1, blocked)
	require.NoError(t, failConcurrentInstallation(repository, *succeeded))
	assertQuotaBalances(t, sqlDB, "80.000000", "0.000000", "20.000000")
}

func failConcurrentInstallation(
	repository *InstallationRepository,
	execution service.ApplyExecution,
) error {
	_, err := repository.FailApply(
		context.Background(),
		execution,
		service.ErrRuntimeInstallationRejected,
	)
	return err
}
