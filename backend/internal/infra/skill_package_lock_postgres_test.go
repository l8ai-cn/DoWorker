package infra

import (
	"context"
	"testing"
	"time"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/stretchr/testify/require"
)

func TestSkillPackageLockSerializesSameKeyAcrossRepositoriesPostgres(t *testing.T) {
	db1, db2 := openSkillMutationLockTestDBs(t)
	repo1 := NewSkillCatalogRepository(db1)
	repo2 := NewSkillCatalogRepository(db2)
	const storageKey = "skills/direct/shared/hash.tar.gz"
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)

	go func() {
		firstDone <- repo1.WithPackageLock(
			context.Background(),
			storageKey,
			func(skilldom.Repository) error {
				close(firstEntered)
				<-releaseFirst
				return nil
			},
		)
	}()
	<-firstEntered

	secondEntered := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- repo2.WithPackageLock(
			context.Background(),
			storageKey,
			func(skilldom.Repository) error {
				close(secondEntered)
				return nil
			},
		)
	}()
	select {
	case <-secondEntered:
		t.Fatal("same package key acquired concurrently")
	case <-time.After(150 * time.Millisecond):
	}
	close(releaseFirst)
	require.NoError(t, <-firstDone)
	require.NoError(t, <-secondDone)
}

func TestSkillPackageLockAllowsDifferentKeysAcrossRepositoriesPostgres(t *testing.T) {
	db1, db2 := openSkillMutationLockTestDBs(t)
	repo1 := NewSkillCatalogRepository(db1)
	repo2 := NewSkillCatalogRepository(db2)
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)

	go func() {
		firstDone <- repo1.WithPackageLock(
			context.Background(),
			"skills/direct/a/hash.tar.gz",
			func(skilldom.Repository) error {
				close(firstEntered)
				<-releaseFirst
				return nil
			},
		)
	}()
	<-firstEntered

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- repo2.WithPackageLock(
			context.Background(),
			"skills/direct/b/hash.tar.gz",
			func(skilldom.Repository) error { return nil },
		)
	}()
	select {
	case err := <-secondDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("different package key was blocked")
	}
	close(releaseFirst)
	require.NoError(t, <-firstDone)
}

func TestSkillPackageLockNestsInsideMutationLockPostgres(t *testing.T) {
	db, _ := openSkillMutationLockTestDBs(t)
	repo := NewSkillCatalogRepository(db)

	err := repo.WithMutationLock(
		context.Background(),
		42,
		func(locked skilldom.Repository) error {
			return locked.WithPackageLock(
				context.Background(),
				"skills/direct/nested/hash.tar.gz",
				func(skilldom.Repository) error { return nil },
			)
		},
	)

	require.NoError(t, err)
}
