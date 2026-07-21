package expert

import (
	"encoding/json"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
)

type CreateExpertRequest struct {
	OrganizationID            int64
	UserID                    int64
	Name                      string
	Slug                      string
	Description               *string
	AgentSlug                 string
	RunnerID                  *int64
	RepositoryID              *int64
	BranchName                *string
	Prompt                    *string
	InteractionMode           string
	AutomationLevel           string
	Perpetual                 bool
	UsedEnvBundles            []string
	SkillSlugs                []string
	KnowledgeMounts           []expertdom.KnowledgeMount
	ConfigOverrides           map[string]interface{}
	AgentfileLayer            *string
	SourcePodKey              *string
	WorkerSpecSnapshotID      *int64
	SourceMarketApplicationID *int64
	SourceMarketReleaseID     *int64
	Metadata                  json.RawMessage
	Avatar                    *AvatarInput
	ExpertType                *string
}

type UpdateExpertRequest struct {
	OrganizationID  int64
	ExpertID        int64
	Name            *string
	Description     *string
	AgentSlug       *string
	RunnerID        *int64
	RepositoryID    *int64
	BranchName      *string
	Prompt          *string
	InteractionMode *string
	AutomationLevel *string
	Perpetual       *bool
	UsedEnvBundles  []string
	SkillSlugs      []string
	KnowledgeMounts []expertdom.KnowledgeMount
	ConfigOverrides map[string]interface{}
	AgentfileLayer  *string
	Avatar          *AvatarInput
	ExpertType      *string
}
