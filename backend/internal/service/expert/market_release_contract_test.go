package expert

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	resourcedom "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func TestMarketSubmissionRejectsPublishedRuntimeContractChanges(t *testing.T) {
	type runtimeMutation struct {
		setup  func(*specdomain.Spec)
		mutate func(*specdomain.Spec)
	}
	tests := map[string]runtimeMutation{
		"worker definition": {
			setup: func(*specdomain.Spec) {},
			mutate: func(spec *specdomain.Spec) {
				spec.Runtime.WorkerType.DefinitionHash = strings.Repeat("c", 64)
			},
		},
		"tool model roles": {
			setup: func(spec *specdomain.Spec) {
				spec.Runtime.ToolModelBindings = []specdomain.ToolModelBinding{
					validMarketToolModelBinding(),
				}
			},
			mutate: func(spec *specdomain.Spec) {
				spec.Runtime.ToolModelBindings = nil
			},
		},
	}
	for name, mutation := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newMarketServiceFixture(t)
			mutation.setup(&fixture.snapshots.source.Spec)
			ctx := context.Background()
			first, err := fixture.service.SubmitMarketApplication(
				ctx,
				fixture.submissionRequest(),
			)
			require.NoError(t, err)
			_, err = fixture.service.ApproveMarketRelease(
				ctx,
				ReviewMarketReleaseRequest{
					ReviewerUserID: 99,
					ReleaseID:      first.Release.ID,
				},
			)
			require.NoError(t, err)

			mutation.mutate(&fixture.snapshots.source.Spec)
			_, err = fixture.service.SubmitMarketApplication(
				ctx,
				fixture.submissionRequest(),
			)

			require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
			require.ErrorContains(t, err, "worker runtime contract")
			require.Len(t, fixture.market.releases, 1)
		})
	}
}

func TestMarketSubmissionKeepsRuntimeContractAfterWithdrawal(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	first, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(ctx, ReviewMarketReleaseRequest{
		ReviewerUserID: 99,
		ReleaseID:      first.Release.ID,
	})
	require.NoError(t, err)
	_, err = fixture.service.WithdrawMarketRelease(ctx, WithdrawMarketReleaseRequest{
		PublisherOrganizationID: fixture.source.OrganizationID,
		ReleaseID:               first.Release.ID,
	})
	require.NoError(t, err)
	fixture.snapshots.source.Spec.Runtime.WorkerType.DefinitionHash =
		strings.Repeat("c", 64)

	_, err = fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)

	require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
	require.Len(t, fixture.market.releases, 1)
}

func TestMarketSubmissionRevalidatesContractInsideApplicationLock(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	first, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	fixture.snapshots.source.Spec.Runtime.WorkerType.DefinitionHash =
		strings.Repeat("c", 64)
	fixture.locker.applicationHook = func() {
		pending := expertmarket.ReleaseStatusPendingReview
		require.NoError(t, fixture.market.UpdateReleaseLifecycleAndLatest(
			ctx,
			first.Application.ID,
			first.Release.ID,
			expertmarket.LifecycleUpdate{
				Status:         expertmarket.ReleaseStatusPublished,
				ExpectedStatus: &pending,
			},
		))
	}

	_, err = fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)

	require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
	require.Len(t, fixture.market.releases, 1)
}

func validMarketToolModelBinding() specdomain.ToolModelBinding {
	return specdomain.ToolModelBinding{
		Role: slugkit.MustNewForTest("seedance-video"),
		ModelBinding: specdomain.ModelBinding{
			ResourceID:         111,
			ResourceRevision:   1,
			ConnectionID:       211,
			ConnectionRevision: 1,
			ProviderKey:        slugkit.MustNewForTest("doubao"),
			ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
			ModelID:            "doubao-seedance-2-0-260128",
		},
		Modality:   resourcedom.ModalityVideo,
		Capability: resourcedom.CapabilityVideoGeneration,
		Environment: specdomain.ToolModelEnvironment{
			APIKey:  "SEEDANCE_API_KEY",
			BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}
}
