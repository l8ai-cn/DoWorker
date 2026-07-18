package runner

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
)

const (
	minQueueTTL = time.Minute
	maxQueueTTL = 24 * time.Hour
)

type PendingCommandQueue struct {
	repo          agentpod.PendingCommandRepository
	eventBus      *eventbus.EventBus
	maxPerRunner  int
	defaultTTL    time.Duration
	enabled       bool
	payloadCipher *pendingPayloadCipher
	connChecker   ConnectionChecker
	drainer       *PendingCommandDrainer
	logger        *slog.Logger
}

func NewPendingCommandQueue(
	repo agentpod.PendingCommandRepository,
	eventBus *eventbus.EventBus,
	maxPerRunner int,
	defaultTTL time.Duration,
	enabled bool,
	payloadEncryptor *crypto.Encryptor,
	logger *slog.Logger,
) *PendingCommandQueue {
	if logger == nil {
		logger = slog.Default()
	}
	return &PendingCommandQueue{
		repo:          repo,
		eventBus:      eventBus,
		maxPerRunner:  maxPerRunner,
		defaultTTL:    defaultTTL,
		enabled:       enabled,
		payloadCipher: newPendingPayloadCipher(payloadEncryptor),
		logger:        logger.With("component", "pending_command_queue"),
	}
}

func (q *PendingCommandQueue) SetDrainer(d *PendingCommandDrainer) {
	q.drainer = d
}

func (q *PendingCommandQueue) SetConnectionChecker(c ConnectionChecker) {
	q.connChecker = c
}

func (q *PendingCommandQueue) Enabled() bool {
	return q.enabled
}

func (q *PendingCommandQueue) SendPromptTTL() time.Duration {
	ttl := q.defaultTTL
	if ttl <= 0 {
		ttl = minQueueTTL
	}
	if ttl < minQueueTTL {
		return minQueueTTL
	}
	if ttl > maxQueueTTL {
		return maxQueueTTL
	}
	return ttl
}

func (q *PendingCommandQueue) MaxPerRunner() int {
	return q.maxPerRunner
}

func (q *PendingCommandQueue) AllowsDurableCommand(runnerID int64) bool {
	if q.enabled {
		return true
	}
	// Connected dispatches still need a transactional outbox row so owner
	// persistence and Runner delivery cannot be separated by a process crash.
	return q.connChecker != nil && q.connChecker.IsConnected(runnerID)
}

func (q *PendingCommandQueue) TriggerDrain(runnerID int64) {
	q.maybeDrain(runnerID)
}

func (q *PendingCommandQueue) CancelByPodKey(ctx context.Context, podKey string) error {
	_, err := q.repo.DeleteByPodKey(ctx, podKey)
	return err
}

func (q *PendingCommandQueue) QueuePosition(ctx context.Context, runnerID int64, podKey string) (int, error) {
	return q.repo.PositionByPodKey(ctx, runnerID, podKey)
}

func (q *PendingCommandQueue) enqueue(
	ctx context.Context,
	orgID, runnerID int64,
	podKey, commandType, commandID string,
	payload []byte,
	ttl time.Duration,
) (time.Time, error) {
	if !q.enabled {
		return time.Time{}, ErrRunnerNotConnected
	}
	if ttl <= 0 {
		ttl = q.SendPromptTTL()
	}
	if ttl < minQueueTTL {
		ttl = minQueueTTL
	}
	if ttl > maxQueueTTL {
		ttl = maxQueueTTL
	}
	encryptedPayload, err := q.payloadCipher.encrypt(payload)
	if err != nil {
		return time.Time{}, err
	}
	expiresAt := time.Now().Add(ttl)
	cmd := &agentpod.PendingCommand{
		OrganizationID: orgID,
		RunnerID:       runnerID,
		PodKey:         podKey,
		CommandType:    commandType,
		CommandID:      commandID,
		Payload:        encryptedPayload,
		ExpiresAt:      expiresAt,
	}
	if err := q.repo.EnqueueWithinCapacity(ctx, cmd, q.maxPerRunner); err != nil {
		return time.Time{}, err
	}
	pos, _ := q.repo.PositionByPodKey(ctx, runnerID, podKey)
	publishQueueEvent(q.eventBus, q.logger, eventbus.EventPodQueued, orgID, podKey, map[string]interface{}{
		"pod_key":        podKey,
		"runner_id":      runnerID,
		"queue_position": pos,
		"expires_at":     expiresAt.UTC().Format(time.RFC3339),
	})
	return expiresAt, nil
}

func (q *PendingCommandQueue) maybeDrain(runnerID int64) {
	if q.drainer == nil || q.connChecker == nil || !q.connChecker.IsConnected(runnerID) {
		return
	}
	q.drainer.DrainRunner(runnerID)
}

func (q *PendingCommandQueue) GetCreatePodExpiry(ctx context.Context, podKey string) (time.Time, error) {
	cmd, err := q.repo.GetCreatePodByPodKey(ctx, podKey)
	if err != nil || cmd == nil {
		return time.Time{}, err
	}
	return cmd.ExpiresAt, nil
}

func marshalServerMessage(msg *runnerv1.ServerMessage) ([]byte, error) {
	return proto.Marshal(msg)
}

func publishQueueEvent(bus *eventbus.EventBus, logger *slog.Logger, typ eventbus.EventType, orgID int64, podKey string, data map[string]interface{}) {
	if bus == nil {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Error("failed to marshal queue event", "type", typ, "error", err)
		return
	}
	event := &eventbus.Event{
		Type:           typ,
		Category:       eventbus.CategoryEntity,
		OrganizationID: orgID,
		EntityType:     "pod",
		EntityID:       podKey,
		Data:           raw,
		Timestamp:      time.Now().UnixMilli(),
	}
	if err := bus.Publish(context.Background(), event); err != nil {
		logger.Error("failed to publish queue event", "type", typ, "error", err)
	}
}
