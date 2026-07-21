package operatorcatalog

import (
	"context"
	"errors"
	"strings"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type bootstrapWorkerPreparer struct {
	calls int
}

func (preparer *bootstrapWorkerPreparer) Revision() string {
	return "test-revision"
}

func (preparer *bootstrapWorkerPreparer) Prepare(
	_ context.Context,
	scope specservice.Scope,
	draft workercreation.Draft,
) (workercreation.Prepared, error) {
	preparer.calls++
	if err := slugkit.Validate(draft.OrganizationSlug.String()); err != nil {
		return workercreation.Prepared{}, err
	}
	if draft.OrganizationSlug != scope.OrgSlug {
		return workercreation.Prepared{}, errors.New("organization slug mismatch")
	}
	spec := specdomain.NewV1(
		specdomain.Runtime{
			ModelBinding: specdomain.ModelBinding{
				ResourceID:       draft.WorkerSpec.ModelResourceID,
				ResourceRevision: 1,
				ConnectionID:     1, ConnectionRevision: 1,
				ProviderKey:     slugkit.MustNewForTest("openai"),
				ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
				ModelID:         "gpt-5",
			},
			WorkerType: specdomain.WorkerType{
				Slug:           draft.WorkerSpec.WorkerTypeSlug,
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: specdomain.RuntimeImage{
				ID:     draft.WorkerSpec.Runtime.RuntimeImageID,
				Digest: "sha256:" + strings.Repeat("b", 64),
			},
		},
		specdomain.Placement{
			Policy: draft.WorkerSpec.Runtime.PlacementPolicy,
			ComputeTarget: specdomain.ComputeTarget{
				ID:   draft.WorkerSpec.Runtime.ComputeTargetID,
				Kind: specdomain.ComputeTargetKindRunnerPool,
			},
			DeploymentMode: draft.WorkerSpec.Runtime.DeploymentMode,
			ResourceProfile: specdomain.ResourceProfile{
				ID: draft.WorkerSpec.Runtime.ResourceProfileID,
				Resources: specdomain.ResourceRequestsLimits{
					CPURequestMilliCPU: 200, CPULimitMilliCPU: 1000,
					MemoryRequestBytes: 256 << 20, MemoryLimitBytes: 1 << 30,
				},
			},
		},
		draft.WorkerSpec.TypeConfig,
		draft.WorkerSpec.Workspace,
		draft.WorkerSpec.Lifecycle,
		draft.WorkerSpec.Metadata,
	)
	resolved, err := specservice.NewResolvedSnapshot(scope.OrgID, spec)
	return workercreation.Prepared{
		Snapshot: resolved,
		Spec:     spec,
		Artifact: &workerdependencyartifact.Artifact{},
	}, err
}

type bootstrapSnapshotStore struct {
	createCalls int
	rows        map[int64]specdomain.Snapshot
}

func newBootstrapSnapshotStore() *bootstrapSnapshotStore {
	return &bootstrapSnapshotStore{rows: map[int64]specdomain.Snapshot{}}
}

func (store *bootstrapSnapshotStore) Create(
	_ context.Context,
	resolved specservice.ResolvedSnapshot,
) (specdomain.Snapshot, error) {
	store.createCalls++
	spec, err := specdomain.DecodeSpec(resolved.SpecJSON())
	if err != nil {
		return specdomain.Snapshot{}, err
	}
	row := specdomain.Snapshot{
		ID: int64(store.createCalls), OrganizationID: resolved.OrganizationID(),
		Spec: spec,
	}
	if store.rows == nil {
		store.rows = map[int64]specdomain.Snapshot{}
	}
	store.rows[row.ID] = row
	return row, nil
}

func (store *bootstrapSnapshotStore) GetByID(
	_ context.Context,
	organizationID, id int64,
) (specdomain.Snapshot, error) {
	row, ok := store.rows[id]
	if !ok || row.OrganizationID != organizationID {
		return specdomain.Snapshot{}, specdomain.ErrNotFound
	}
	return row, nil
}

func (*bootstrapSnapshotStore) Delete(context.Context, int64, int64) error {
	return nil
}
