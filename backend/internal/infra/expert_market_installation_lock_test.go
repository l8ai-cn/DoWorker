package infra

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestExpertMarketLockRejectsUnsupportedDatabase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	applied := false
	err = NewExpertMarketInstallationLocker(db).WithinMarketInstallationLock(
		context.Background(),
		1,
		2,
		func() error {
			applied = true
			return nil
		},
	)

	require.ErrorIs(t, err, ErrExpertMarketLockRequiresPostgres)
	require.False(t, applied)
	require.Contains(t, err.Error(), "sqlite")
	require.True(t, errors.Is(err, ErrExpertMarketLockRequiresPostgres))
}
