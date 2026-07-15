package v1

import expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"

type createExpertRequest struct {
	Name            string                     `json:"name"`
	Slug            string                     `json:"slug"`
	Description     *string                    `json:"description"`
	AgentSlug       string                     `json:"agent_slug"`
	RunnerID        *int64                     `json:"runner_id"`
	RepositoryID    *int64                     `json:"repository_id"`
	BranchName      *string                    `json:"branch_name"`
	Prompt          *string                    `json:"prompt"`
	InteractionMode string                     `json:"interaction_mode"`
	AutomationLevel string                     `json:"automation_level"`
	Perpetual       bool                       `json:"perpetual"`
	UsedEnvBundles  []string                   `json:"used_env_bundles"`
	SkillSlugs      []string                   `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
	ConfigOverrides map[string]interface{}     `json:"config_overrides"`
	AgentfileLayer  *string                    `json:"agentfile_layer"`
	Avatar          *avatarInput               `json:"avatar"`
	ExpertType      *string                    `json:"expert_type"`
}

// avatarInput is a base64-encoded avatar upload. content_base64 is decoded,
// size-capped, and MIME-sniffed in the handler before it reaches the service;
// the client-supplied filename is advisory only (the platform derives the
// stored path assets/avatar.<ext>).
type avatarInput struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
}

type updateExpertRequest struct {
	Name            *string                    `json:"name"`
	Description     *string                    `json:"description"`
	AgentSlug       *string                    `json:"agent_slug"`
	RunnerID        *int64                     `json:"runner_id"`
	RepositoryID    *int64                     `json:"repository_id"`
	BranchName      *string                    `json:"branch_name"`
	Prompt          *string                    `json:"prompt"`
	InteractionMode *string                    `json:"interaction_mode"`
	AutomationLevel *string                    `json:"automation_level"`
	Perpetual       *bool                      `json:"perpetual"`
	UsedEnvBundles  []string                   `json:"used_env_bundles"`
	SkillSlugs      []string                   `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
	ConfigOverrides map[string]interface{}     `json:"config_overrides"`
	AgentfileLayer  *string                    `json:"agentfile_layer"`
	Avatar          *avatarInput               `json:"avatar"`
	ExpertType      *string                    `json:"expert_type"`
}

type publishExpertRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

type runExpertRequest struct {
	Alias          *string `json:"alias"`
	PromptOverride *string `json:"prompt_override"`
	Cols           int32   `json:"cols"`
	Rows           int32   `json:"rows"`
}

type installMarketApplicationRequest struct {
	ModelResourceID int64 `json:"model_resource_id" binding:"required,gt=0"`
}
