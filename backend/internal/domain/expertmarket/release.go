package expertmarket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/lib/pq"
)

var (
	ErrInvalidStatus   = errors.New("expert market release status is invalid")
	ErrInvalidVersion  = errors.New("expert market release version is invalid")
	ErrInvalidSnapshot = errors.New("expert market release snapshot is invalid")
)

type ReleaseStatus string

const (
	ReleaseStatusDraft         ReleaseStatus = "draft"
	ReleaseStatusPendingReview ReleaseStatus = "pending_review"
	ReleaseStatusPublished     ReleaseStatus = "published"
	ReleaseStatusRejected      ReleaseStatus = "rejected"
	ReleaseStatusWithdrawn     ReleaseStatus = "withdrawn"
)

func (status ReleaseStatus) Valid() bool {
	switch status {
	case ReleaseStatusDraft,
		ReleaseStatusPendingReview,
		ReleaseStatusPublished,
		ReleaseStatusRejected,
		ReleaseStatusWithdrawn:
		return true
	default:
		return false
	}
}

type Release struct {
	ID                      int64         `gorm:"primaryKey" json:"id"`
	ApplicationID           int64         `gorm:"not null;index" json:"application_id"`
	SourceExpertID          int64         `gorm:"not null" json:"source_expert_id"`
	PublisherOrganizationID int64         `gorm:"not null;index" json:"publisher_organization_id"`
	PublisherUserID         int64         `gorm:"not null" json:"publisher_user_id"`
	Version                 int           `gorm:"not null" json:"version"`
	Status                  ReleaseStatus `gorm:"size:32;not null;index" json:"status"`

	Name        string         `gorm:"size:255;not null" json:"name"`
	Summary     string         `gorm:"type:text;not null" json:"summary"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Category    string         `gorm:"size:100;not null" json:"category"`
	Icon        string         `gorm:"size:100;not null" json:"icon"`
	Tags        pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"tags"`
	Outcomes    pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"outcomes"`
	Featured    bool           `gorm:"not null;default:false" json:"featured"`

	ExpertSnapshot     json.RawMessage `gorm:"type:jsonb;not null" json:"expert_snapshot"`
	WorkerSpecSnapshot json.RawMessage `gorm:"type:jsonb;not null" json:"worker_spec_snapshot"`
	SkillDependencies  json.RawMessage `gorm:"type:jsonb;not null" json:"skill_dependencies"`

	ReviewerUserID  *int64  `json:"reviewer_user_id,omitempty"`
	RejectionReason *string `gorm:"type:text" json:"rejection_reason,omitempty"`

	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	RejectedAt  *time.Time `json:"rejected_at,omitempty"`
	WithdrawnAt *time.Time `json:"withdrawn_at,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

func (Release) TableName() string {
	return "expert_market_releases"
}

func (release Release) Validate() error {
	if release.Version <= 0 {
		return ErrInvalidVersion
	}
	if !release.Status.Valid() {
		return ErrInvalidStatus
	}
	if err := validateVersionedObject(release.ExpertSnapshot); err != nil {
		return fmt.Errorf("%w: expert_snapshot", ErrInvalidSnapshot)
	}
	if err := validateVersionedObject(release.WorkerSpecSnapshot); err != nil {
		return fmt.Errorf("%w: worker_spec_snapshot", ErrInvalidSnapshot)
	}
	if err := validateJSONArray(release.SkillDependencies); err != nil {
		return fmt.Errorf("%w: skill_dependencies", ErrInvalidSnapshot)
	}
	return nil
}

func validateVersionedObject(raw json.RawMessage) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil || value == nil {
		return ErrInvalidSnapshot
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ErrInvalidSnapshot
	}
	version, ok := value["version"].(json.Number)
	if !ok {
		return ErrInvalidSnapshot
	}
	parsed, err := version.Int64()
	if err != nil || parsed <= 0 {
		return ErrInvalidSnapshot
	}
	return nil
}

func validateJSONArray(raw json.RawMessage) error {
	var value []json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return ErrInvalidSnapshot
	}
	return nil
}
