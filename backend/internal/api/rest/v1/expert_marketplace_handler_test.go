package v1

import (
	"errors"
	"net/http"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/stretchr/testify/require"
)

func TestSubmitMarketApplicationUsesTenantOwnedExpert(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	fixture.experts.rows[7] = &expertdom.Expert{
		ID:             7,
		OrganizationID: 41,
		Slug:           "video-director",
		Name:           "Video Director",
	}

	recorder := fixture.perform(
		http.MethodPost,
		"/experts/video-director/market-submissions",
		`{
			"slug":"video-director",
			"summary":"Plans short-form videos",
			"description":"Creates a production-ready plan",
			"category":"video",
			"icon":"film",
			"tags":["short-video"],
			"outcomes":["shooting-plan"]
		}`,
	)

	require.Equal(t, http.StatusConflict, recorder.Code)
	require.Equal(t, int64(41), fixture.experts.lastOrganizationID)
	require.Equal(t, int64(7), fixture.experts.lastExpertID)
}

func TestSubmitMarketApplicationValidatesRequest(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	recorder := fixture.perform(
		http.MethodPost,
		"/experts/video-director/market-submissions",
		`{"slug":"Bad Slug","summary":"x","category":"video","icon":"unknown"}`,
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Empty(t, fixture.experts.lastSlug)
}

func TestListMarketSubmissionsUsesOrganizationPagination(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	fixture.market.applications[3] = &expertmarket.Application{
		ID: 3, Slug: "video-director", PublisherOrganizationID: 41,
	}
	fixture.market.releases[1] = &expertmarket.Release{
		ID: 1, ApplicationID: 3, PublisherOrganizationID: 41,
		Status: expertmarket.ReleaseStatusPendingReview,
	}
	fixture.market.releases[2] = &expertmarket.Release{
		ID: 2, PublisherOrganizationID: 99,
		Status: expertmarket.ReleaseStatusPendingReview,
	}

	recorder := fixture.perform(
		http.MethodGet,
		"/marketplace/submissions?limit=10&offset=0",
		"",
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{
		"releases":[{
			"id":1,"application_id":3,"application_slug":"video-director",
			"source_expert_id":0,
			"publisher_organization_id":41,"publisher_user_id":0,
			"version":0,"status":"pending_review","name":"","summary":"",
			"description":"","category":"","icon":"","tags":null,"outcomes":null,
			"featured":false,"expert_snapshot":null,
			"worker_spec_snapshot":null,"skill_dependencies":null,
			"created_at":"0001-01-01T00:00:00Z"
		}],
		"total":1,"limit":10,"offset":0
	}`, recorder.Body.String())
}

func TestListMarketSubmissionsRejectsInvalidPagination(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	recorder := fixture.perform(
		http.MethodGet,
		"/marketplace/submissions?limit=-1&offset=nope",
		"",
	)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestWithdrawMarketReleaseEnforcesPublisherOwnership(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	fixture.market.applications[3] = &expertmarket.Application{
		ID: 3, PublisherOrganizationID: 99,
	}
	fixture.market.releases[8] = &expertmarket.Release{
		ID: 8, ApplicationID: 3,
		Status: expertmarket.ReleaseStatusPublished,
	}

	recorder := fixture.perform(
		http.MethodPost,
		"/marketplace/releases/8/withdraw",
		"",
	)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestWithdrawMarketReleaseRejectsInvalidReleaseID(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	recorder := fixture.perform(
		http.MethodPost,
		"/marketplace/releases/not-a-number/withdraw",
		"",
	)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestWithdrawMarketReleaseReturnsWithdrawnRelease(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	fixture.market.applications[3] = &expertmarket.Application{
		ID: 3, PublisherOrganizationID: 41,
	}
	fixture.market.releases[8] = &expertmarket.Release{
		ID: 8, ApplicationID: 3,
		Status: expertmarket.ReleaseStatusPublished,
	}

	recorder := fixture.perform(
		http.MethodPost,
		"/marketplace/releases/8/withdraw",
		"",
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"status":"withdrawn"`)
}

func TestMarketUpgradeAvailabilityUsesInstalledApplication(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	currentReleaseID := int64(10)
	latestReleaseID := int64(11)
	applicationID := int64(5)
	fixture.experts.rows[7] = &expertdom.Expert{
		ID: 7, OrganizationID: 41, Slug: "video-editor",
		SourceMarketApplicationID: &applicationID,
		SourceMarketReleaseID:     &currentReleaseID,
	}
	fixture.market.applications[5] = &expertmarket.Application{
		ID: 5, LatestPublishedReleaseID: &latestReleaseID,
	}

	recorder := fixture.perform(
		http.MethodGet,
		"/experts/video-editor/market-upgrade",
		"",
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"upgrade_available":true}`, recorder.Body.String())
}

func TestMarketUpgradeAvailabilityRejectsAuthoredExpert(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	fixture.experts.rows[7] = &expertdom.Expert{
		ID: 7, OrganizationID: 41, Slug: "authored-expert",
	}

	recorder := fixture.perform(
		http.MethodGet,
		"/experts/authored-expert/market-upgrade",
		"",
	)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestUpgradeMarketApplicationReturnsNoChange(t *testing.T) {
	fixture := newExpertMarketplaceAPIFixture()
	releaseID := int64(11)
	applicationID := int64(5)
	fixture.experts.rows[7] = &expertdom.Expert{
		ID: 7, OrganizationID: 41, Slug: "video-editor",
		SourceMarketApplicationID: &applicationID,
		SourceMarketReleaseID:     &releaseID,
	}
	fixture.market.applications[5] = &expertmarket.Application{
		ID: 5, LatestPublishedReleaseID: &releaseID,
	}
	fixture.market.releases[11] = &expertmarket.Release{
		ID: 11, ApplicationID: 5,
		Status: expertmarket.ReleaseStatusPublished,
	}

	recorder := fixture.perform(
		http.MethodPost,
		"/experts/video-editor/market-upgrade",
		"",
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"upgraded":false`)
}

func TestExpertMarketplaceErrorMapping(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{expertdom.ErrNotFound, http.StatusNotFound},
		{expertmarket.ErrNotFound, http.StatusNotFound},
		{expertsvc.ErrMarketApplicationNotFound, http.StatusNotFound},
		{expertsvc.ErrMarketUnavailable, http.StatusServiceUnavailable},
		{expertsvc.ErrMarketApplicationOwnership, http.StatusForbidden},
		{expertsvc.ErrMarketApplicationSlugMismatch, http.StatusConflict},
		{expertsvc.ErrMarketInvalidTransition, http.StatusConflict},
		{expertsvc.ErrMarketReleaseNotPublished, http.StatusConflict},
		{expertmarket.ErrPendingReleaseExists, http.StatusConflict},
		{expertsvc.ErrMarketSourceSnapshotRequired, http.StatusConflict},
		{errors.New("database offline"), http.StatusInternalServerError},
	}
	for _, test := range tests {
		fixture := newExpertMarketplaceAPIFixture()
		recorder := fixture.errorResponse(test.err)
		require.Equal(t, test.status, recorder.Code)
	}
}
