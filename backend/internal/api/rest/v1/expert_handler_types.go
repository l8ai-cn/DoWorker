package v1

import expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"

type createExpertRequest struct {
	Name            string                    `json:"name"`
	Slug            string                    `json:"slug"`
	Description     *string                   `json:"description"`
	AgentSlug       string                    `json:"agent_slug"`
	RunnerID        *int64                    `json:"runner_id"`
	RepositoryID    *int64                    `json:"repository_id"`
	BranchName      *string                   `json:"branch_name"`
	Prompt          *string                   `json:"prompt"`
	InteractionMode string                    `json:"interaction_mode"`
	Perpetual       bool                      `json:"perpetual"`
	UsedEnvBundles  []string                  `json:"used_env_bundles"`
	SkillSlugs      []string                  `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
	ConfigOverrides map[string]interface{}    `json:"config_overrides"`
	AgentfileLayer  *string                   `json:"agentfile_layer"`
}

type updateExpertRequest struct {
	Name            *string                   `json:"name"`
	Description     *string                   `json:"description"`
	AgentSlug       *string                   `json:"agent_slug"`
	RunnerID        *int64                    `json:"runner_id"`
	RepositoryID    *int64                    `json:"repository_id"`
	BranchName      *string                   `json:"branch_name"`
	Prompt          *string                   `json:"prompt"`
	InteractionMode *string                 `json:"interaction_mode"`
	Perpetual       *bool                     `json:"perpetual"`
	UsedEnvBundles  []string                  `json:"used_env_bundles"`
	SkillSlugs      []string                  `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
	ConfigOverrides map[string]interface{}    `json:"config_overrides"`
	AgentfileLayer  *string                   `json:"agentfile_layer"`
}

type publishExpertRequest struct {
	Name            string                    `json:"name"`
	Slug            string                    `json:"slug"`
	Description     *string                   `json:"description"`
	AgentfileLayer  *string                   `json:"agentfile_layer"`
	UsedEnvBundles  []string                  `json:"used_env_bundles"`
	SkillSlugs      []string                  `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
}

type runExpertRequest struct {
	Alias          *string `json:"alias"`
	PromptOverride *string `json:"prompt_override"`
	RunnerID       *int64  `json:"runner_id"`
	Cols           int32   `json:"cols"`
	Rows           int32   `json:"rows"`
}
