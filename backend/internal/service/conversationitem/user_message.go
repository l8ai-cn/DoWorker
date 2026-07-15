package conversationitem

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
)

type PositionedAppender interface {
	NextPosition(context.Context, string) (int64, error)
	Append(context.Context, *domain.Item) error
}

func AppendUserText(
	ctx context.Context,
	items PositionedAppender,
	sessionID string,
	text string,
) error {
	if items == nil {
		return errors.New("conversation item service unavailable")
	}
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("session is required")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	itemID, err := NewItemID()
	if err != nil {
		return err
	}
	responseID, err := NewResponseID()
	if err != nil {
		return err
	}
	position, err := items.NextPosition(ctx, sessionID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{
		"id":          itemID,
		"type":        "message",
		"response_id": responseID,
		"status":      "completed",
		"role":        "user",
		"content": []map[string]string{{
			"type": "input_text",
			"text": text,
		}},
	})
	if err != nil {
		return err
	}
	return items.Append(ctx, &domain.Item{
		ID:         itemID,
		SessionID:  sessionID,
		ItemType:   "message",
		ResponseID: responseID,
		Status:     "completed",
		Position:   position,
		Payload:    payload,
		CreatedAt:  time.Now(),
	})
}
