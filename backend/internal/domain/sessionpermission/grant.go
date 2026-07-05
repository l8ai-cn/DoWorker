package sessionpermission

type Grant struct {
	SessionID string `gorm:"primaryKey;size:100"`
	UserID    string `gorm:"primaryKey;size:255"`
	Level     int    `gorm:"not null"`
}

func (Grant) TableName() string { return "session_permissions" }
