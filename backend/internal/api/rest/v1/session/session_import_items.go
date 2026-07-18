package sessionapi

import (
	"context"
	"encoding/json"
	"time"

	itemdomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/codeximport"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

func importConversationItems(
	ctx context.Context,
	store *itemsvc.Service,
	sessionID string,
	items []codeximport.Item,
) error {
	currentResp, err := itemsvc.NewResponseID()
	if err != nil {
		return err
	}
	now := time.Now()
	for i, src := range items {
		if src.StartsTurn && i != 0 {
			currentResp, err = itemsvc.NewResponseID()
			if err != nil {
				return err
			}
		}
		itemID, err := itemsvc.NewItemID()
		if err != nil {
			return err
		}
		status := src.Status
		if status == "" {
			status = "completed"
		}
		payload := src.Payload
		if payload == nil {
			payload = map[string]any{}
		}
		payload["id"] = itemID
		payload["response_id"] = currentResp
		payload["status"] = status
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if err := store.Append(ctx, &itemdomain.Item{
			ID: itemID, SessionID: sessionID, ItemType: src.Type,
			ResponseID: currentResp, Status: status, Position: int64(i + 1),
			Payload: encoded, CreatedAt: now.Add(time.Duration(i) * time.Millisecond),
		}); err != nil {
			return err
		}
	}
	return nil
}
