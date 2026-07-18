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
	input PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	maxPerRunner int,
	expiresAt time.Time,
) error {
	var existing domainitem.Item
	err := tx.Where("id = ?", input.Item.ID).Take(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return insertPromptCommand(ctx, tx, input, payload, queue, maxPerRunner, expiresAt)
	case err != nil:
		return err
	case !samePromptItem(existing, *input.Item):
		return ErrPromptCommandConflict
	}
	return ensurePendingPrompt(ctx, tx, input, payload, queue, maxPerRunner, expiresAt)
}

func insertPromptCommand(
	ctx context.Context,
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	maxPerRunner int,
	expiresAt time.Time,
) error {
	var command agentpod.PendingCommand
	err := tx.Where("command_id = ?", input.Item.ID).Take(&command).Error
	if err == nil {
		return ErrPromptCommandConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := ensureQueueCapacity(ctx, tx, input.RunnerID, maxPerRunner); err != nil {
		return err
	}
	position, err := nextItemPosition(ctx, tx, input.Item.SessionID)
	if err != nil {
		return err
	}
	input.Item.Position = position
	if err := tx.Create(input.Item).Error; err != nil {
		return err
	}
	return createPendingPrompt(tx, input, payload, queue, expiresAt)
}

func ensurePendingPrompt(
	ctx context.Context,
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	maxPerRunner int,
	expiresAt time.Time,
) error {
	var command agentpod.PendingCommand
	err := tx.Where("command_id = ?", input.Item.ID).Take(&command).Error
	if err == nil {
		same, matchErr := samePendingPrompt(command, input, payload, queue)
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
	if err := ensureQueueCapacity(ctx, tx, input.RunnerID, maxPerRunner); err != nil {
		return err
	}
	return createPendingPrompt(tx, input, payload, queue, expiresAt)
}

func createPendingPrompt(
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
	expiresAt time.Time,
) error {
	encryptedPayload, err := queue.SealPayload(payload)
	if err != nil {
		return err
	}
	return tx.Create(&agentpod.PendingCommand{
		OrganizationID: input.OrganizationID,
		RunnerID:       input.RunnerID,
		PodKey:         input.PodKey,
		CommandType:    agentpod.CommandTypeSendPrompt,
		CommandID:      input.Item.ID,
		Payload:        encryptedPayload,
		ExpiresAt:      expiresAt,
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
	input PromptInput,
	payload []byte,
	queue *runnerservice.PendingCommandQueue,
) (bool, error) {
	if command.OrganizationID != input.OrganizationID ||
		command.RunnerID != input.RunnerID ||
		command.PodKey != input.PodKey ||
		command.CommandType != agentpod.CommandTypeSendPrompt ||
		command.CommandID != input.Item.ID {
		return false, nil
	}
	return queue.PayloadMatches(command.Payload, payload)
}
