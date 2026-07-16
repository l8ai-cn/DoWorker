package adminconnect

import (
	"context"
	"encoding/json"
	"testing"

	"connectrpc.com/connect"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/infra/database"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	adminv1 "github.com/anthropics/agentsmesh/proto/gen/go/admin/v1"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExpertMarketReviewListDefaultsToPending(t *testing.T) {
	srv, repo := newExpertMarketAdminTestServer(t, true)
	pending := seedExpertMarketRelease(t, repo, "pending-video", expertmarket.ReleaseStatusPendingReview)
	seedExpertMarketRelease(t, repo, "published-video", expertmarket.ReleaseStatusPublished)

	resp, err := srv.ListExpertMarketReleases(
		adminCtx(99),
		connect.NewRequest(&adminv1.ListExpertMarketReleasesRequest{}),
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Msg.Total)
	require.Len(t, resp.Msg.Items, 1)
	require.Equal(t, pending.ID, resp.Msg.Items[0].Id)
	require.Equal(t, "pending-video", resp.Msg.Items[0].ApplicationSlug)
	require.Equal(t, "pending", resp.Msg.Items[0].Status)
	require.JSONEq(t, `{"version":1,"slug":"pending-video"}`, resp.Msg.Items[0].ExpertSnapshotJson)
	require.JSONEq(t, `[{"slug":"render","version":1}]`, resp.Msg.Items[0].SkillDependenciesJson)
}

func TestExpertMarketReviewListReturnsPaginationMetadata(t *testing.T) {
	srv, repo := newExpertMarketAdminTestServer(t, true)
	seedExpertMarketRelease(t, repo, "first-video", expertmarket.ReleaseStatusPendingReview)
	seedExpertMarketRelease(t, repo, "second-video", expertmarket.ReleaseStatusPendingReview)

	resp, err := srv.ListExpertMarketReleases(
		adminCtx(99),
		connect.NewRequest(&adminv1.ListExpertMarketReleasesRequest{
			Limit:  int32Ptr(1),
			Offset: int32Ptr(1),
		}),
	)

	require.NoError(t, err)
	require.Equal(t, int64(2), resp.Msg.Total)
	require.Equal(t, int32(1), resp.Msg.Limit)
	require.Equal(t, int32(1), resp.Msg.Offset)
	require.Len(t, resp.Msg.Items, 1)
	require.NotEmpty(t, resp.Msg.Items[0].ApplicationSlug)
}

func TestExpertMarketReviewListSupportsPublishedStatus(t *testing.T) {
	srv, repo := newExpertMarketAdminTestServer(t, true)
	seedExpertMarketRelease(t, repo, "pending-video", expertmarket.ReleaseStatusPendingReview)
	published := seedExpertMarketRelease(t, repo, "published-video", expertmarket.ReleaseStatusPublished)

	resp, err := srv.ListExpertMarketReleases(
		adminCtx(99),
		connect.NewRequest(&adminv1.ListExpertMarketReleasesRequest{
			Status: stringPtr("published"),
		}),
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Msg.Total)
	require.Equal(t, published.ID, resp.Msg.Items[0].Id)
	require.Equal(t, "published", resp.Msg.Items[0].Status)
}

func TestExpertMarketReviewGetReturnsSnapshotFields(t *testing.T) {
	srv, repo := newExpertMarketAdminTestServer(t, true)
	release := seedExpertMarketRelease(t, repo, "detail-video", expertmarket.ReleaseStatusRejected)

	resp, err := srv.GetExpertMarketRelease(
		adminCtx(99),
		connect.NewRequest(&adminv1.GetExpertMarketReleaseRequest{ReleaseId: release.ID}),
	)

	require.NoError(t, err)
	require.Equal(t, release.Name, resp.Msg.Name)
	require.Equal(t, "detail-video", resp.Msg.ApplicationSlug)
	require.Equal(t, "rejected", resp.Msg.Status)
	require.Equal(t, "film", resp.Msg.Icon)
	require.Equal(t, []string{"video"}, resp.Msg.Tags)
	require.Equal(t, []string{"render"}, resp.Msg.Outcomes)
	require.JSONEq(t, `{"version":1,"spec":{"agent":"codex"}}`, resp.Msg.WorkerSpecSnapshotJson)
}

func TestExpertMarketReviewApprovePublishesPendingRelease(t *testing.T) {
	srv, repo := newExpertMarketAdminTestServer(t, true)
	release := seedExpertMarketRelease(t, repo, "approve-video", expertmarket.ReleaseStatusPendingReview)

	resp, err := srv.ApproveExpertMarketRelease(
		adminCtx(99),
		connect.NewRequest(&adminv1.ApproveExpertMarketReleaseRequest{ReleaseId: release.ID}),
	)

	require.NoError(t, err)
	require.Equal(t, "published", resp.Msg.Status)
	require.Equal(t, int64(99), resp.Msg.GetReviewerUserId())
}

