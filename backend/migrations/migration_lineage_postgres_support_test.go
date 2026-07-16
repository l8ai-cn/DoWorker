package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"
)

const latestMigrationVersion = 224

//go:embed testdata/legacy_000209/*.sql
var legacy000209Fixtures embed.FS

func newMigrationLineageSchema(t *testing.T) string {
	t.Helper()
	dsn, err := migrationPostgresDSN()
	require.NoError(t, err)
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}

	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	require.Contains(t, []string{"postgres", "postgresql"}, parsed.Scheme)

	schema := fmt.Sprintf("migration_lineage_%d", time.Now().UnixNano())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	_, err = db.Exec(`CREATE SCHEMA ` + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`)
		_ = db.Close()
	})

	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func newPostgresMigrator(t *testing.T, sourceFS fs.FS, dsn string) *migrate.Migrate {
	t.Helper()
	source, err := iofs.New(sourceFS, ".")
	require.NoError(t, err)
	instance, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	require.NoError(t, err)
	return instance
}

func closePostgresMigrator(t *testing.T, instance *migrate.Migrate) {
	t.Helper()
	sourceErr, databaseErr := instance.Close()
	require.NoError(t, sourceErr)
	require.NoError(t, databaseErr)
}

func migrationFSUntil(t *testing.T, maxVersion uint) fstest.MapFS {
	t.Helper()
	result := fstest.MapFS{}
	entries, err := FS.ReadDir(".")
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationName.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, err := strconv.ParseUint(matches[1], 10, 64)
		require.NoError(t, err)
		if version > uint64(maxVersion) {
			continue
		}
		content, err := FS.ReadFile(entry.Name())
		require.NoError(t, err)
		result[entry.Name()] = &fstest.MapFile{Data: content}
	}
	return result
}

func legacy000209MigrationFS(t *testing.T) fstest.MapFS {
	t.Helper()
	result := migrationFSUntil(t, 206)
	fixtures, err := fs.Sub(legacy000209Fixtures, "testdata/legacy_000209")
	require.NoError(t, err)
	entries, err := fs.ReadDir(fixtures, ".")
	require.NoError(t, err)
	for _, entry := range entries {
		content, err := fs.ReadFile(fixtures, entry.Name())
		require.NoError(t, err)
		result[entry.Name()] = &fstest.MapFile{Data: content}
	}
	return result
}

func failing000222MigrationFS(t *testing.T) fstest.MapFS {
	t.Helper()
	result := migrationFSUntil(t, 221)
	result["000222_dirty_probe.up.sql"] = &fstest.MapFile{
		Data: []byte(`SELECT 1 / 0;`),
	}
	result["000222_dirty_probe.down.sql"] = &fstest.MapFile{
		Data: []byte(`SELECT 1;`),
	}
	return result
}
