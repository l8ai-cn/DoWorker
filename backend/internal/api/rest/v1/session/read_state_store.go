package sessionapi

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type readStateEntry struct {
	LastSeen int64
	Unread   bool
}

type readStateRow struct {
	UserID    int64     `gorm:"primaryKey"`
	SessionID string    `gorm:"primaryKey;size:100"`
	LastSeen  int64     `gorm:"not null;default:0"`
	Unread    bool      `gorm:"not null;default:false"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (readStateRow) TableName() string { return "session_read_states" }

type ReadStateStore struct {
	db *gorm.DB
}

func NewReadStateStore(db *gorm.DB) *ReadStateStore {
	return &ReadStateStore{db: db}
}

func (s *ReadStateStore) Get(userID int64, sessionID string) (readStateEntry, bool) {
	if s == nil || s.db == nil || sessionID == "" {
		return readStateEntry{}, false
	}
	var row readStateRow
	err := s.db.Where("user_id = ? AND session_id = ?", userID, sessionID).First(&row).Error
	if err != nil {
		return readStateEntry{}, false
	}
	return readStateEntry{LastSeen: row.LastSeen, Unread: row.Unread}, true
}

func (s *ReadStateStore) Put(userID int64, sessionID string, entry readStateEntry) {
	if s == nil || s.db == nil || sessionID == "" {
		return
	}
	row := readStateRow{
		UserID: userID, SessionID: sessionID,
		LastSeen: entry.LastSeen, Unread: entry.Unread,
		UpdatedAt: time.Now(),
	}
	_ = s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "session_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_seen", "unread", "updated_at"}),
	}).Create(&row).Error
}
