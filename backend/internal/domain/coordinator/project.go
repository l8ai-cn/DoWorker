package coordinator

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("coordinator: not found")

const (
	PlatformTypeCNB    = "cnb"
	PlatformTypeLinear = "linear"
)

const (
	SourceTypeIssues = "issues"
	SourceTypePulls  = "pulls"
)

// ClaimPolicy decides which discovered external tasks become candidates. It is
// the subset of auto-harness config.ClaimPolicySpec that AgentsMesh evaluates
// in-process; persisted as JSONB on the project row.
type ClaimPolicy struct {
	Labels         []string `json:"labels,omitempty"`
	States         []string `json:"states,omitempty"`
	Priorities     []string `json:"priorities,omitempty"`
	TaskTypes      []string `json:"taskTypes,omitempty"`
	UnassignedOnly bool     `json:"unassignedOnly,omitempty"`
	TitleKeywords  []string `json:"titleKeywords,omitempty"`
	BodyKeywords   []string `json:"bodyKeywords,omitempty"`
	MaxActiveTasks int      `json:"maxActiveTasks,omitempty"`
}

// Project is an org-scoped coordinator config bound to a single repository. It
// maps onto auto-harness ProjectSpec but reuses AgentsMesh repository + ticket
// primitives instead of carrying its own platform credentials.
type Project struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`
	RepositoryID   int64 `gorm:"not null;index" json:"repository_id"`

	Slug string `gorm:"size:100;not null;uniqueIndex:idx_coordinator_projects_org_slug" json:"slug"`
	Name string `gorm:"size:255;not null" json:"name"`

	PlatformType string `gorm:"size:32;not null;default:'cnb'" json:"platform_type"`
	SourceType   string `gorm:"size:32;not null;default:'issues'" json:"source_type"`

	LabelFilter pq.StringArray  `gorm:"type:text[];not null;default:'{}'" json:"label_filter"`
	ClaimPolicy json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"claim_policy"`

	AgentSlug           string `gorm:"size:100;not null;default:'do-agent'" json:"agent_slug"`
	ScanIntervalSeconds int    `gorm:"not null;default:300" json:"scan_interval_seconds"`
	MaxConcurrent       int    `gorm:"not null;default:1" json:"max_concurrent"`
	Enabled             bool   `gorm:"not null;default:true;index" json:"enabled"`

	CreatedByID int64 `gorm:"not null" json:"created_by_id"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Project) TableName() string { return "coordinator_projects" }

// DecodeClaimPolicy parses the persisted JSONB. The label_filter column is
// merged in so callers can express the common "match these labels" case
// without hand-writing the policy JSON.
func (p *Project) DecodeClaimPolicy() ClaimPolicy {
	var policy ClaimPolicy
	if len(p.ClaimPolicy) > 0 {
		_ = json.Unmarshal(p.ClaimPolicy, &policy)
	}
	for _, label := range p.LabelFilter {
		if !containsFold(policy.Labels, label) {
			policy.Labels = append(policy.Labels, label)
		}
	}
	return policy
}

func (p *Project) ScanInterval() time.Duration {
	if p.ScanIntervalSeconds <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(p.ScanIntervalSeconds) * time.Second
}
