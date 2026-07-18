package runner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type queueExpiryMarker interface {
	MarkQueueExpired(ctx context.Context, podKey, errorCode, errorMessage string) error
}

type PendingCommandDrainer struct {
	repo           agentpod.PendingCommandRepository
	podStore       PodStore
	runnerRepo     runnerDomain.RunnerRepository
	msgSender      ServerMessageSender
	connChecker    ConnectionChecker
	coordinator    *PodCoordinator
	expiryMarker   queueExpiryMarker
	eventBus       *eventbus.EventBus
	payloadCipher  *pendingPayloadCipher
	sweepInterval  time.Duration
	inflight       sync.Map
	onQueueExpired func(ctx context.Context, podKey string)
	logger         *slog.Logger
}

func NewPendingCommandDrainer(
	repo agentpod.PendingCommandRepository,
	podStore PodStore,
	runnerRepo runnerDomain.RunnerRepository,
	msgSender ServerMessageSender,
	connChecker ConnectionChecker,
	coordinator *PodCoordinator,
	expiryMarker queueExpiryMarker,
	eventBus *eventbus.EventBus,
	sweepInterval time.Duration,
	payloadEncryptor *crypto.Encryptor,
	logger *slog.Logger,
) *PendingCommandDrainer {
	if logger == nil {
		logger = slog.Default()
	}
	if sweepInterval <= 0 {
		sweepInterval = 60 * time.Second
	}
	return &PendingCommandDrainer{
		repo:          repo,
		podStore:      podStore,
		runnerRepo:    runnerRepo,
		msgSender:     msgSender,
		connChecker:   connChecker,
		coordinator:   coordinator,
		expiryMarker:  expiryMarker,
		eventBus:      eventBus,
		payloadCipher: newPendingPayloadCipher(payloadEncryptor),
		sweepInterval: sweepInterval,
		logger:        logger.With("component", "pending_command_drainer"),
	}
}

func (d *PendingCommandDrainer) DrainRunner(runnerID int64) {
	if _, loaded := d.inflight.LoadOrStore(runnerID, struct{}{}); loaded {
		return
	}
	go func() {
		defer d.inflight.Delete(runnerID)
		d.drainRunner(context.Background(), runnerID)
	}()
}

func (d *PendingCommandDrainer) SetMessageSender(sender ServerMessageSender) {
	d.msgSender = sender
}

func (d *PendingCommandDrainer) SetQueueExpiredNotifier(fn func(ctx context.Context, podKey string)) {
	d.onQueueExpired = fn
}

func (d *PendingCommandDrainer) drainRunner(ctx context.Context, runnerID int64) {
	if d.msgSender == nil {
		return
	}
	for {
		if d.connChecker == nil || !d.connChecker.IsConnected(runnerID) {
			return
		}
		batch, err := d.repo.ListByRunnerFIFO(ctx, runnerID, 10)
		if err != nil || len(batch) == 0 {
			return
		}
		for _, row := range batch {
			if time.Now().After(row.ExpiresAt) {
				d.handleExpired(ctx, row)
				continue
			}
			if !d.connChecker.IsConnected(runnerID) {
				return
			}
			if ok, stop := d.dispatchOne(ctx, runnerID, row); stop || !ok {
				return
			}
		}
	}
}

func (d *PendingCommandDrainer) dispatchCreatePod(ctx context.Context, runnerID int64, row *agentpod.PendingCommand, cmd *runnerv1.CreatePodCommand) (bool, bool) {
	if cmd == nil {
		_ = d.repo.Delete(ctx, row.ID)
		return true, false
	}
	run, err := d.runnerRepo.GetByID(ctx, runnerID)
	if err != nil || run == nil {
		return false, true
	}
	if run.CurrentPods >= run.MaxConcurrentPods {
		return false, true
	}
	// CAS claim closes the cancel race: cancel does queued→completed,
	// drain does queued→initializing — exactly one transition wins.
	claimed, err := d.podStore.UpdateByKeyAndStatusCounted(ctx, row.PodKey, agentpod.StatusQueued, map[string]interface{}{
		"status": agentpod.StatusInitializing,
	})
	if err != nil {
		return false, true
	}
	if claimed == 0 {
		_ = d.repo.Delete(ctx, row.ID)
		return true, false
	}
	if err := d.coordinator.IncrementPods(ctx, runnerID); err != nil {
		d.revertClaim(ctx, row.PodKey)
		return false, true
	}
	sendErr := d.msgSender.SendServerMessage(ctx, runnerID, &runnerv1.ServerMessage{
		Payload:   &runnerv1.ServerMessage_CreatePod{CreatePod: cmd},
		Timestamp: time.Now().UnixMilli(),
	})
	if sendErr != nil {
		_ = d.coordinator.DecrementPods(ctx, runnerID)
		d.revertClaim(ctx, row.PodKey)
		return false, true
	}
	_ = d.repo.Delete(ctx, row.ID)
	publishQueueEvent(d.eventBus, d.logger, eventbus.EventPodQueueDispatched, row.OrganizationID, row.PodKey, map[string]interface{}{
		"pod_key":        row.PodKey,
		"runner_id":      runnerID,
		"waited_seconds": time.Since(row.CreatedAt).Seconds(),
	})
	return true, false
}

func (d *PendingCommandDrainer) revertClaim(ctx context.Context, podKey string) {
	if _, err := d.podStore.UpdateByKeyAndStatusCounted(ctx, podKey, agentpod.StatusInitializing, map[string]interface{}{
		"status": agentpod.StatusQueued,
	}); err != nil {
		d.logger.Error("failed to revert queued-pod claim", "pod_key", podKey, "error", err)
	}
}

func (d *PendingCommandDrainer) dispatchSendPrompt(ctx context.Context, runnerID int64, row *agentpod.PendingCommand, cmd *runnerv1.SendPromptCommand) (bool, bool) {
	if cmd == nil {
		_ = d.repo.Delete(ctx, row.ID)
		return true, false
	}
	pod, err := d.podStore.GetByKey(ctx, row.PodKey)
	if err != nil || pod == nil || !pod.IsActive() {
		_ = d.repo.Delete(ctx, row.ID)
		return true, false
	}
	if err := d.msgSender.SendServerMessage(ctx, runnerID, &runnerv1.ServerMessage{
		Payload:   &runnerv1.ServerMessage_SendPrompt{SendPrompt: cmd},
		Timestamp: time.Now().UnixMilli(),
	}); err != nil {
		return false, true
	}
	_ = d.repo.Delete(ctx, row.ID)
	return true, false
}
