package adminconnect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/database"
	adminv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/admin/v1"
)

func expertMarketStatus(raw string) (expertmarket.ReleaseStatus, error) {
	switch strings.TrimSpace(raw) {
	case "", "pending":
		return expertmarket.ReleaseStatusPendingReview, nil
	case "published":
		return expertmarket.ReleaseStatusPublished, nil
	case "rejected":
		return expertmarket.ReleaseStatusRejected, nil
	case "withdrawn":
		return expertmarket.ReleaseStatusWithdrawn, nil
	default:
		return "", invalidExpertMarketArgument("unsupported status")
	}
}

func expertMarketPagination(limitValue, offsetValue int32) (int, int, error) {
	if limitValue < 0 || offsetValue < 0 {
		return 0, 0, invalidExpertMarketArgument(
			"limit and offset must not be negative",
		)
	}
	limit := int(limitValue)
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		return 0, 0, invalidExpertMarketArgument(
			"limit must not exceed 100",
		)
	}
	return limit, int(offsetValue), nil
}

func invalidExpertMarketArgument(message string) error {
	return connect.NewError(
		connect.CodeInvalidArgument,
		errors.New(message),
	)
}

func toProtoExpertMarketRelease(
	release *expertmarket.Release,
) *adminv1.ExpertMarketRelease {
	if release == nil {
		return nil
	}
	return &adminv1.ExpertMarketRelease{
		Id:                      release.ID,
		ApplicationId:           release.ApplicationID,
		ApplicationSlug:         release.ApplicationSlug,
		SourceExpertId:          release.SourceExpertID,
		PublisherOrganizationId: release.PublisherOrganizationID,
		PublisherUserId:         release.PublisherUserID,
		Version:                 int32(release.Version),
		Status:                  expertMarketStatusLabel(release.Status),
		Name:                    release.Name,
		Summary:                 release.Summary,
		Description:             release.Description,
		Category:                release.Category,
		Icon:                    release.Icon,
		Tags:                    append([]string{}, release.Tags...),
		Outcomes:                append([]string{}, release.Outcomes...),
		Featured:                release.Featured,
		ExpertSnapshotJson:      string(release.ExpertSnapshot),
		WorkerSpecSnapshotJson:  string(release.WorkerSpecSnapshot),
		SkillDependenciesJson:   string(release.SkillDependencies),
		ReviewerUserId:          release.ReviewerUserID,
		RejectionReason:         release.RejectionReason,
		SubmittedAt:             protoTime(release.SubmittedAt),
		ReviewedAt:              protoTime(release.ReviewedAt),
		PublishedAt:             protoTime(release.PublishedAt),
		RejectedAt:              protoTime(release.RejectedAt),
		WithdrawnAt:             protoTime(release.WithdrawnAt),
		CreatedAt:               release.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func loadExpertMarketApplicationSlugs(
	ctx context.Context,
	db database.DB,
	releases []*expertmarket.Release,
) error {
	if len(releases) == 0 {
		return nil
	}
	applicationIDs := make([]int64, 0, len(releases))
	seen := make(map[int64]struct{}, len(releases))
	for _, release := range releases {
		if _, exists := seen[release.ApplicationID]; exists {
			continue
		}
		seen[release.ApplicationID] = struct{}{}
		applicationIDs = append(applicationIDs, release.ApplicationID)
	}
	var applications []expertmarket.Application
	if err := db.WithContext(ctx).
		Where("id IN ?", applicationIDs).
		Find(&applications); err != nil {
		return err
	}
	slugs := make(map[int64]string, len(applications))
	for _, application := range applications {
		slugs[application.ID] = application.Slug.String()
	}
	for index := range releases {
		slug, exists := slugs[releases[index].ApplicationID]
		if !exists {
			return fmt.Errorf(
				"expert market application %d is missing",
				releases[index].ApplicationID,
			)
		}
		releases[index].ApplicationSlug = slug
	}
	return nil
}

func expertMarketStatusLabel(status expertmarket.ReleaseStatus) string {
	if status == expertmarket.ReleaseStatusPendingReview {
		return "pending"
	}
	return string(status)
}

func protoTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339Nano)
	return &formatted
}
