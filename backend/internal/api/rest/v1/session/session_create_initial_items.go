package sessionapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

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

func (d *Deps) persistInitialUserItems(ctx context.Context, sessionID string, items []json.RawMessage) error {
	if len(items) == 0 {
		return nil
	}
	if d.Items == nil {
		return errors.New("conversation item service unavailable")
	}
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
			return err
		}
		respID, err := itemsvc.NewResponseID()
		if err != nil {
			return err
		}
		pos, err := d.Items.NextPosition(ctx, sessionID)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"id": itemID, "type": "message", "response_id": respID, "status": "completed",
			"role": "user", "content": content,
		})
		if err != nil {
			return err
		}
		if err := d.Items.Append(ctx, &domainitem.Item{
			ID: itemID, SessionID: sessionID, ItemType: "message", ResponseID: respID,
			Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
		}); err != nil {
			return err
		}
		if d.Stream != nil {
			d.Stream.PublishInputConsumed(sessionID, itemID, "", content)
		}
	}
	return nil
}
