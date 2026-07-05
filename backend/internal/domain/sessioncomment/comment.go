package sessioncomment

import "time"

type Comment struct {
	ID            string    `gorm:"primaryKey;size:100"`
	SessionID     string    `gorm:"size:100;not null;index:idx_session_comments_session_path,priority:1"`
	Path          string    `gorm:"size:500;not null;index:idx_session_comments_session_path,priority:2"`
	StartIndex    int       `gorm:"not null"`
	EndIndex      int       `gorm:"not null"`
	Body          string    `gorm:"type:text;not null"`
	Status        string    `gorm:"size:20;not null;default:draft"`
	AnchorContent *string   `gorm:"type:text"`
	CreatedBy     *string   `gorm:"size:255"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (Comment) TableName() string { return "session_comments" }
