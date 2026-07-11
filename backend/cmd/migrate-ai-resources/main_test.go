package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

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

func TestHelpDoesNotPrintResolvedCipherKey(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret-that-must-not-appear")

	fs := flag.NewFlagSet("migrate-ai-resources", flag.ContinueOnError)
	var output bytes.Buffer
	fs.SetOutput(&output)
	registerFlags(fs)

	err := fs.Parse([]string{"-h"})
	if err != flag.ErrHelp {
		t.Fatalf("parse help error = %v, want %v", err, flag.ErrHelp)
	}
	help := output.String()
	if strings.Contains(help, "secret-that-must-not-appear") {
		t.Fatalf("help output leaked resolved cipher key: %s", help)
	}
	if !strings.Contains(help, "$JWT_SECRET") {
		t.Fatalf("help output should document JWT_SECRET fallback: %s", help)
	}
}
