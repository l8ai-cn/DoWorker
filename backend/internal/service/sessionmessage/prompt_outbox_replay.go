package sessionmessage

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"gorm.io/gorm"
)

var ErrPromptCommandConflict = errors.New("session prompt command conflict")

func persistPromptCommand(
	ctx context.Context,
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
	maxPerRunner int,
	expiresAt time.Time,
) error {
	var existing domainitem.Item
	err := tx.Where("id = ?", input.Item.ID).Take(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return insertPromptCommand(ctx, tx, input, payload, maxPerRunner, expiresAt)
	case err != nil:
		return err
	case !samePromptItem(existing, *input.Item):
		return ErrPromptCommandConflict
	}
	return ensurePendingPrompt(ctx, tx, input, payload, maxPerRunner, expiresAt)
}

func insertPromptCommand(
	ctx context.Context,
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
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
	return createPendingPrompt(tx, input, payload, expiresAt)
}

func ensurePendingPrompt(
	ctx context.Context,
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
	maxPerRunner int,
	expiresAt time.Time,
) error {
	var command agentpod.PendingCommand
	err := tx.Where("command_id = ?", input.Item.ID).Take(&command).Error
	if err == nil {
		if samePendingPrompt(command, input, payload) {
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
	return createPendingPrompt(tx, input, payload, expiresAt)
}

func createPendingPrompt(
	tx *gorm.DB,
	input PromptInput,
	payload []byte,
	expiresAt time.Time,
) error {
	return tx.Create(&agentpod.PendingCommand{
		OrganizationID: input.OrganizationID,
		RunnerID:       input.RunnerID,
		PodKey:         input.PodKey,
		CommandType:    agentpod.CommandTypeSendPrompt,
		CommandID:      input.Item.ID,
		Payload:        payload,
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
) bool {
	return command.OrganizationID == input.OrganizationID &&
		command.RunnerID == input.RunnerID &&
		command.PodKey == input.PodKey &&
		command.CommandType == agentpod.CommandTypeSendPrompt &&
		command.CommandID == input.Item.ID &&
		bytes.Equal(command.Payload, payload)
}
