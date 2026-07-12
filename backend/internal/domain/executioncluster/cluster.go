package executioncluster

import (
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const (
	KindOnline = "online"
	KindLocal  = "local"
)

const (
	StatusReady   = "ready"
	StatusPending = "pending"
	StatusOffline = "offline"
)

type Cluster struct {
	ID             int64        `gorm:"primaryKey" json:"id"`
	OrganizationID int64        `gorm:"not null;index" json:"organization_id"`
	Slug           slugkit.Slug `gorm:"size:100;not null" json:"slug"`
	Name           string       `gorm:"size:255;not null" json:"name"`
	Kind           string       `gorm:"size:32;not null" json:"kind"`
	Status         string       `gorm:"size:32;not null" json:"status"`
	CreatedAt      time.Time    `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time    `gorm:"not null;default:now()" json:"updated_at"`
}

func (Cluster) TableName() string {
	return "execution_clusters"
}

func (c *Cluster) ValidateIdentifiers() error {
	return slugkit.ValidateIdentifier("execution_clusters.slug", c.Slug.String())
}

func (c Cluster) Validate() error {
	if c.OrganizationID <= 0 {
		return errors.New("execution cluster organization_id is required")
	}
	if c.Name == "" {
		return errors.New("execution cluster name is required")
	}
	if c.Slug == "" {
		return errors.New("execution cluster slug is required")
	}
	if err := c.ValidateIdentifiers(); err != nil {
		return err
	}
	if c.Kind != KindOnline && c.Kind != KindLocal {
		return errors.New("execution cluster kind is invalid")
	}
	if c.Status != StatusReady && c.Status != StatusPending && c.Status != StatusOffline {
		return errors.New("execution cluster status is invalid")
	}
	return nil
}
