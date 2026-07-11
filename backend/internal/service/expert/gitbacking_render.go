package expert

import (
	"encoding/json"
	"fmt"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

type expertConfig struct {
	Schema          int                        `json:"schema"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description,omitempty"`
	Avatar          string                     `json:"avatar,omitempty"`
	ExpertType      string                     `json:"expertType,omitempty"`
	AgentSlug       string                     `json:"agentSlug"`
	InteractionMode string                     `json:"interactionMode"`
	AutomationLevel string                     `json:"automationLevel,omitempty"`
	Perpetual       bool                       `json:"perpetual"`
	SkillSlugs      []string                   `json:"skillSlugs,omitempty"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledgeMounts,omitempty"`
	UsedEnvBundles  []string                   `json:"usedEnvBundles,omitempty"`
	ConfigOverrides map[string]any             `json:"configOverrides,omitempty"`
	Repository      *expertConfigRepository    `json:"repository,omitempty"`
}

type expertConfigRepository struct {
	RepositoryID *int64 `json:"repositoryId,omitempty"`
	Branch       string `json:"branch,omitempty"`
}

type AvatarInput struct {
	Data []byte
	Ext  string
}

func (a *AvatarInput) repoPath() string {
	ext := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(a.Ext)), ".")
	if ext == "" {
		ext = "png"
	}
	return "assets/avatar." + ext
}

const expertConfigSchema = 1

func (s *Service) buildExpertConfig(
	expert *expertdom.Expert,
	avatarPath string,
) expertConfig {
	config := expertConfig{
		Schema:          expertConfigSchema,
		Name:            expert.Name,
		AgentSlug:       expert.AgentSlug,
		InteractionMode: expert.InteractionMode,
		AutomationLevel: expert.AutomationLevel,
		Perpetual:       expert.Perpetual,
		SkillSlugs:      []string(expert.SkillSlugs),
		UsedEnvBundles:  []string(expert.UsedEnvBundles),
		KnowledgeMounts: expertdom.ParseKnowledgeMounts(expert.KnowledgeMounts),
	}
	if expert.Description != nil {
		config.Description = *expert.Description
	}
	metadata := parseExpertMetadata(expert.Metadata)
	config.ExpertType = metadata.ExpertType
	config.Avatar = metadata.Avatar
	if avatarPath != "" {
		config.Avatar = avatarPath
	}
	if len(expert.ConfigOverrides) > 0 {
		var overrides map[string]any
		if err := json.Unmarshal(
			expert.ConfigOverrides,
			&overrides,
		); err == nil && len(overrides) > 0 {
			config.ConfigOverrides = overrides
		}
	}
	if expert.RepositoryID != nil {
		branch := ""
		if expert.BranchName != nil {
			branch = *expert.BranchName
		}
		config.Repository = &expertConfigRepository{
			RepositoryID: expert.RepositoryID,
			Branch:       branch,
		}
	}
	return config
}

func (s *Service) renderExpertFiles(
	expert *expertdom.Expert,
	layer string,
	avatar *AvatarInput,
	includeReadme bool,
) ([]gitops.FileChange, error) {
	avatarPath := ""
	if avatar != nil && len(avatar.Data) > 0 {
		avatarPath = avatar.repoPath()
	}
	configJSON, err := json.MarshalIndent(
		s.buildExpertConfig(expert, avatarPath),
		"",
		"  ",
	)
	if err != nil {
		return nil, fmt.Errorf("expert: render expert.json: %w", err)
	}
	changes := []gitops.FileChange{
		{Path: "agent.md", Content: []byte(layer)},
		{Path: "expert.json", Content: append(configJSON, '\n')},
	}
	if includeReadme {
		changes = append(changes, gitops.FileChange{
			Path:    "README.md",
			Content: []byte(renderExpertReadme(expert)),
		})
	}
	if avatarPath != "" {
		changes = append(changes, gitops.FileChange{
			Path:    avatarPath,
			Content: avatar.Data,
		})
	}
	return changes, nil
}

func renderExpertReadme(expert *expertdom.Expert) string {
	description := ""
	if expert.Description != nil {
		description = strings.TrimSpace(*expert.Description)
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "# %s\n\n", expert.Name)
	if description != "" {
		fmt.Fprintf(&builder, "%s\n\n", description)
	}
	builder.WriteString(
		"This repository stores expert metadata. Executable runtime state is " +
			"bound to the immutable WorkerSpec snapshot created during publishing; " +
			"`agent.md` does not override that snapshot.\n",
	)
	return builder.String()
}
