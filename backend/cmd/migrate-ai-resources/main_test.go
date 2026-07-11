package main

import "testing"

func TestDefaultDSNPrefersDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://explicit")
	t.Setenv("DB_HOST", "ignored-host")

	if got := defaultDSN(); got != "postgres://explicit" {
		t.Fatalf("default DSN = %q, want explicit DATABASE_URL", got)
	}
}

func TestDefaultDSNBuildsFromDatabaseEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "worker")
	t.Setenv("DB_PASSWORD", "test-password")
	t.Setenv("DB_NAME", "workerdb")
	t.Setenv("DB_SSLMODE", "require")

	want := "host=postgres port=5433 user=worker password=test-password dbname=workerdb sslmode=require"
	if got := defaultDSN(); got != want {
		t.Fatalf("default DSN = %q, want %q", got, want)
	}
}
