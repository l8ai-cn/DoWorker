package conversationitem

import (
	"encoding/json"
	"time"
)

type Item struct {
	ID         string          `gorm:"primaryKey;size:100"`
	SessionID  string          `gorm:"size:100;not null;index:idx_conversation_items_session_pos,priority:1"`
	ItemType   string          `gorm:"size:50;not null"`
	ResponseID string          `gorm:"size:100;not null"`
	Status     string          `gorm:"size:20;not null;default:completed"`
	Position   int64           `gorm:"not null;index:idx_conversation_items_session_pos,priority:2,sort:desc"`
	Payload    json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time       `gorm:"not null"`
}

func (Item) TableName() string { return "conversation_items" }
