package sessionmessage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var ErrUnavailable = errors.New("session prompt outbox unavailable")

type PromptOutbox struct {
	db    *gorm.DB
	queue *runnerservice.PendingCommandQueue
}

type PromptInput struct {
	OrganizationID int64
	RunnerID       int64
	PodKey         string
	Item           *domainitem.Item
	Prompt         string
}

func NewPromptOutbox(db *gorm.DB, queue *runnerservice.PendingCommandQueue) *PromptOutbox {
	return &PromptOutbox{db: db, queue: queue}
}

func (s *PromptOutbox) PersistAndQueue(ctx context.Context, in PromptInput) error {
	if s == nil || s.db == nil || s.queue == nil || !s.queue.Enabled() || in.Item == nil {
		return ErrUnavailable
	}
	payload, err := promptCommandPayload(in.PodKey, in.Item.ID, in.Prompt)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := acquirePromptLocks(tx, in.RunnerID, in.Item.SessionID); err != nil {
			return err
		}
		if err := ensureQueueCapacity(ctx, tx, in.RunnerID, s.queue.MaxPerRunner()); err != nil {
			return err
		}
		position, err := nextItemPosition(ctx, tx, in.Item.SessionID)
		if err != nil {
			return err
		}
		in.Item.Position = position
		if err := tx.Create(in.Item).Error; err != nil {
			return err
		}
		return tx.Create(&agentpod.PendingCommand{
			OrganizationID: in.OrganizationID,
			RunnerID:       in.RunnerID,
			PodKey:         in.PodKey,
			CommandType:    agentpod.CommandTypeSendPrompt,
			CommandID:      in.Item.ID,
			Payload:        payload,
			ExpiresAt:      time.Now().Add(s.queue.SendPromptTTL()),
		}).Error
	}); err != nil {
		return err
	}
	s.queue.TriggerDrain(in.RunnerID)
	return nil
}

func acquirePromptLocks(tx *gorm.DB, runnerID int64, sessionID string) error {
	if tx.Name() != "postgres" {
		return nil
	}
	for _, key := range []int64{
		promptLockKey("runner", strconv.FormatInt(runnerID, 10)),
		promptLockKey("session", sessionID),
	} {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", key).Error; err != nil {
			return fmt.Errorf("acquire prompt outbox lock: %w", err)
		}
	}
	return nil
}

func promptLockKey(scope, value string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(scope))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(value))
	return int64(h.Sum64())
}

func promptCommandPayload(podKey, commandID, prompt string) ([]byte, error) {
	return proto.Marshal(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_SendPrompt{
			SendPrompt: &runnerv1.SendPromptCommand{
				PodKey: podKey, CommandId: commandID, Prompt: prompt,
			},
		},
	})
}

func ensureQueueCapacity(ctx context.Context, tx *gorm.DB, runnerID int64, max int) error {
	var count int64
	if err := tx.WithContext(ctx).Model(&agentpod.PendingCommand{}).
		Where("runner_id = ?", runnerID).Count(&count).Error; err != nil {
		return err
	}
	if int(count) >= max {
		return agentpod.ErrQueueFull
	}
	return nil
}

func nextItemPosition(ctx context.Context, tx *gorm.DB, sessionID string) (int64, error) {
	var position int64
	err := tx.WithContext(ctx).Model(&domainitem.Item{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(position), 0) + 1").
		Scan(&position).Error
	if err != nil {
		return 0, fmt.Errorf("next conversation item position: %w", err)
	}
	return position, nil
}

func UserItem(id, sessionID, responseID string, content any) (*domainitem.Item, error) {
	payload, err := json.Marshal(map[string]any{
		"id": id, "type": "message", "response_id": responseID, "status": "completed",
		"role": "user", "content": content,
	})
	if err != nil {
		return nil, err
	}
	return &domainitem.Item{
		ID: id, SessionID: sessionID, ItemType: "message", ResponseID: responseID,
		Status: "completed", Payload: payload, CreatedAt: time.Now(),
	}, nil
}
