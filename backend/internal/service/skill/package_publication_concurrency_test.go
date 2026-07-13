package skill

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type packageLockSet struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func (s *packageLockSet) with(storageKey string, mutate func() error) error {
	s.mu.Lock()
	if s.locks == nil {
		s.locks = make(map[string]*sync.Mutex)
	}
	lock := s.locks[storageKey]
	if lock == nil {
		lock = &sync.Mutex{}
		s.locks[storageKey] = lock
	}
	s.mu.Unlock()
	lock.Lock()
	defer lock.Unlock()
	return mutate()
}

type publicationRequestStore struct {
	*fakeStore
	locks        *packageLockSet
	createErr    error
	beforeCreate func()
}

func (s *publicationRequestStore) WithPackageLock(
	_ context.Context,
	storageKey string,
	mutate func(skilldom.Repository) error,
) error {
	return s.locks.with(storageKey, func() error { return mutate(s) })
}

func (s *publicationRequestStore) Create(ctx context.Context, row *skilldom.Skill) error {
	if s.beforeCreate != nil {
		s.beforeCreate()
	}
	if s.createErr != nil {
		return s.createErr
	}
	return s.fakeStore.Create(ctx, row)
}

type publicationArtifacts struct {
	mu      sync.Mutex
	exists  bool
	uploads int
	deletes int
}

type publicationPackager struct {
	artifacts *publicationArtifacts
	onStore   func()
}

const sharedCatalogKey = "skills/catalog/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef/shared-content.tar.gz"

func (p *publicationPackager) PrepareCatalogFromDir(
	context.Context,
	string,
	string,
) (*extensionsvc.PreparedSkill, error) {
	return &extensionsvc.PreparedSkill{
		Slug:        "shared-skill",
		ContentSha:  "shared-content",
		StorageKey:  sharedCatalogKey,
		PackageSize: 4,
		Data:        []byte("data"),
	}, nil
}

func (p *publicationPackager) StorePrepared(
	_ context.Context,
	prepared *extensionsvc.PreparedSkill,
) (*extensionsvc.PackagedSkill, error) {
	p.artifacts.mu.Lock()
	created := !p.artifacts.exists
	if created {
		p.artifacts.exists = true
		p.artifacts.uploads++
	}
	p.artifacts.mu.Unlock()
	if p.onStore != nil {
		p.onStore()
	}
	return &extensionsvc.PackagedSkill{
		Slug:        prepared.Slug,
		ContentSha:  prepared.ContentSha,
		StorageKey:  prepared.StorageKey,
		PackageSize: prepared.PackageSize,
		Created:     created,
	}, nil
}

func (p *publicationPackager) DeletePackage(context.Context, string) error {
	p.artifacts.mu.Lock()
	p.artifacts.exists = false
	p.artifacts.deletes++
	p.artifacts.mu.Unlock()
	return nil
}

func TestConcurrentCreateSerializesPackagePublicationAndFailedCleanup(t *testing.T) {
	rows := newFakeStore()
	locks := &packageLockSet{}
	artifacts := &publicationArtifacts{}
	aAtCreate := make(chan struct{})
	allowAFailure := make(chan struct{})
	bStored := make(chan struct{})
	storeA := &publicationRequestStore{
		fakeStore: rows,
		locks:     locks,
		createErr: errors.New("request A database failure"),
		beforeCreate: func() {
			close(aAtCreate)
			<-allowAFailure
		},
	}
	storeB := &publicationRequestStore{fakeStore: rows, locks: locks}
	svcA := newTestService(
		storeA,
		gitops.NewFake("am-skills"),
		&publicationPackager{artifacts: artifacts},
	)
	svcB := newTestService(
		storeB,
		gitops.NewFake("am-skills"),
		&publicationPackager{
			artifacts: artifacts,
			onStore:   func() { close(bStored) },
		},
	)
	request := func(orgID int64) *CreateSkillRequest {
		return &CreateSkillRequest{
			OrganizationID: orgID,
			Name:           "Shared Skill",
			Slug:           "shared-skill",
			Instructions:   "Same content.",
		}
	}

	aDone := make(chan error, 1)
	go func() {
		_, err := svcA.Create(context.Background(), request(1))
		aDone <- err
	}()
	<-aAtCreate
	bDone := make(chan struct {
		row *skilldom.Skill
		err error
	}, 1)
	go func() {
		row, err := svcB.Create(context.Background(), request(2))
		bDone <- struct {
			row *skilldom.Skill
			err error
		}{row: row, err: err}
	}()

	select {
	case <-bStored:
		t.Fatal("request B stored the package before request A compensated")
	case <-time.After(100 * time.Millisecond):
	}
	close(allowAFailure)
	require.ErrorContains(t, <-aDone, "request A database failure")
	resultB := <-bDone
	require.NoError(t, resultB.err)
	require.NotNil(t, resultB.row)

	artifacts.mu.Lock()
	defer artifacts.mu.Unlock()
	assert.True(t, artifacts.exists)
	assert.Equal(t, 2, artifacts.uploads)
	assert.Equal(t, 1, artifacts.deletes)
	assert.Equal(t, resultB.row.StorageKey, sharedCatalogKey)
	saved, err := rows.GetByID(context.Background(), 2, resultB.row.ID)
	require.NoError(t, err)
	assert.Equal(t, resultB.row.StorageKey, saved.StorageKey)
}
