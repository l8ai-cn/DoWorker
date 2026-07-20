package sessionmessage

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"gorm.io/gorm"
)

var ErrPromptCommandConflict = errors.New("session prompt command conflict")

func persistPromptCommand(
	ctx context.Context,
	tx *gorm.DB,
	in PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	maxPerRunner int,
	promptTTL time.Duration,
) error {
	var existing domainitem.Item
	err := tx.Where("id = ?", in.Item.ID).Take(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return insertPromptCommand(ctx, tx, in, payload, queue, maxPerRunner, promptTTL)
	case err != nil:
		return err
	case !samePromptItem(existing, *in.Item):
		return ErrPromptCommandConflict
	}
	return ensurePendingPrompt(ctx, tx, in, payload, queue, maxPerRunner, promptTTL)
}

func insertPromptCommand(
	ctx context.Context,
	tx *gorm.DB,
	in PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	maxPerRunner int,
	promptTTL time.Duration,
) error {
	var command agentpod.PendingCommand
	err := tx.Where("command_id = ?", in.Item.ID).Take(&command).Error
	if err == nil {
		return ErrPromptCommandConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := ensureQueueCapacity(ctx, tx, in.RunnerID, maxPerRunner); err != nil {
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
	return createPendingPrompt(tx, in, payload, queue, promptTTL)
}

func ensurePendingPrompt(
	ctx context.Context,
	tx *gorm.DB,
	in PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	maxPerRunner int,
	promptTTL time.Duration,
) error {
	var command agentpod.PendingCommand
	err := tx.Where("command_id = ?", in.Item.ID).Take(&command).Error
	if err == nil {
		same, matchErr := samePendingPrompt(command, in, payload, queue)
		if matchErr != nil {
			return matchErr
		}
		if same {
			return nil
		}
		return ErrPromptCommandConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := ensureQueueCapacity(ctx, tx, in.RunnerID, maxPerRunner); err != nil {
		return err
	}
	return createPendingPrompt(tx, in, payload, queue, promptTTL)
}

func createPendingPrompt(
	tx *gorm.DB,
	in PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	promptTTL time.Duration,
) error {
	sealedPayload, err := queue.SealPayload(payload)
	if err != nil {
		return err
	}
	return tx.Create(&agentpod.PendingCommand{
		OrganizationID: in.OrganizationID,
		RunnerID:       in.RunnerID,
		PodKey:         in.PodKey,
		CommandType:    agentpod.CommandTypeSendPrompt,
		CommandID:      in.Item.ID,
		Payload:        sealedPayload,
		ExpiresAt:      time.Now().Add(promptTTL),
	}).Error
}

func samePromptItem(left, right domainitem.Item) bool {
	return left.ID == right.ID &&
		left.SessionID == right.SessionID &&
		left.ItemType == right.ItemType &&
		left.ResponseID == right.ResponseID &&
		left.Status == right.Status &&
		bytes.Equal(left.Payload, right.Payload)
}

func samePendingPrompt(
	command agentpod.PendingCommand,
	in PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
) (bool, error) {
	if command.OrganizationID != in.OrganizationID ||
		command.RunnerID != in.RunnerID ||
		command.PodKey != in.PodKey ||
		command.CommandType != agentpod.CommandTypeSendPrompt ||
		command.CommandID != in.Item.ID {
		return false, nil
	}
	return queue.PayloadMatches(command.Payload, payload)
}
