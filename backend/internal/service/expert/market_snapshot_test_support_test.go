package expert

import (
	"context"
	"sync"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type fakeMarketSnapshots struct {
	source         specdomain.Snapshot
	created        []specdomain.Snapshot
	preparedScopes []specservice.Scope
	preparedModels []int64
	err            error
	deleteContexts []context.Context
	deleteErrors   []error
}

type fakeMarketInstallationLocker struct {
	applicationMutex  sync.Mutex
	installationMutex sync.Mutex
	calls             int
	applicationCalls  int
	applicationHook   func()
}

func (locker *fakeMarketInstallationLocker) WithinMarketApplicationLock(
	_ context.Context,
	_ int64,
	apply func() error,
) error {
	locker.applicationMutex.Lock()
	defer locker.applicationMutex.Unlock()
	locker.applicationCalls++
	if locker.applicationHook != nil {
		hook := locker.applicationHook
		locker.applicationHook = nil
		hook()
	}
	return apply()
}

func (locker *fakeMarketInstallationLocker) WithinMarketInstallationLock(
	_ context.Context,
	_, _ int64,
	apply func() error,
) error {
	locker.installationMutex.Lock()
	defer locker.installationMutex.Unlock()
	locker.calls++
	return apply()
}

func (snapshots *fakeMarketSnapshots) GetByID(
	_ context.Context,
	organizationID, snapshotID int64,
) (specdomain.Snapshot, error) {
	if snapshots.err != nil {
		return specdomain.Snapshot{}, snapshots.err
	}
	for _, snapshot := range snapshots.created {
		if snapshot.ID == snapshotID &&
			snapshot.OrganizationID == organizationID {
			return snapshot, nil
		}
	}
	if snapshots.source.ID != snapshotID ||
		snapshots.source.OrganizationID != organizationID {
		return specdomain.Snapshot{}, specdomain.ErrNotFound
	}
	return snapshots.source, nil
}

func (snapshots *fakeMarketSnapshots) Create(
	_ context.Context,
	resolved specservice.ResolvedSnapshot,
) (specdomain.Snapshot, error) {
	if snapshots.err != nil {
		return specdomain.Snapshot{}, snapshots.err
	}
	spec, err := specdomain.DecodeSpec(resolved.SpecJSON())
	if err != nil {
		return specdomain.Snapshot{}, err
	}
	created := specdomain.Snapshot{
		ID:             int64(len(snapshots.created) + 1000),
		OrganizationID: resolved.OrganizationID(),
		Spec:           spec,
	}
	snapshots.created = append(snapshots.created, created)
	return created, nil
}

func (snapshots *fakeMarketSnapshots) PrepareMarketSnapshot(
	_ context.Context,
	scope specservice.Scope,
	source specdomain.Spec,
	modelResourceID int64,
) (specservice.ResolvedSnapshot, error) {
	if snapshots.err != nil {
		return specservice.ResolvedSnapshot{}, snapshots.err
	}
	source.Runtime.ModelBinding.ResourceID = modelResourceID
	snapshots.preparedScopes = append(snapshots.preparedScopes, scope)
	snapshots.preparedModels = append(snapshots.preparedModels, modelResourceID)
	return specservice.NewResolvedSnapshot(scope.OrgID, source)
}

func (snapshots *fakeMarketSnapshots) Delete(
	ctx context.Context,
	organizationID, snapshotID int64,
) error {
	snapshots.deleteContexts = append(snapshots.deleteContexts, ctx)
	err := ctx.Err()
	snapshots.deleteErrors = append(snapshots.deleteErrors, err)
	if err != nil {
		return err
	}
	for index, snapshot := range snapshots.created {
		if snapshot.OrganizationID == organizationID && snapshot.ID == snapshotID {
			snapshots.created = append(
				snapshots.created[:index],
				snapshots.created[index+1:]...,
			)
			return nil
		}
	}
	return specdomain.ErrNotFound
}
