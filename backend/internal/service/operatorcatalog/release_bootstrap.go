package operatorcatalog

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	expertsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/expert"
)

func (bootstrapper *Bootstrapper) ensurePublished(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	source *expertdom.Expert,
) (bool, error) {
	published, err := bootstrapper.experts.GetPublishedMarketApplication(
		ctx,
		definition.Slug,
	)
	if err == nil {
		if !publishedMatches(published, definition, source.ID) {
			return false, ErrCatalogConflict
		}
		return false, nil
	}
	if !errors.Is(err, expertsvc.ErrMarketApplicationNotFound) {
		return false, err
	}
	submission, err := bootstrapper.experts.SubmitMarketApplication(
		ctx,
		expertsvc.SubmitMarketApplicationRequest{
			OrganizationID:  request.OrganizationID,
			UserID:          request.PublisherUserID,
			SourceExpertID:  source.ID,
			Slug:            definition.Slug,
			Summary:         definition.Summary,
			Description:     definition.Description,
			Category:        definition.Category,
			Icon:            definition.Icon,
			Tags:            definition.Tags,
			Outcomes:        definition.Outcomes,
			Featured:        true,
			IsOperatorOwned: true,
		},
	)
	if err == nil {
		_, err = bootstrapper.experts.ApproveMarketRelease(
			ctx,
			expertsvc.ReviewMarketReleaseRequest{
				ReviewerUserID: request.ReviewerUserID,
				ReleaseID:      submission.Release.ID,
			},
		)
		return err == nil, err
	}
	if !errors.Is(err, expertmarket.ErrPendingReleaseExists) {
		return false, err
	}
	release, err := bootstrapper.pendingRelease(
		ctx,
		request.OrganizationID,
		source.ID,
	)
	if err != nil {
		return false, err
	}
	_, err = bootstrapper.experts.ApproveMarketRelease(
		ctx,
		expertsvc.ReviewMarketReleaseRequest{
			ReviewerUserID: request.ReviewerUserID,
			ReleaseID:      release.ID,
		},
	)
	return err == nil, err
}

func (bootstrapper *Bootstrapper) pendingRelease(
	ctx context.Context,
	organizationID, sourceExpertID int64,
) (*expertmarket.Release, error) {
	releases, _, err := bootstrapper.experts.ListPublisherMarketReleases(
		ctx,
		organizationID,
		100,
		0,
	)
	if err != nil {
		return nil, err
	}
	for index := range releases {
		if releases[index].SourceExpertID == sourceExpertID &&
			releases[index].Status == expertmarket.ReleaseStatusPendingReview {
			return &releases[index], nil
		}
	}
	return nil, ErrCatalogConflict
}

func expertMatches(
	expert *expertdom.Expert,
	definition ExpertDefinition,
) bool {
	return expert != nil &&
		expert.Name == definition.Name &&
		optionalString(expert.Description) == definition.Description &&
		expert.AgentSlug == "video-studio" &&
		optionalString(expert.Prompt) == definition.Prompt &&
		expert.InteractionMode == expertdom.InteractionModePTY &&
		expert.AutomationLevel == expertdom.AutomationLevelAutoEdit &&
		expert.WorkerSpecSnapshotID != nil &&
		slices.Equal([]string(expert.SkillSlugs), definition.SkillSlugs)
}

func publishedMatches(
	published *expertsvc.PublishedMarketApplication,
	definition ExpertDefinition,
	sourceExpertID int64,
) bool {
	if published == nil {
		return false
	}
	release := published.Release
	return published.Application.IsOperatorOwned &&
		release.SourceExpertID == sourceExpertID &&
		release.Name == definition.Name &&
		release.Summary == strings.TrimSpace(definition.Summary) &&
		release.Description == strings.TrimSpace(definition.Description) &&
		release.Category == definition.Category &&
		release.Icon == definition.Icon &&
		sortedEqual(release.Tags, definition.Tags) &&
		sortedEqual(release.Outcomes, definition.Outcomes)
}

func sortedEqual(left, right []string) bool {
	left = append([]string{}, left...)
	right = append([]string{}, right...)
	sort.Strings(left)
	sort.Strings(right)
	return slices.Equal(left, right)
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
