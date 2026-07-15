package agent

import (
	"context"
	"fmt"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (b *ConfigBuilder) buildPinnedWorkerSkillResources(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) ([]*runnerv1.ResourceToDownload, error) {
	provider, ok := b.extensionProvider.(WorkerSkillProvider)
	if !ok {
		return nil, fmt.Errorf("exact worker skill resolver is unavailable")
	}
	skills, err := provider.GetWorkerSkillsByPackages(
		ctx,
		req.RequiredSkillPackages,
		agentSlug,
	)
	if err != nil {
		return nil, fmt.Errorf("load pinned worker skills: %w", err)
	}
	if err := validatePinnedWorkerSkills(req.RequiredSkillPackages, skills); err != nil {
		return nil, err
	}
	return skillResources(agentSlug, skills, true)
}

func validatePinnedWorkerSkills(
	packages []specdomain.SkillPackageBinding,
	skills []*extensionservice.ResolvedSkill,
) error {
	expected := make(map[int64]specdomain.SkillPackageBinding, len(packages))
	for _, pkg := range packages {
		expected[pkg.SkillID] = pkg
	}
	for _, skill := range skills {
		if skill == nil {
			return fmt.Errorf("pinned worker skill resolution returned nil")
		}
		pkg, exists := expected[skill.CatalogSkillID]
		if !exists || skill.Slug != pkg.Slug ||
			skill.ContentSha != pkg.ContentSHA ||
			skill.PackageSize != pkg.PackageSize {
			return fmt.Errorf(
				"pinned worker skill %d does not match snapshot",
				skill.CatalogSkillID,
			)
		}
		delete(expected, skill.CatalogSkillID)
	}
	if len(expected) != 0 {
		return fmt.Errorf("pinned worker skill resolution is incomplete")
	}
	return nil
}
