package podconnect

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	workerspecdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
)

type workerSpecLoaderStub struct {
	snapshot  workerspecdom.Snapshot
	snapshots []workerspecdom.Snapshot
	batchIDs  []int64
}

func (stub workerSpecLoaderStub) GetByID(
	context.Context,
	int64,
	int64,
) (workerspecdom.Snapshot, error) {
	return stub.snapshot, nil
}

func (stub *workerSpecLoaderStub) GetByIDs(
	_ context.Context,
	_ int64,
	ids []int64,
) ([]workerspecdom.Snapshot, error) {
	stub.batchIDs = append([]int64(nil), ids...)
	return stub.snapshots, nil
}

func TestApplyWorkerSkillsProjectsSnapshotBindings(t *testing.T) {
	loader := &workerSpecLoaderStub{
		snapshot: workerspecdom.Snapshot{
			Spec: workerspecdom.Spec{
				Workspace: workerspecdom.Workspace{
					SkillPackages: []workerspecdom.SkillPackageBinding{
						{Slug: "seedance-expert"},
						{Slug: "video-delivery-qa"},
					},
				},
			},
		},
	}
	server := NewServer(nil, nil, WithWorkerSpecSnapshotLoader(loader))
	item := &podv1.Pod{WorkerSpecSnapshotId: int64Pointer(19)}

	err := server.applyWorkerSkills(context.Background(), 7, []*podv1.Pod{item})

	require.NoError(t, err)
	require.Equal(t, []string{"seedance-expert", "video-delivery-qa"}, item.WorkerSkillSlugs)
}

func TestApplyWorkerSkillsLoadsMultipleSnapshotsOnce(t *testing.T) {
	loader := &workerSpecLoaderStub{
		snapshots: []workerspecdom.Snapshot{
			{ID: 19, Spec: workerspecdom.Spec{Workspace: workerspecdom.Workspace{
				SkillPackages: []workerspecdom.SkillPackageBinding{{Slug: "seedance-expert"}},
			}}},
			{ID: 20, Spec: workerspecdom.Spec{Workspace: workerspecdom.Workspace{
				SkillPackages: []workerspecdom.SkillPackageBinding{{Slug: "video-delivery-qa"}},
			}}},
		},
	}
	server := NewServer(nil, nil, WithWorkerSpecSnapshotLoader(loader))
	items := []*podv1.Pod{
		{WorkerSpecSnapshotId: int64Pointer(19)},
		{WorkerSpecSnapshotId: int64Pointer(20)},
	}

	err := server.applyWorkerSkills(context.Background(), 7, items)

	require.NoError(t, err)
	require.ElementsMatch(t, []int64{19, 20}, loader.batchIDs)
	require.Equal(t, []string{"seedance-expert"}, items[0].WorkerSkillSlugs)
	require.Equal(t, []string{"video-delivery-qa"}, items[1].WorkerSkillSlugs)
}

func int64Pointer(value int64) *int64 {
	return &value
}
