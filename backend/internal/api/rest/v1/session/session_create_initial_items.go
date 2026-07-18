package sessionapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

var errSessionItemsUnavailable = errors.New("session item service unavailable")

type persistedSessionInput struct {
	id      string
	content []map[string]any
}

func initialItemsContainAttachments(items []json.RawMessage) bool {
	for _, raw := range items {
		var evt struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(raw, &evt) == nil && evt.Type == "message" && len(messageAttachments(evt.Data)) > 0 {
			return true
		}
	}
	return false
}

func persistInitialUserItems(
	ctx context.Context,
	store *itemsvc.Service,
	sessionID string,
	items []json.RawMessage,
) ([]persistedSessionInput, error) {
	if len(items) == 0 {
		return nil, nil
	}
	if store == nil {
		return nil, errSessionItemsUnavailable
	}
	persisted := make([]persistedSessionInput, 0, len(items))
	for _, raw := range items {
		var evt struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(raw, &evt) != nil || evt.Type != "message" {
			continue
		}
		var data struct {
			Role string `json:"role"`
		}
		if json.Unmarshal(evt.Data, &data) != nil || data.Role != "user" {
			continue
		}
		content, prompt := parseMessageContent(evt.Data)
		if !messageHasContent(content) || prompt == "" {
			continue
		}
		itemID, err := itemsvc.NewItemID()
		if err != nil {
			return nil, err
		}
		respID, err := itemsvc.NewResponseID()
		if err != nil {
			return nil, err
		}
		pos, err := store.NextPosition(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		payload, err := json.Marshal(map[string]any{
			"id": itemID, "type": "message", "response_id": respID, "status": "completed",
			"role": "user", "content": content,
		})
		if err != nil {
			return nil, err
		}
		if err := store.Append(ctx, &domainitem.Item{
			ID: itemID, SessionID: sessionID, ItemType: "message", ResponseID: respID,
			Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
		}); err != nil {
			return nil, err
		}
		persisted = append(persisted, persistedSessionInput{id: itemID, content: content})
	}
	return persisted, nil
}
