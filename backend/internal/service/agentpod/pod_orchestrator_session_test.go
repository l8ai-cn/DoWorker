package agentpod

import (
	"context"
	"errors"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
)

type recordingSessionProvisioner struct {
	order       *[]string
	prepareErr  error
	rollbackErr error
	pod         *podDomain.Pod
	spec        sessionDomain.ProvisionSpec
	receipt     *sessionDomain.ProvisionReceipt
	rolledBack  *sessionDomain.ProvisionReceipt
	rollbackCtx error
}

func (p *recordingSessionProvisioner) PrepareForPod(
	_ context.Context,
	pod *podDomain.Pod,
	spec sessionDomain.ProvisionSpec,
) (*sessionDomain.ProvisionReceipt, error) {
	*p.order = append(*p.order, "session")
	p.pod = pod
	p.spec = spec
	if p.prepareErr != nil {
		return nil, p.prepareErr
	}
	p.receipt = &sessionDomain.ProvisionReceipt{
		Session: &sessionDomain.Session{ID: spec.ID, PodKey: pod.PodKey},
		Created: !spec.UpdateExisting,
	}
	return p.receipt, nil
}

func (p *recordingSessionProvisioner) RollbackProvision(
	ctx context.Context,
	receipt *sessionDomain.ProvisionReceipt,
) error {
	*p.order = append(*p.order, "rollback")
	p.rolledBack = receipt
	p.rollbackCtx = ctx.Err()
	return p.rollbackErr
}

type orderedPodCoordinator struct {
	order       *[]string
	dispatchErr error
}

func (c *orderedPodCoordinator) CreatePod(context.Context, int64, *runnerv1.CreatePodCommand) error {
	*c.order = append(*c.order, "dispatch")
	return c.dispatchErr
}

func (c *orderedPodCoordinator) CreatePodOrQueue(
	ctx context.Context,
	runnerID int64,
	cmd *runnerv1.CreatePodCommand,
	_ podDomain.CreatePodQueueOpts,
) error {
	return c.CreatePod(ctx, runnerID, cmd)
}

func TestCreatePodProvisionsSessionBeforeRunnerDispatch(t *testing.T) {
	order := []string{}
	provisioner := &recordingSessionProvisioner{order: &order}
	coordinator := &orderedPodCoordinator{order: &order}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coordinator),
		func(deps *PodOrchestratorDeps) { deps.SessionProvisioner = provisioner },
	)

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: "conv_seedance",
		},
		PrepareSession: func(_ context.Context, row *sessionDomain.Session) error {
			require.Equal(t, "conv_seedance", row.ID)
			order = append(order, "items")
			return nil
		},
	})

	require.NoError(t, err)
	require.Equal(t, []string{"session", "items", "dispatch"}, order)
	require.Same(t, result.Pod, provisioner.pod)
	require.Equal(t, "conv_seedance", provisioner.spec.ID)
}

func TestCreatePodDoesNotDispatchWhenSessionProvisionFails(t *testing.T) {
	order := []string{}
	provisioner := &recordingSessionProvisioner{
		order:      &order,
		prepareErr: errors.New("database unavailable"),
	}
	coordinator := &orderedPodCoordinator{order: &order}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coordinator),
		func(deps *PodOrchestratorDeps) { deps.SessionProvisioner = provisioner },
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: "conv_seedance",
		},
	})

	require.ErrorIs(t, err, ErrSessionProvisionFailed)
	require.Equal(t, []string{"session"}, order)
}

func TestCreatePodRollsBackSessionWhenPreparationFails(t *testing.T) {
	order := []string{}
	provisioner := &recordingSessionProvisioner{order: &order}
	coordinator := &orderedPodCoordinator{order: &order}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coordinator),
		func(deps *PodOrchestratorDeps) { deps.SessionProvisioner = provisioner },
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: "conv_seedance",
		},
		PrepareSession: func(context.Context, *sessionDomain.Session) error {
			order = append(order, "items")
			return errors.New("item insert failed")
		},
	})

	require.ErrorIs(t, err, ErrSessionPreparationFailed)
	require.Equal(t, []string{"session", "items", "rollback"}, order)
	require.Same(t, provisioner.receipt, provisioner.rolledBack)
}

func TestCreatePodRollsBackSessionAfterRequestCancellation(t *testing.T) {
	order := []string{}
	provisioner := &recordingSessionProvisioner{order: &order}
	coordinator := &orderedPodCoordinator{order: &order}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coordinator),
		func(deps *PodOrchestratorDeps) { deps.SessionProvisioner = provisioner },
	)
	ctx, cancel := context.WithCancel(context.Background())

	_, err := orch.CreatePod(ctx, &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: "conv_seedance",
		},
		PrepareSession: func(context.Context, *sessionDomain.Session) error {
			order = append(order, "items")
			cancel()
			return context.Canceled
		},
	})

	require.ErrorIs(t, err, ErrSessionPreparationFailed)
	require.Equal(t, []string{"session", "items", "rollback"}, order)
	require.NoError(t, provisioner.rollbackCtx)
}

func TestCreatePodRollsBackSessionWhenRunnerDispatchFails(t *testing.T) {
	order := []string{}
	provisioner := &recordingSessionProvisioner{order: &order}
	coordinator := &orderedPodCoordinator{
		order:       &order,
		dispatchErr: errors.New("runner offline"),
	}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coordinator),
		func(deps *PodOrchestratorDeps) { deps.SessionProvisioner = provisioner },
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: "conv_seedance",
		},
	})

	require.ErrorIs(t, err, ErrRunnerDispatchFailed)
	require.Equal(t, []string{"session", "dispatch", "rollback"}, order)
	require.Same(t, provisioner.receipt, provisioner.rolledBack)
}

func TestCreatePodRejectsSessionProvisionWithoutProvisioner(t *testing.T) {
	orch, _, _ := setupOrchestrator(t)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SessionProvision: &sessionDomain.ProvisionSpec{
			ID: "conv_seedance",
		},
	})

	require.ErrorIs(t, err, ErrSessionProvisionerUnavailable)
}
