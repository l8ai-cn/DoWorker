package expert

import (
	"encoding/json"
	"errors"
	"fmt"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

var (
	ErrMarketApplicationNotFound     = errors.New("market application not found")
	ErrMarketUnavailable             = errors.New("expert marketplace is not configured")
	ErrMarketSourceSnapshotRequired  = errors.New("market source expert requires a workerspec snapshot")
	ErrMarketApplicationOwnership    = errors.New("market application belongs to another publisher")
	ErrMarketApplicationSlugMismatch = errors.New(
		"market application slug does not match the source expert release history",
	)
	ErrMarketInvalidTransition       = errors.New("expert market release transition is invalid")
	ErrMarketRejectionReasonRequired = errors.New("market rejection reason is required")
	ErrMarketReleaseNotPublished     = errors.New("expert market release is not published")
	ErrMarketSnapshotInvalid         = errors.New("expert market release snapshot is invalid")
)

type MarketDependencyError struct {
	Missing []string
}

func (err *MarketDependencyError) Error() string {
	return fmt.Sprintf(
		"market expert requires active platform skills: %v",
		err.Missing,
	)
}

type MarketSkillDependency struct {
	SkillID     int64  `json:"skill_id"`
	Slug        string `json:"slug"`
	Version     int    `json:"version"`
	ContentSHA  string `json:"content_sha"`
	StorageKey  string `json:"storage_key"`
	PackageSize int64  `json:"package_size"`
}

type MarketApplication struct {
	ID          int64    `json:"id"`
	ReleaseID   int64    `json:"release_id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Icon        string   `json:"icon"`
	AgentSlug   string   `json:"agent_slug"`
	SkillSlugs  []string `json:"skill_slugs"`
	Tags        []string `json:"tags"`
	Outcomes    []string `json:"outcomes"`
	Version     int      `json:"version"`
	Featured    bool     `json:"featured"`
}

type SubmitMarketApplicationRequest struct {
	OrganizationID  int64
	UserID          int64
	SourceExpertID  int64
	Slug            string
	Summary         string
	Description     string
	Category        string
	Icon            string
	Tags            []string
	Outcomes        []string
	Featured        bool
	IsOperatorOwned bool
}

type MarketSubmission struct {
	Application expertmarket.Application
	Release     expertmarket.Release
}

type ReviewMarketReleaseRequest struct {
	ReviewerUserID  int64
	ReleaseID       int64
	RejectionReason string
}

type WithdrawMarketReleaseRequest struct {
	PublisherOrganizationID int64
	ReleaseID               int64
}

type InstallMarketApplicationRequest struct {
	OrganizationID       int64
	UserID               int64
	ModelResourceID      int64
	ToolModelResourceIDs map[string]int64
	MarketSlug           string
}

type UpgradeMarketApplicationRequest struct {
	OrganizationID int64
	UserID         int64
	ExpertID       int64
}

type PublishedMarketApplication struct {
	Application expertmarket.Application
	Release     expertmarket.Release
}

type marketExpertSnapshot struct {
	Version         int                        `json:"version"`
	Slug            string                     `json:"slug"`
	Name            string                     `json:"name"`
	Description     *string                    `json:"description,omitempty"`
	AgentSlug       string                     `json:"agent_slug"`
	Prompt          *string                    `json:"prompt,omitempty"`
	InteractionMode string                     `json:"interaction_mode"`
	AutomationLevel string                     `json:"automation_level"`
	Perpetual       bool                       `json:"perpetual"`
	UsedEnvBundles  []string                   `json:"used_env_bundles"`
	SkillSlugs      []string                   `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
	ConfigOverrides map[string]any             `json:"config_overrides"`
	Metadata        json.RawMessage            `json:"metadata"`
}

type marketWorkerSpecSnapshot struct {
	Version int                `json:"version"`
	Spec    specdomain.Spec    `json:"spec"`
	Summary specdomain.Summary `json:"summary"`
}
