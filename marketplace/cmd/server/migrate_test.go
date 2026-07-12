package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrateUpRejectsInvalidDatabaseURL(t *testing.T) {
	require.Error(t, migrateUp("not-a-postgres-url"))
}
