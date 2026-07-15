package expert

import (
	"fmt"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func marketWorkerSpec(
	source specdomain.Spec,
	skills []skilldom.Skill,
) (specdomain.Spec, error) {
	packages := source.Workspace.SkillPackages
	if len(packages) == 0 {
		packages = make([]specdomain.SkillPackageBinding, 0, len(skills))
		for _, skill := range skills {
			packages = append(packages, skillPackageBinding(skill))
		}
	} else if err := validateMarketPackageOwnership(packages, skills); err != nil {
		return specdomain.Spec{}, err
	}
	source.Workspace.SkillPackages = append(
		[]specdomain.SkillPackageBinding{},
		packages...,
	)
	return specdomain.NormalizeAndValidate(source)
}

func validateMarketPackageOwnership(
	packages []specdomain.SkillPackageBinding,
	skills []skilldom.Skill,
) error {
	skillsByID := make(map[int64]skilldom.Skill, len(skills))
	for _, skill := range skills {
		skillsByID[skill.ID] = skill
	}
	if len(packages) != len(skillsByID) {
		return fmt.Errorf("skill package bindings do not match platform skills")
	}
	for _, pkg := range packages {
		skill, exists := skillsByID[pkg.SkillID]
		if !exists || skill.Slug != pkg.Slug {
			return fmt.Errorf("skill package %d is not platform-owned", pkg.SkillID)
		}
	}
	return nil
}

func skillPackageBinding(skill skilldom.Skill) specdomain.SkillPackageBinding {
	return specdomain.SkillPackageBinding{
		SkillID:     skill.ID,
		Slug:        skill.Slug,
		Version:     skill.Version,
		ContentSHA:  skill.ContentSha,
		StorageKey:  skill.StorageKey,
		PackageSize: skill.PackageSize,
	}
}
