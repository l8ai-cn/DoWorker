package agentpod

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func artifactSkillPackages(
	document *workerdependency.Document,
) []specdomain.SkillPackageBinding {
	if document == nil || len(document.Skills) == 0 {
		return nil
	}
	packages := make([]specdomain.SkillPackageBinding, len(document.Skills))
	for index, skill := range document.Skills {
		packages[index] = specdomain.SkillPackageBinding{
			SkillID:     skill.Pin.DomainID,
			Slug:        skill.Slug.String(),
			Version:     skill.Version,
			ContentSHA:  skill.ContentDigest,
			StorageKey:  skill.StorageKey,
			PackageSize: skill.PackageSize,
		}
	}
	return packages
}
