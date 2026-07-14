package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000208ExpertMarketplaceUpDownPostgres(t *testing.T) {
	dsn, err := migrationPostgresDSN()
	require.NoError(t, err)
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("expert_marketplace_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execMigrationSQL(ctx, conn, expertMarketplaceBaseDDL))

	up, err := FS.ReadFile("000208_expert_marketplace.up.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(up)))

	requireMarketplaceConstraints(t, ctx, conn)

	down, err := FS.ReadFile("000208_expert_marketplace.down.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(down)))
	require.False(t, postgresColumnExists(ctx, t, conn, "experts", "source_market_application_id"))
	require.False(t, postgresColumnExists(ctx, t, conn, "experts", "source_market_release_id"))
	require.False(t, expertMarketTableExists(ctx, t, conn, "expert_market_releases"))
	require.False(t, expertMarketTableExists(ctx, t, conn, "expert_market_applications"))
}

func requireMarketplaceConstraints(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	require.NoError(t, execSQL(ctx, conn, `
INSERT INTO expert_market_applications
	(id, slug, publisher_organization_id, publisher_user_id)
VALUES (10, 'video-production', 1, 1), (20, 'video-editing', 2, 2)
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO expert_market_applications
	(slug, publisher_organization_id, publisher_user_id)
VALUES ('Bad_Slug', 1, 1)
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO expert_market_applications
	(slug, publisher_organization_id, publisher_user_id)
VALUES ('video-production', 2, 2)
`))

	require.NoError(t, insertMarketRelease(ctx, conn, 100, 10, 1, "pending_review"))
	invalidSnapshots := []string{
		`{}`,
		`{"version":null}`,
		`{"version":"1"}`,
		`{"version":0}`,
		`{"version":-1}`,
		`{"version":1.5}`,
	}
	for index, snapshot := range invalidSnapshots {
		require.Error(t, insertMarketReleaseWithSnapshots(
			ctx, conn, 200+index, 10, 10+index,
			snapshot, `{"version":1}`, `[]`,
		))
		require.Error(t, insertMarketReleaseWithSnapshots(
			ctx, conn, 300+index, 10, 20+index,
			`{"version":1}`, snapshot, `[]`,
		))
	}
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO expert_market_releases (
	id, application_id, source_expert_id,
	publisher_organization_id, publisher_user_id,
	version, status, name,
	expert_snapshot, worker_spec_snapshot, skill_dependencies
)
VALUES (
	105, 10, 1, 2, 1, 2, 'draft', 'Wrong Publisher',
	'{"version":1}', '{"version":1}', '[]'
)
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO expert_market_releases (
	id, application_id, source_expert_id,
	publisher_organization_id, publisher_user_id,
	version, status, name,
	expert_snapshot, worker_spec_snapshot, skill_dependencies
)
VALUES (
	106, 10, 9002, 1, 1, 30, 'draft', 'Foreign Expert',
	'{"version":1}', '{"version":1}', '[]'
)
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO expert_market_releases (
	id, application_id, source_expert_id,
	publisher_organization_id, publisher_user_id,
	version, status, name,
	expert_snapshot, worker_spec_snapshot, skill_dependencies
)
VALUES (
	107, 10, 9999, 1, 1, 31, 'draft', 'Missing Expert',
	'{"version":1}', '{"version":1}', '[]'
)
`))
	require.Error(t, insertMarketRelease(ctx, conn, 101, 10, 1, "draft"))
	require.Error(t, insertMarketReleaseWithSnapshots(ctx, conn, 102, 10, 2, `[]`, `{"version":1}`, `[]`))
	require.Error(t, insertMarketReleaseWithSnapshots(ctx, conn, 103, 10, 2, `{"version":1}`, `[]`, `[]`))
	require.Error(t, insertMarketReleaseWithSnapshots(ctx, conn, 104, 10, 2, `{"version":1}`, `{"version":1}`, `{}`))

	require.Error(t, execSQL(ctx, conn, `UPDATE expert_market_releases SET name = 'changed' WHERE id = 100`))
	require.Error(t, execSQL(ctx, conn, `UPDATE expert_market_releases SET id = 999 WHERE id = 100`))
	require.Error(t, execSQL(ctx, conn, `UPDATE expert_market_releases SET created_at = now() + interval '1 day' WHERE id = 100`))
	require.NoError(t, execSQL(ctx, conn, `
UPDATE expert_market_releases
SET status = 'published', reviewer_user_id = 2, reviewed_at = now(), published_at = now()
WHERE id = 100
`))
	require.Error(t, execSQL(ctx, conn, `
UPDATE expert_market_applications SET latest_published_release_id = 100 WHERE id = 20
`))
	require.NoError(t, execSQL(ctx, conn, `
UPDATE expert_market_applications SET latest_published_release_id = 100 WHERE id = 10
`))

	require.NoError(t, execSQL(ctx, conn, `
INSERT INTO experts
	(id, organization_id, source_market_application_id, source_market_release_id)
VALUES (1, 2, 10, 100), (2, 1, 10, 100)
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO experts
	(id, organization_id, source_market_application_id, source_market_release_id)
VALUES (3, 2, 10, 100)
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO experts
	(id, organization_id, source_market_application_id, source_market_release_id)
VALUES (4, 2, 10, NULL)
`))

	require.NoError(t, execSQL(ctx, conn, `DELETE FROM experts WHERE id = 9001`))
	var sourceExpertID int64
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT source_expert_id FROM expert_market_releases WHERE id = 100
`).Scan(&sourceExpertID))
	require.Equal(t, int64(9001), sourceExpertID)

	require.NoError(t, execSQL(ctx, conn, `DELETE FROM organizations WHERE id = 1`))
	var appID, releaseID sql.NullInt64
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT source_market_application_id, source_market_release_id FROM experts WHERE id = 1
`).Scan(&appID, &releaseID))
	require.False(t, appID.Valid)
	require.False(t, releaseID.Valid)
}

func insertMarketRelease(
	ctx context.Context,
	conn *sql.Conn,
	id, applicationID, version int,
	status string,
) error {
	return insertMarketReleaseWithSnapshots(
		ctx, conn, id, applicationID, version,
		`{"version":1}`, `{"version":1}`, `[]`,
		status,
	)
}

func insertMarketReleaseWithSnapshots(
	ctx context.Context,
	conn *sql.Conn,
	id, applicationID, version int,
	expertSnapshot, workerSnapshot, dependencies string,
	status ...string,
) error {
	releaseStatus := "draft"
	if len(status) > 0 {
		releaseStatus = status[0]
	}
	_, err := conn.ExecContext(ctx, `
INSERT INTO expert_market_releases (
	id, application_id, source_expert_id,
	publisher_organization_id, publisher_user_id,
	version, status, name,
	expert_snapshot, worker_spec_snapshot, skill_dependencies
)
VALUES ($1, $2, 9001, 1, 1, $3, $4, 'Video Expert', $5::jsonb, $6::jsonb, $7::jsonb)
`, id, applicationID, version, releaseStatus, expertSnapshot, workerSnapshot, dependencies)
	return err
}

func expertMarketTableExists(
	ctx context.Context,
	t *testing.T,
	conn *sql.Conn,
	table string,
) bool {
	t.Helper()
	var exists bool
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1 FROM information_schema.tables
	WHERE table_schema = current_schema() AND table_name = $1
)`, table).Scan(&exists))
	return exists
}

const expertMarketplaceBaseDDL = `
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
CREATE TABLE experts (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE
);
INSERT INTO users(id) VALUES (1), (2);
INSERT INTO organizations(id) VALUES (1), (2);
INSERT INTO experts(id, organization_id) VALUES (9001, 1), (9002, 2);
`
