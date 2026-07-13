package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000207SkillTagsUpDownPostgres(t *testing.T) {
	dsn, err := migrationPostgresDSN()
	require.NoError(t, err)
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("skill_tags_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `CREATE TABLE skills (id BIGINT PRIMARY KEY)`))

	up, err := FS.ReadFile("000207_skill_tags.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))

	var dataType, udtName, nullable, defaultValue string
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT data_type, udt_name, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = current_schema() AND table_name = 'skills' AND column_name = 'tags'
`).Scan(&dataType, &udtName, &nullable, &defaultValue))
	require.Equal(t, "ARRAY", dataType)
	require.Equal(t, "_text", udtName)
	require.Equal(t, "NO", nullable)
	require.Contains(t, defaultValue, "'{}'")

	require.NoError(t, execSQL(ctx, conn, `INSERT INTO skills (id) VALUES (1)`))
	var tags pq.StringArray
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT tags FROM skills WHERE id = 1`).Scan(&tags))
	require.Empty(t, tags)

	var indexDefinition string
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT indexdef
FROM pg_indexes
WHERE schemaname = current_schema() AND tablename = 'skills' AND indexname = 'idx_skills_tags'
`).Scan(&indexDefinition))
	require.Contains(t, indexDefinition, "USING gin (tags)")

	down, err := FS.ReadFile("000207_skill_tags.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.False(t, postgresColumnExists(ctx, t, conn, "skills", "tags"))

	var indexCount int
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT count(*)
FROM pg_indexes
WHERE schemaname = current_schema() AND tablename = 'skills' AND indexname = 'idx_skills_tags'
`).Scan(&indexCount))
	require.Zero(t, indexCount)
}
