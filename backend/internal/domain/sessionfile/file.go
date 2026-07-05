package sessionfile

import "time"

type File struct {
	ID          string    `gorm:"primaryKey;size:100"`
	SessionID   string    `gorm:"size:100;not null;index:idx_session_files_session,priority:1,sort:desc"`
	Filename    string    `gorm:"size:255;not null"`
	Bytes       int64     `gorm:"not null"`
	ContentType string    `gorm:"size:100;not null"`
	MinioKey    string    `gorm:"size:500;not null;uniqueIndex"`
	CreatedAt   time.Time `gorm:"not null;index:idx_session_files_session,priority:2,sort:desc"`
}

func (File) TableName() string { return "session_files" }
