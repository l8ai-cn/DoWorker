package coordinator

import (
	"context"
	"errors"
	"testing"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	workerspecdom "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	runnersvc "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func TestRunProjectEnsuresRunnerForSnapshotWorkerType(t *testing.T) {
	selector := &fakeRunnerSelector{}
	snapshots := &fakeWorkerSpecSnapshots{snapshot: testCoordinatorSnapshot("codex-cli")}
	svc := NewService(Deps{
		Store:         newFakeStore(),
		Tickets:       &fakeTickets{},
		Dispatch:      &fakeDispatch{},
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: &fakePlatform{}, repo: "org/repo"},
		RunnerEnsurer: NewRunnerEnsurer(selector, &fakeRunnerLauncher{}, nil),
		Snapshots:     snapshots,
	})
	project := &coordinatordom.Project{
		ID: 7, OrganizationID: 11, CreatedByID: 23, RepositoryID: 3,
		PlatformType:         coordinatordom.PlatformTypeCNB,
		AgentSlug:            "do-agent",
		WorkerSpecSnapshotID: testCoordinatorSnapshotID(),
	}

	_, err := svc.RunProject(context.Background(), project)

	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if selector.agentSlug != "codex-cli" {
		t.Fatalf("runner worker type = %q, want codex-cli", selector.agentSlug)
	}
	if snapshots.organizationID != 11 || snapshots.id != 101 {
		t.Fatalf("snapshot lookup = org %d id %d, want org 11 id 101", snapshots.organizationID, snapshots.id)
	}
}

func TestRunProjectRequiresSnapshotStoreBeforeRunnerEnsure(t *testing.T) {
	svc := NewService(Deps{
		Store:         newFakeStore(),
		Tickets:       &fakeTickets{},
		Dispatch:      &fakeDispatch{},
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: &fakePlatform{}, repo: "org/repo"},
		RunnerEnsurer: NewRunnerEnsurer(&fakeRunnerSelector{}, &fakeRunnerLauncher{}, nil),
	})
	project := &coordinatordom.Project{
		ID: 7, OrganizationID: 11, CreatedByID: 23, RepositoryID: 3,
		PlatformType:         coordinatordom.PlatformTypeCNB,
		WorkerSpecSnapshotID: testCoordinatorSnapshotID(),
	}

	_, err := svc.RunProject(context.Background(), project)

	if !errors.Is(err, ErrCoordinatorWorkerSpecSnapshotStoreRequired) {
		t.Fatalf("RunProject error = %v, want snapshot store required", err)
	}
}

func TestRunProjectLaunchesRunnerForSnapshotWorkerType(t *testing.T) {
	selector := &fakeRunnerSelector{err: runnersvc.ErrNoRunnerForAgent}
	launcher := &fakeRunnerLauncher{}
	snapshots := &fakeWorkerSpecSnapshots{snapshot: testCoordinatorSnapshot("codex-cli")}
	ensurer := NewRunnerEnsurer(selector, launcher, nil)
	ensurer.wait = 0
	ensurer.poll = 0
	svc := NewService(Deps{
		Store:         newFakeStore(),
		Tickets:       &fakeTickets{},
		Dispatch:      &fakeDispatch{},
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: &fakePlatform{}, repo: "org/repo"},
		RunnerEnsurer: ensurer,
		Snapshots:     snapshots,
	})
	project := &coordinatordom.Project{
		ID: 7, OrganizationID: 11, CreatedByID: 23, RepositoryID: 3,
		PlatformType:         coordinatordom.PlatformTypeCNB,
		AgentSlug:            "do-agent",
		WorkerSpecSnapshotID: testCoordinatorSnapshotID(),
	}

	_, _ = svc.RunProject(context.Background(), project)

	if launcher.agentSlug != "codex-cli" {
		t.Fatalf("launcher worker type = %q, want codex-cli", launcher.agentSlug)
	}
}

type fakeWorkerSpecSnapshots struct {
	snapshot       workerspecdom.Snapshot
	organizationID int64
	id             int64
	err            error
}

func (f *fakeWorkerSpecSnapshots) GetByID(
	_ context.Context,
	organizationID int64,
	id int64,
) (workerspecdom.Snapshot, error) {
	f.organizationID = organizationID
	f.id = id
	return f.snapshot, f.err
}

func testCoordinatorSnapshot(workerType string) workerspecdom.Snapshot {
	return workerspecdom.Snapshot{
		ID:             101,
		OrganizationID: 11,
		Spec: workerspecdom.Spec{
			Runtime: workerspecdom.Runtime{
				WorkerType: workerspecdom.WorkerType{
					Slug: slugkit.MustNewForTest(workerType),
				},
			},
		},
	}
}
