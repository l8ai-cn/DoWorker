package infra

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

func TestSkillMutationLockSerializesRepositoryInstancesPostgres(t *testing.T) {
	dsn := os.Getenv("MIGRATIONS_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}
	db1, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	db2, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB1, err := db1.DB()
	require.NoError(t, err)
	defer sqlDB1.Close()
	sqlDB2, err := db2.DB()
	require.NoError(t, err)
	defer sqlDB2.Close()
	repo1 := NewSkillCatalogRepository(db1)
	repo2 := NewSkillCatalogRepository(db2)
	skillID := time.Now().UnixNano()
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- repo1.WithMutationLock(
			context.Background(),
			skillID,
			func(skilldom.Repository) error {
				close(firstEntered)
				<-releaseFirst
				return nil
			},
		)
	}()
	select {
	case <-firstEntered:
	case err := <-firstDone:
		require.NoError(t, err)
		t.Fatal("first repository exited before entering the lock")
	case <-time.After(2 * time.Second):
		t.Fatal("first repository did not acquire the skill lock")
	}

	secondEntered := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- repo2.WithMutationLock(
			context.Background(),
			skillID,
			func(skilldom.Repository) error {
				close(secondEntered)
				return nil
			},
		)
	}()
	select {
	case <-secondEntered:
		t.Fatal("second repository acquired the same skill lock early")
	case <-time.After(150 * time.Millisecond):
	}
	close(releaseFirst)
	require.NoError(t, <-firstDone)
	select {
	case <-secondEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("second repository did not acquire the released skill lock")
	}
	require.NoError(t, <-secondDone)
}
