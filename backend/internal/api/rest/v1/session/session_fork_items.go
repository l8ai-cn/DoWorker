package sessionapi

import (
	"encoding/json"
	"errors"

	itemdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
	"github.com/gin-gonic/gin"
)

func (d *Deps) copyConversationItems(
	c *gin.Context,
	writer *itemsvc.Service,
	sourceID, destID string,
	upToResponseID *string,
) error {
	targetResponseID := ""
	if upToResponseID != nil && *upToResponseID != "" {
		targetResponseID = *upToResponseID
	}
	afterID := ""
	targetFound := false
	for {
		page, err := d.Items.ListPage(c.Request.Context(), sourceID, 100, afterID, false)
		if err != nil {
			return err
		}
		for _, src := range page.Items {
			if targetFound && src.ResponseID != targetResponseID {
				return nil
			}
			if err := d.copyConversationItem(c, writer, destID, src); err != nil {
				return err
			}
			if targetResponseID != "" && src.ResponseID == targetResponseID {
				targetFound = true
			}
		}
		if !page.HasMore {
			if targetResponseID != "" && !targetFound {
				return errForkResponseNotFound
			}
			return nil
		}
		if len(page.Items) == 0 {
			return errors.New("conversation item pagination stalled")
		}
		afterID = page.Items[len(page.Items)-1].ID
	}
}

func (d *Deps) copyConversationItem(
	c *gin.Context,
	writer *itemsvc.Service,
	destID string,
	src itemdomain.Item,
) error {
	id, err := itemsvc.NewItemID()
	if err != nil {
		return err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(src.Payload, &payload); err != nil || payload == nil {
		return errors.New("conversation item payload is invalid")
	}
	encodedID, err := json.Marshal(id)
	if err != nil {
		return err
	}
	payload["id"] = encodedID
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writer.Append(c.Request.Context(), &itemdomain.Item{
		ID: id, SessionID: destID, ItemType: src.ItemType,
		ResponseID: src.ResponseID, Status: src.Status,
		Position: src.Position, Payload: encodedPayload, CreatedAt: src.CreatedAt,
	})
}
