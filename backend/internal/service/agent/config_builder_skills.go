package agent

import (
	"context"
	"fmt"
	"log/slog"

	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (b *ConfigBuilder) buildSkillResources(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
	requestedSlugs []string,
) ([]*runnerv1.ResourceToDownload, error) {
	if len(req.RequiredSkillPackages) > 0 {
		return b.buildPinnedWorkerSkillResources(ctx, req, agentSlug)
	}
	if len(req.RequiredSkillIDs) > 0 {
		return b.buildWorkerSpecSkillResources(ctx, req, agentSlug)
	}
	if b.extensionProvider == nil || req.RepositoryID == nil {
		return nil, nil
	}
	skills, err := b.extensionProvider.GetEffectiveSkills(
		ctx,
		req.OrganizationID,
		req.UserID,
		*req.RepositoryID,
		agentSlug,
	)
	if err != nil {
		slog.WarnContext(
			ctx,
			"Failed to load skills for agentfile",
			"agent_slug",
			agentSlug,
			"error",
			err,
		)
		return nil, nil
	}
	return skillResources(agentSlug, filterResolvedSkillsBySlugs(skills, requestedSlugs), false)
}

func (b *ConfigBuilder) buildWorkerSpecSkillResources(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) ([]*runnerv1.ResourceToDownload, error) {
	provider, ok := b.extensionProvider.(WorkerSkillProvider)
	if !ok {
		return nil, fmt.Errorf("exact worker skill resolver is unavailable")
	}
	skills, err := provider.GetWorkerSkillsByIDs(
		ctx,
		req.OrganizationID,
		req.RequiredSkillIDs,
		agentSlug,
	)
	if err != nil {
		return nil, fmt.Errorf("load required worker skills: %w", err)
	}
	expected := make(map[int64]struct{}, len(req.RequiredSkillIDs))
	for _, id := range req.RequiredSkillIDs {
		expected[id] = struct{}{}
	}
	for _, skill := range skills {
		if skill == nil {
			return nil, fmt.Errorf("required worker skill resolution returned nil")
		}
		if _, exists := expected[skill.CatalogSkillID]; !exists {
			return nil, fmt.Errorf(
				"required worker skill resolution substituted id %d",
				skill.CatalogSkillID,
			)
		}
		delete(expected, skill.CatalogSkillID)
	}
	if len(expected) != 0 {
		return nil, fmt.Errorf("required worker skill resolution is incomplete")
	}
	return skillResources(agentSlug, skills, true)
}

func skillResources(
	agentSlug string,
	skills []*extensionservice.ResolvedSkill,
	strict bool,
) ([]*runnerv1.ResourceToDownload, error) {
	resources := make([]*runnerv1.ResourceToDownload, 0, len(skills))
	for _, skill := range skills {
		if skill == nil {
			continue
		}
		if skill.ContentSha == "" || skill.DownloadURL == "" || skill.Slug == "" {
			if strict {
				return nil, fmt.Errorf(
					"required worker skill %d has incomplete download metadata",
					skill.CatalogSkillID,
				)
			}
			continue
		}
		resources = append(resources, &runnerv1.ResourceToDownload{
			Sha:          skill.ContentSha,
			DownloadUrl:  skill.DownloadURL,
			TargetPath:   skillTargetPath(agentSlug, skill.Slug),
			ResourceType: "skill_package",
			SizeBytes:    skill.PackageSize,
		})
	}
	return resources, nil
}

func filterResolvedSkillsBySlugs(
	skills []*extensionservice.ResolvedSkill,
	requestedSlugs []string,
) []*extensionservice.ResolvedSkill {
	if len(requestedSlugs) == 0 {
		return skills
	}
	requested := make(map[string]struct{}, len(requestedSlugs))
	for _, slug := range requestedSlugs {
		requested[slug] = struct{}{}
	}
	for _, skill := range skills {
		if skill != nil {
			if _, exists := requested[skill.Slug]; exists {
				return filterSkillsBySlugSet(skills, requested)
			}
		}
	}
	return skills
}

func filterSkillsBySlugSet(
	skills []*extensionservice.ResolvedSkill,
	requested map[string]struct{},
) []*extensionservice.ResolvedSkill {
	filtered := make([]*extensionservice.ResolvedSkill, 0, len(skills))
	for _, skill := range skills {
		if skill != nil {
			if _, exists := requested[skill.Slug]; exists {
				filtered = append(filtered, skill)
			}
		}
	}
	return filtered
}

func skillTargetPath(agentSlug, skillSlug string) string {
	switch agentSlug {
	case "codex-cli", "codex", "pattern-designer", "video-studio":
		return "{{.sandbox.root_path}}/codex-home/skills/" + skillSlug
	case "claude-code", "claude":
		return "{{.sandbox.work_dir}}/.claude/skills/" + skillSlug
	case "do-agent", "seedance-expert":
		return "{{.sandbox.work_dir}}/.agent/skills/" + skillSlug
	default:
		return "{{.sandbox.root_path}}/skills/" + skillSlug
	}
}
