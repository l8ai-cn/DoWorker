package agentsession

import "time"

type Session struct {
	ID              string `gorm:"primaryKey;size:100"`
	OrganizationID  int64  `gorm:"not null;index:idx_agent_sessions_org_user,priority:1"`
	UserID          int64  `gorm:"not null;index:idx_agent_sessions_org_user,priority:2"`
	PodKey          string `gorm:"size:100;not null;uniqueIndex"`
	AgentSlug       string `gorm:"size:50;not null"`
	RunnerNodeID    *string `gorm:"size:100"`
	Title           *string `gorm:"type:text"`
	Status          string `gorm:"size:20;not null;default:idle"`
	ParentSessionID *string `gorm:"size:100"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (Session) TableName() string { return "agent_sessions" }
