package infra

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestExpertMarketConcurrentFirstSubmissionKeepsOneSourceApplication(
	t *testing.T,
) {
	db := openExpertMarketPostgresTestDB(t)
	repo := NewExpertMarketRepository(db)
	ctx := context.Background()
	start := make(chan struct{})
	errs := make(chan error, 2)
	var workers sync.WaitGroup

	for _, slug := range []string{"video-first", "video-second"} {
		workers.Add(1)
		go func(applicationSlug string) {
			defer workers.Done()
			<-start
			application := expertmarket.Application{
				Slug:                    slugkit.Slug(applicationSlug),
				PublisherOrganizationID: 1,
				SourceExpertID:          9001,
				PublisherUserID:         1,
			}
			release := postgresTestRelease(0, 9001, 1)
			release.Status = expertmarket.ReleaseStatusPendingReview
			errs <- repo.CreateSubmission(ctx, &application, &release)
		}(slug)
	}

	close(start)
	workers.Wait()
	close(errs)
	var succeeded, conflicted int
	for err := range errs {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, expertmarket.ErrConflict):
			conflicted++
		default:
			require.NoError(t, err)
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, conflicted)

	application, err := repo.GetApplicationBySourceExpert(ctx, 1, 9001)
	require.NoError(t, err)
	require.Contains(t, []string{"video-first", "video-second"}, application.Slug.String())
	_, total, err := repo.ListApplications(
		ctx,
		expertmarket.ApplicationListFilter{
			PublisherOrganizationID: int64PointerForInfra(1),
			Limit:                   10,
		},
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
}

func int64PointerForInfra(value int64) *int64 {
	return &value
}
