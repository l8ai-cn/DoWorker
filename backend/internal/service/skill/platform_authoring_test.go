package skill

import (
	"context"
	"errors"
	"sync"
	"testing"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/stretchr/testify/require"
)

func TestEnsurePlatformSkillCreatesPackagedPlatformSkill(t *testing.T) {
	store := newFakeStore()
	service := NewPlatformCatalogService(
		store,
		&fakePackager{},
	)

	row, created, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)

	require.NoError(t, err)
	require.True(t, created)
	require.Nil(t, row.OrganizationID)
	require.Equal(t, "video-delivery-qa", row.Slug)
	require.Equal(t, []string{"qa", "video"}, []string(row.Tags))
	require.Equal(t, skilldom.SourceOperator, row.InstallSource)
	require.Empty(t, row.GitRepoPath)
	require.NotEmpty(t, row.ContentSha)
	require.NotEmpty(t, row.StorageKey)
	require.Positive(t, row.PackageSize)
}

func TestEnsurePlatformSkillIsIdempotentForExactManifest(t *testing.T) {
	store := newFakeStore()
	service := NewPlatformCatalogService(
		store,
		&fakePackager{},
	)
	first, created, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)
	require.NoError(t, err)
	require.True(t, created)

	second, created, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, store.rows, 1)
}

func TestEnsurePlatformSkillUsesStableOperatorPackageIdentity(t *testing.T) {
	packager := &fakePackager{}
	service := NewPlatformCatalogService(newFakeStore(), packager)

	_, _, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)

	require.NoError(t, err)
	require.Equal(
		t,
		[]string{"operator-catalog/video-delivery-qa"},
		packager.catalogIdentities,
	)
}

func TestEnsurePlatformSkillRejectsManifestDrift(t *testing.T) {
	service := NewPlatformCatalogService(
		newFakeStore(),
		&fakePackager{},
	)
	_, _, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)
	require.NoError(t, err)
	changed := platformSkillRequest()
	changed.Instructions = "Use a different release gate."

	_, _, err = service.EnsurePlatformSkill(context.Background(), changed)

	require.ErrorIs(t, err, ErrPlatformSkillConflict)
}

func TestEnsurePlatformSkillDeletesNewPackageOnCatalogConflict(t *testing.T) {
	store := newFakeStore()
	packager := &fakePackager{}
	service := NewPlatformCatalogService(store, packager)
	_, _, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)
	require.NoError(t, err)
	packager.reused = false
	changed := platformSkillRequest()
	changed.Instructions = "Use a different release gate."

	_, _, err = service.EnsurePlatformSkill(context.Background(), changed)

	require.ErrorIs(t, err, ErrPlatformSkillConflict)
	require.Len(t, packager.deletedKeys, 1)
}

func TestEnsurePlatformSkillSerializesConflictingCatalogWrites(t *testing.T) {
	store := newKeyedLockStore()
	first := NewPlatformCatalogService(store, &fakePackager{})
	second := NewPlatformCatalogService(store, &fakePackager{})
	changed := platformSkillRequest()
	changed.Instructions = "Use a different release gate."
	results := make(chan error, 2)
	go func() {
		_, _, err := first.EnsurePlatformSkill(
			context.Background(),
			platformSkillRequest(),
		)
		results <- err
	}()
	go func() {
		_, _, err := second.EnsurePlatformSkill(context.Background(), changed)
		results <- err
	}()

	errA, errB := <-results, <-results

	require.True(t,
		(errA == nil && errors.Is(errB, ErrPlatformSkillConflict)) ||
			(errB == nil && errors.Is(errA, ErrPlatformSkillConflict)),
	)
	require.Len(t, store.rows, 1)
}

func TestEnsurePlatformSkillValidatesIdentifierBeforeProvisioning(t *testing.T) {
	service := NewPlatformCatalogService(
		newFakeStore(),
		&fakePackager{},
	)
	request := platformSkillRequest()
	request.Slug = "Video_QA"

	_, _, err := service.EnsurePlatformSkill(context.Background(), request)

	require.Error(t, err)
}

func TestEnsurePlatformSkillPersistsTrimmedIdentifier(t *testing.T) {
	service := NewPlatformCatalogService(
		newFakeStore(),
		&fakePackager{},
	)
	request := platformSkillRequest()
	request.Slug = " video-delivery-qa "

	row, created, err := service.EnsurePlatformSkill(
		context.Background(),
		request,
	)

	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "video-delivery-qa", row.Slug)
}

func platformSkillRequest() *EnsurePlatformSkillRequest {
	return &EnsurePlatformSkillRequest{
		UserID:       9,
		Slug:         "video-delivery-qa",
		Name:         "Video Delivery QA",
		Description:  "Checks short-form video delivery.",
		License:      "Apache-2.0",
		Instructions: "Validate the rendered video before delivery.",
		Tags:         []string{"video", "qa"},
	}
}

var _ skilldom.Repository = (*fakeStore)(nil)

type keyedLockStore struct {
	*fakeStore
	guard sync.Mutex
	locks map[string]*sync.Mutex
}

func newKeyedLockStore() *keyedLockStore {
	return &keyedLockStore{
		fakeStore: newFakeStore(),
		locks:     make(map[string]*sync.Mutex),
	}
}

func (s *keyedLockStore) WithPackageLock(
	_ context.Context,
	key string,
	mutate func(skilldom.Repository) error,
) error {
	s.guard.Lock()
	lock := s.locks[key]
	if lock == nil {
		lock = &sync.Mutex{}
		s.locks[key] = lock
	}
	s.guard.Unlock()
	lock.Lock()
	defer lock.Unlock()
	return mutate(s)
}
