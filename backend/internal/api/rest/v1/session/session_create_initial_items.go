package sessionapi

import (
	"context"
	"encoding/json"
	"time"

	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
)

func (d *Deps) persistInitialUserItems(ctx context.Context, sessionID string, items []json.RawMessage) {
	if d.Items == nil || len(items) == 0 {
		return
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
			continue
		}
		respID, err := itemsvc.NewResponseID()
		if err != nil {
			continue
		}
		pos, err := d.Items.NextPosition(ctx, sessionID)
		if err != nil {
			continue
		}
		payload, _ := json.Marshal(map[string]any{
			"id": itemID, "type": "message", "response_id": respID, "status": "completed",
			"role": "user", "content": content,
		})
		_ = d.Items.Append(ctx, &domainitem.Item{
			ID: itemID, SessionID: sessionID, ItemType: "message", ResponseID: respID,
			Status: "completed", Position: pos, Payload: payload, CreatedAt: time.Now(),
		})
		if d.Stream != nil {
			d.Stream.PublishInputConsumed(sessionID, itemID, "", content)
		}
	}
}
