package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestInstallationRepositoryReservesAndSettlesOnce(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() { _, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`) })
	for _, migration := range []string{
		"000001_market_foundation.up.sql",
		"000002_core_catalog.up.sql",
		"000003_catalog_version_integrity.up.sql",
		"000004_console_revisions.up.sql",
		"000005_console_publication_integrity.up.sql",
		"000006_quota_foundation.up.sql",
		"000007_entitlement_installation.up.sql",
		"000008_quota_ledger_audit.up.sql",
	} {
		applyMigration(t, sqlDB, migration)
	}
	seedInstallableListing(t, sqlDB)
	db, err := gorm.Open(gormpostgres.Open(dsn))
	require.NoError(t, err)
	repository := NewInstallationRepository(db)
	runtime := &installationRuntimeStub{}
	orchestration := service.NewInstallationOrchestrationService(
		repository,
		runtime,
		func() time.Time { return time.Now().UTC() },
	)
	ctx := context.Background()

	plan, err := orchestration.CreatePlan(ctx, service.CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ListingVersionID: 61, TargetOrganizationID: 9,
		ActorUserID:            14,
		RequestedConfiguration: json.RawMessage(`{"model_resource_id":"18"}`),
	})
	require.NoError(t, err)
	command := service.ApplyInstallationCommand{
		OperationID: plan.OperationID, PlanID: plan.PlanID,
		PlanDigest:     plan.PlanDigest,
		IdempotencyKey: "33333333-3333-4333-8333-333333333333",
		ActorUserID:    14,
	}
	first, err := orchestration.Apply(ctx, command)
	require.NoError(t, err)
	second, err := orchestration.Apply(ctx, command)
	require.NoError(t, err)

	require.Equal(t, service.ApplySucceeded, first.Status)
	require.Equal(t, first, second)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, "expert-18", first.RuntimeRef)
	assertQuotaBalances(t, sqlDB, "80.000000", "0.000000", "20.000000")
	var entitlementCount, installationCount int
	require.NoError(t, sqlDB.QueryRow(
		`SELECT COUNT(*) FROM marketplace.marketplace_entitlements`,
	).Scan(&entitlementCount))
	require.NoError(t, sqlDB.QueryRow(
		`SELECT COUNT(*) FROM marketplace.marketplace_installations`,
	).Scan(&installationCount))
	require.Equal(t, 1, entitlementCount)
	require.Equal(t, 1, installationCount)
	_, err = orchestration.GetOperation(ctx, first.OperationID, 15)
	require.ErrorIs(t, err, service.ErrOperationNotFound)

	retryPlan, err := orchestration.CreatePlan(ctx, service.CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ListingVersionID: 61, TargetOrganizationID: 9, ActorUserID: 14,
	})
	require.NoError(t, err)
	retryCommand := service.ApplyInstallationCommand{
		OperationID: retryPlan.OperationID, PlanID: retryPlan.PlanID,
		PlanDigest: retryPlan.PlanDigest, IdempotencyKey: retryPlan.OperationID,
		ActorUserID: 14,
	}
	firstExecution, existing, err := repository.BeginApply(ctx, retryCommand)
	require.NoError(t, err)
	require.False(t, existing)
	secondExecution, existing, err := repository.BeginApply(ctx, retryCommand)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, firstExecution, secondExecution)
	_, err = repository.FailApply(
		ctx,
		secondExecution,
		service.ErrRuntimeInstallationRejected,
	)
	require.NoError(t, err)
	assertConcurrentQuotaReservation(t, sqlDB, orchestration, repository)

	_, err = sqlDB.Exec(`
INSERT INTO marketplace.marketplace_quota_plans
  (id, marketplace_id, slug, name, period, grant_credits, charge_scope, status)
VALUES (72, 1, 'other-plan', '其他额度', 'total', 100, 'organization', 'active');
INSERT INTO marketplace.marketplace_quota_accounts
  (id, marketplace_id, subject_type, subject_ref, quota_plan_id, status,
   period_start, period_end)
VALUES ('cccccccc-cccc-4ccc-8ccc-cccccccccccc', 1, 'organization', 10, 72,
  'active', NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days');
`)
	require.NoError(t, err)
	_, err = orchestration.CreatePlan(ctx, service.CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ListingVersionID: 61, TargetOrganizationID: 10,
		ActorUserID: 15,
	})
	require.NoError(t, err)
	var automaticAccountCount, automaticGrantCount int
	require.NoError(t, sqlDB.QueryRow(`
SELECT COUNT(*) FROM marketplace.marketplace_quota_accounts
WHERE marketplace_id = 1 AND subject_type = 'organization'
  AND subject_ref = 10 AND quota_plan_id = 71
`).Scan(&automaticAccountCount))
	require.NoError(t, sqlDB.QueryRow(`
SELECT COUNT(*) FROM marketplace.marketplace_quota_ledger_entries le
JOIN marketplace.marketplace_quota_accounts qa ON qa.id = le.quota_account_id
WHERE qa.marketplace_id = 1 AND qa.subject_ref = 10
  AND le.entry_type = 'grant'
`).Scan(&automaticGrantCount))
	require.Equal(t, 1, automaticAccountCount)
	require.Equal(t, 1, automaticGrantCount)

	_, err = sqlDB.Exec(`
INSERT INTO marketplace.marketplace_catalog_items
  (id, publisher_id, slug, resource_type, name, summary, platform_resource_type,
   platform_resource_id, status, created_by_platform_user_id, revision)
SELECT 42, publisher_id, 'invalid-listing', resource_type, name, summary,
  platform_resource_type, 102, status,
  created_by_platform_user_id, revision
FROM marketplace.marketplace_catalog_items WHERE id = 41;
INSERT INTO marketplace.marketplace_catalog_item_versions
  (id, catalog_item_id, version, source_revision, content_digest, manifest,
   validation_status, created_by_platform_user_id)
SELECT 52, 42, version, source_revision, content_digest,
  '{"installation_credits":"bad","runtime_snapshot":{}}',
  validation_status, created_by_platform_user_id
FROM marketplace.marketplace_catalog_item_versions WHERE id = 51;
UPDATE marketplace.marketplace_catalog_items SET latest_version_id = 52 WHERE id = 42;
INSERT INTO marketplace.marketplace_listings
  (id, marketplace_id, catalog_item_id, slug, status, visibility, access_mode,
   created_at, updated_at, revision)
VALUES (42, 1, 42, 'invalid-listing', 'approved', 'public', 'direct',
  NOW(), NOW(), 1);
INSERT INTO marketplace.marketplace_listing_versions
  (id, listing_id, catalog_item_id, catalog_item_version_id, revision,
   display_name, tagline, description, quota_plan_id, review_status)
VALUES (62, 42, 42, 52, 1, '无效应用', '无效配置', '测试无效清单', 71, 'approved');
INSERT INTO marketplace.marketplace_listing_spaces
  (marketplace_id, listing_id, space_id, is_primary)
VALUES (1, 42, 11, true);
UPDATE marketplace.marketplace_listings
SET status = 'published', current_version_id = 62, published_at = NOW()
WHERE id = 42;
`)
	require.NoError(t, err)
	_, err = orchestration.CreatePlan(ctx, service.CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "invalid-listing",
		ListingVersionID: 62, TargetOrganizationID: 9,
		ActorUserID: 14,
	})
	require.ErrorIs(t, err, service.ErrListingNotFound)
}

func assertConcurrentQuotaReservation(
	t *testing.T,
	sqlDB *sql.DB,
	orchestration *service.InstallationOrchestrationService,
	repository *InstallationRepository,
) {
	t.Helper()
	_, err := sqlDB.Exec(`
INSERT INTO marketplace.marketplace_quota_ledger_entries
  (id, marketplace_id, quota_account_id, entry_type, available_delta, reason)
VALUES (gen_random_uuid(), 1, 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
  'adjust', -60, 'concurrency_test')
`)
	require.NoError(t, err)
	plans := make([]service.InstallationPlanResult, 2)
	for index := range plans {
		plans[index], err = orchestration.CreatePlan(context.Background(), service.CreateInstallationPlanCommand{
			MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
			ListingVersionID: 61, TargetOrganizationID: 9, ActorUserID: 14,
		})
		require.NoError(t, err)
	}
	lockTx, err := sqlDB.Begin()
	require.NoError(t, err)
	_, err = lockTx.Exec(`
SELECT id FROM marketplace.marketplace_quota_accounts
WHERE id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa' FOR UPDATE
`)
	require.NoError(t, err)

	type result struct {
		execution service.ApplyExecution
		err       error
	}
	results := make(chan result, 2)
	for _, plan := range plans {
		command := service.ApplyInstallationCommand{
			OperationID: plan.OperationID, PlanID: plan.PlanID,
			PlanDigest: plan.PlanDigest, IdempotencyKey: plan.OperationID,
			ActorUserID: 14,
		}
		go func() {
			execution, _, beginErr := repository.BeginApply(context.Background(), command)
			results <- result{execution: execution, err: beginErr}
		}()
	}
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, lockTx.Commit())

	var succeeded *service.ApplyExecution
	var insufficient int
	for range plans {
		outcome := <-results
		switch {
		case outcome.err == nil:
			execution := outcome.execution
			succeeded = &execution
		case errors.Is(outcome.err, service.ErrQuotaInsufficient):
			insufficient++
		default:
			require.NoError(t, outcome.err)
		}
	}
	require.NotNil(t, succeeded)
	require.Equal(t, 1, insufficient)
	_, err = repository.FailApply(
		context.Background(),
		*succeeded,
		service.ErrRuntimeInstallationRejected,
	)
	require.NoError(t, err)
}

type installationRuntimeStub struct{ calls int }

func (r *installationRuntimeStub) Authorize(context.Context, int64, int64) error {
	return nil
}

func (r *installationRuntimeStub) Install(
	context.Context,
	service.RuntimeInstallRequest,
) (service.RuntimeInstallResult, error) {
	r.calls++
	return service.RuntimeInstallResult{
		RuntimeRef: "expert-18",
		Result:     json.RawMessage(`{"created":true}`),
	}, nil
}

func assertQuotaBalances(
	t *testing.T,
	db *sql.DB,
	available string,
	reserved string,
	consumed string,
) {
	t.Helper()
	var actualAvailable, actualReserved, actualConsumed string
	err := db.QueryRow(`
SELECT SUM(available_delta)::text, SUM(reserved_delta)::text,
  SUM(consumed_delta)::text
FROM marketplace.marketplace_quota_ledger_entries
`).Scan(&actualAvailable, &actualReserved, &actualConsumed)
	require.NoError(t, err)
	require.Equal(t, available, actualAvailable)
	require.Equal(t, reserved, actualReserved)
	require.Equal(t, consumed, actualConsumed)
}

func seedInstallableListing(t *testing.T, db *sql.DB) {
	t.Helper()
	seed := strings.ReplaceAll(
		installableListingSQL,
		"$CONTENT_DIGEST$",
		strings.Repeat("a", 64),
	)
	_, err := db.Exec(seed)
	require.NoError(t, err)
}

const installableListingSQL = `
INSERT INTO marketplace.marketplaces
  (id, slug, name, summary, status, visibility, owner_platform_org_id,
   created_by_platform_user_id)
VALUES (1, 'commerce-market', '跨境电商市场', '开箱即用', 'published', 'public', 9, 14);
INSERT INTO marketplace.marketplace_spaces
  (id, marketplace_id, slug, name, summary, status, created_by_platform_user_id)
VALUES (11, 1, 'operations', '运营', '运营应用', 'published', 14);
INSERT INTO marketplace.marketplace_publishers
  (id, slug, publisher_type, display_name, verification_status)
VALUES (31, 'commerce-lab', 'platform', 'Commerce Lab', 'verified');
INSERT INTO marketplace.marketplace_catalog_items
  (id, publisher_id, slug, resource_type, name, summary, platform_resource_type,
   platform_resource_id, status, created_by_platform_user_id, revision)
VALUES (41, 31, 'listing-optimizer', 'application', '商品优化应用', '开箱即用',
  'expert', 101, 'active', 14, 1);
INSERT INTO marketplace.marketplace_catalog_item_versions
  (id, catalog_item_id, version, source_revision, content_digest, manifest,
   validation_status, created_by_platform_user_id)
VALUES (51, 41, '1.0.0', 'sha-1', '$CONTENT_DIGEST$',
  '{"installation_credits":"20","runtime_snapshot":{"name":"商品优化应用","agent_slug":"codex-cli"}}',
  'passed', 14);
UPDATE marketplace.marketplace_catalog_items SET latest_version_id = 51 WHERE id = 41;
INSERT INTO marketplace.marketplace_quota_plans
  (id, marketplace_id, slug, name, period, grant_credits, charge_scope,
   status)
VALUES (71, 1, 'organization-standard', '组织标准额度', 'total', 100,
  'organization', 'active');
INSERT INTO marketplace.marketplace_quota_accounts
  (id, marketplace_id, subject_type, subject_ref, quota_plan_id, status,
   period_start, period_end)
VALUES ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 1, 'organization', 9, 71,
  'active', NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days');
INSERT INTO marketplace.marketplace_listings
  (id, marketplace_id, catalog_item_id, slug, status, visibility, access_mode,
   created_at, updated_at, revision)
VALUES (41, 1, 41, 'listing-optimizer', 'approved', 'public', 'direct',
  NOW(), NOW(), 1);
INSERT INTO marketplace.marketplace_listing_versions
  (id, listing_id, catalog_item_id, catalog_item_version_id, revision,
   display_name, tagline, description, quota_plan_id, review_status)
VALUES (61, 41, 41, 51, 1, '商品优化应用', '批量优化', '面向跨境运营', 71, 'approved');
INSERT INTO marketplace.marketplace_listing_spaces
  (marketplace_id, listing_id, space_id, is_primary)
VALUES (1, 41, 11, true);
UPDATE marketplace.marketplace_listings
SET status = 'published', current_version_id = 61, published_at = NOW()
WHERE id = 41;
INSERT INTO marketplace.marketplace_quota_ledger_entries
  (id, marketplace_id, quota_account_id, entry_type, available_delta,
   period_start, reason)
VALUES ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 1,
  'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'grant', 100, NOW(), 'initial_grant');
`
