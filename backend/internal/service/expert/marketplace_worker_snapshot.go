package expert

import (
	"context"
	"errors"
	"sort"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/lib/pq"
)

func (s *Service) prepareMarketplaceWorkerSnapshot(
	ctx context.Context,
	request MarketplaceInstallationRequest,
	snapshot marketplaceRuntimeSnapshot,
) (int64, error) {
	if s.workerSpecWriter == nil || s.marketWorkerSpecs == nil ||
		s.marketSkills == nil {
		return 0, ErrMarketUnavailable
	}
	skills, err := s.marketSkills.ListActivePlatformBySlugs(
		ctx,
		snapshot.Expert.SkillSlugs,
	)
	if err != nil {
		return 0, err
	}
	if err := validateMarketplaceSkills(snapshot.Expert.SkillSlugs, skills); err != nil {
		return 0, err
	}
	source := snapshot.WorkerSpec
	source.Workspace.SkillIDs = marketSkillIDs(skills)
	source, err = marketWorkerSpec(source, skills)
	if err != nil {
		return 0, errors.Join(ErrMarketplaceInstallationInvalid, err)
	}
	if err := validateExpertMatchesWorkerSpec(
		marketplaceSnapshotExpert(snapshot.Expert),
		source,
		skills,
	); err != nil {
		return 0, errors.Join(ErrMarketplaceInstallationInvalid, err)
	}
	resolved, err := s.marketWorkerSpecs.PrepareMarketSnapshot(
		ctx,
		specservice.Scope{
			OrgID:  request.TargetOrganizationID,
			UserID: request.ActorUserID,
		},
		source,
		request.ModelResourceID,
	)
	if err != nil {
		return 0, err
	}
	created, err := s.workerSpecWriter.Create(ctx, resolved)
	if err != nil {
		return 0, err
	}
	if created.ID <= 0 ||
		created.OrganizationID != request.TargetOrganizationID {
		return 0, ErrMarketSnapshotInvalid
	}
	return created.ID, nil
}

func validateMarketplaceSkills(
	requiredSlugs []string,
	skills []skilldom.Skill,
) error {
	required := normalizeMarketStrings(requiredSlugs)
	actual := marketSkillSlugs(skills)
	if len(required) != len(actual) {
		return &MarketDependencyError{Missing: missingMarketSkills(required, actual)}
	}
	for index := range required {
		if required[index] != actual[index] {
			return &MarketDependencyError{Missing: missingMarketSkills(required, actual)}
		}
	}
	return nil
}

func missingMarketSkills(required, actual []string) []string {
	found := make(map[string]struct{}, len(actual))
	for _, slug := range actual {
		found[slug] = struct{}{}
	}
	missing := make([]string, 0)
	for _, slug := range required {
		if _, ok := found[slug]; !ok {
			missing = append(missing, slug)
		}
	}
	return missing
}

func marketSkillIDs(skills []skilldom.Skill) []int64 {
	ids := make([]int64, 0, len(skills))
	for _, skill := range skills {
		ids = append(ids, skill.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func marketplaceSnapshotExpert(snapshot marketExpertSnapshot) *expertdom.Expert {
	return &expertdom.Expert{
		Slug:            snapshot.Slug,
		Name:            snapshot.Name,
		Description:     snapshot.Description,
		AgentSlug:       snapshot.AgentSlug,
		Prompt:          snapshot.Prompt,
		InteractionMode: snapshot.InteractionMode,
		AutomationLevel: snapshot.AutomationLevel,
		Perpetual:       snapshot.Perpetual,
		UsedEnvBundles:  pq.StringArray(snapshot.UsedEnvBundles),
		SkillSlugs:      pq.StringArray(snapshot.SkillSlugs),
		KnowledgeMounts: encodeKnowledgeMounts(snapshot.KnowledgeMounts),
		ConfigOverrides: encodeConfigOverrides(snapshot.ConfigOverrides),
		Metadata:        snapshot.Metadata,
	}
}