func TestExpertMarketReviewRejectRequiresReason(t *testing.T) {
	srv, repo := newExpertMarketAdminTestServer(t, true)
	release := seedExpertMarketRelease(t, repo, "reject-video", expertmarket.ReleaseStatusPendingReview)

	_, err := srv.RejectExpertMarketRelease(
		adminCtx(99),
		connect.NewRequest(&adminv1.RejectExpertMarketReleaseRequest{
			ReleaseId: release.ID,
			Reason:    "  ",
		}),
	)

	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestExpertMarketReviewRequiresSystemAdmin(t *testing.T) {
	srv, _ := newExpertMarketAdminTestServer(t, false)

	_, err := srv.ListExpertMarketReleases(
		adminCtx(99),
		connect.NewRequest(&adminv1.ListExpertMarketReleasesRequest{}),
	)

	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

func newExpertMarketAdminTestServer(
	t *testing.T,
	isAdmin bool,
) (*Server, expertmarket.Repository) {
	t.Helper()
	gdb := newExpertMarketAdminTestDB(t)
	require.NoError(t, gdb.Exec(
		`INSERT INTO users (id, email, username, is_active, is_system_admin) VALUES (?, ?, ?, ?, ?)`,
		99,
		"admin@example.test",
		"admin-user",
		true,
		isAdmin,
	).Error)
	repo := infra.NewExpertMarketRepository(gdb)
	svc := expertsvc.NewService(expertsvc.Deps{
		Market:            repo,
		MarketInstallLock: expertMarketAdminTestLock{},
	})
	srv := NewServer(nil, database.NewGormWrapper(gdb), WithExpertService(svc))
	return srv, repo
}

func newExpertMarketAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.Exec(`PRAGMA foreign_keys = ON`).Error)
	require.NoError(t, db.Exec(`
CREATE TABLE users (
	id INTEGER PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	username TEXT NOT NULL UNIQUE,
	is_active BOOLEAN NOT NULL DEFAULT true,
	is_system_admin BOOLEAN NOT NULL DEFAULT false,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`).Error)
	require.NoError(t, db.Exec(`
CREATE TABLE expert_market_applications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	slug TEXT NOT NULL UNIQUE,
	publisher_organization_id INTEGER NOT NULL,
	source_expert_id INTEGER NOT NULL,
	publisher_user_id INTEGER NOT NULL,
	is_operator_owned BOOLEAN NOT NULL DEFAULT false,
	latest_published_release_id INTEGER,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`).Error)
	require.NoError(t, db.Exec(`
CREATE TABLE expert_market_releases (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	application_id INTEGER NOT NULL,
	source_expert_id INTEGER NOT NULL,
	publisher_organization_id INTEGER NOT NULL,
	publisher_user_id INTEGER NOT NULL,
	version INTEGER NOT NULL,
	status TEXT NOT NULL,
	name TEXT NOT NULL,
	summary TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	category TEXT NOT NULL DEFAULT '',
	icon TEXT NOT NULL DEFAULT '',
	tags TEXT NOT NULL DEFAULT '{}',
	outcomes TEXT NOT NULL DEFAULT '{}',
	featured BOOLEAN NOT NULL DEFAULT false,
	expert_snapshot BLOB NOT NULL,
	worker_spec_snapshot BLOB NOT NULL,
	skill_dependencies BLOB NOT NULL,
	reviewer_user_id INTEGER,
	rejection_reason TEXT,
	submitted_at DATETIME,
	reviewed_at DATETIME,
	published_at DATETIME,
	rejected_at DATETIME,
	withdrawn_at DATETIME,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(application_id, version),
	UNIQUE(application_id, id),
	FOREIGN KEY(application_id) REFERENCES expert_market_applications(id)
)`).Error)
	return db
}

func seedExpertMarketRelease(
	t *testing.T,
	repo expertmarket.Repository,
	slug string,
	status expertmarket.ReleaseStatus,
) expertmarket.Release {
	t.Helper()
	app := expertmarket.Application{
		Slug:                    slugkit.Slug(slug),
		PublisherOrganizationID: 10,
		SourceExpertID:          30,
		PublisherUserID:         20,
	}
	require.NoError(t, repo.CreateApplication(context.Background(), &app))
	release := expertmarket.Release{
		ApplicationID:           app.ID,
		SourceExpertID:          30,
		PublisherOrganizationID: 10,
		PublisherUserID:         20,
		Version:                 1,
		Status:                  status,
		Name:                    "Video Expert",
		Summary:                 "Cut clips",
		Description:             "Video production workflow",
		Category:                "video",
		Icon:                    "film",
		Tags:                    []string{"video"},
		Outcomes:                []string{"render"},
		Featured:                true,
		ExpertSnapshot:          json.RawMessage(`{"version":1,"slug":"` + slug + `"}`),
		WorkerSpecSnapshot:      json.RawMessage(`{"version":1,"spec":{"agent":"codex"}}`),
		SkillDependencies:       json.RawMessage(`[{"slug":"render","version":1}]`),
	}
	require.NoError(t, repo.CreateRelease(context.Background(), &release))
	return release
}

func adminCtx(userID int64) context.Context {
	return middleware.SetTenant(
		context.Background(),
		&middleware.TenantContext{UserID: userID},
	)
}

func stringPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}

type expertMarketAdminTestLock struct{}

func (expertMarketAdminTestLock) WithinMarketApplicationLock(
	_ context.Context,
	_ int64,
	apply func() error,
) error {
	return apply()
}

func (expertMarketAdminTestLock) WithinMarketInstallationLock(
	_ context.Context,
	_, _ int64,
	apply func() error,
) error {
	return apply()
}
