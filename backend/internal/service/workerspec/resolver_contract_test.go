package workerspec

import (
	"context"
	"errors"
	"strings"
	"testing"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolverBuildsCanonicalSnapshotFromScopedResolutions(t *testing.T) {
	ports := newResolverPortsForTest()
	draft := validDraftForTest()

	resolved, err := NewResolver(ports.deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		draft,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"worker-type",
		"runtime",
		"model",
		"secret:api-token",
		"secret:signing-key",
		"workspace",
	}, ports.calls)
	assert.Equal(t, []Scope{
		validScopeForTest(),
		validScopeForTest(),
		validScopeForTest(),
		validScopeForTest(),
		validScopeForTest(),
		validScopeForTest(),
	}, ports.scopes)
	assert.Equal(t, draft.Runtime, ports.runtimeSelection)
	assert.Equal(t, draft.WorkerTypeSlug, ports.runtimeWorkerType)
	assert.Equal(t, draft.ModelResourceID, ports.modelResourceID)

	spec, err := domain.DecodeSpec(resolved.SpecJSON())
	require.NoError(t, err)
	assert.Equal(t, ports.workerType.WorkerType, spec.Runtime.WorkerType)
	assert.Equal(t, ports.modelBinding, spec.Runtime.ModelBinding)
	assert.Equal(t, ports.runtime.RuntimeImage, spec.Runtime.Image)
	assert.Equal(t, ports.runtime.Placement, spec.Placement)
	assert.Equal(t, ports.workerType.TypeSchema.Version, spec.TypeConfig.SchemaVersion)
	assert.Equal(t, []int64{3, 9}, spec.Workspace.SkillIDs)
	assert.Equal(t, "worker", spec.Metadata.Alias)

	summary, err := domain.DecodeSummary(resolved.SummaryJSON())
	require.NoError(t, err)
	assert.Equal(t, spec.Runtime.ModelBinding, summary.ModelBinding)
	assert.Equal(t, spec.Runtime.WorkerType, summary.WorkerType)
}

func TestResolverSkipsModelResolutionForWorkerWithoutModelRequirement(t *testing.T) {
	ports := newResolverPortsForTest()
	ports.workerType.WorkerType.Slug = mustSlugForTest("cursor-cli")
	ports.workerType.ModelRequirement = domain.ModelRequirement{}
	draft := validDraftForTest()
	draft.WorkerTypeSlug = mustSlugForTest("cursor-cli")
	draft.ModelResourceID = 0

	resolved, err := NewResolver(ports.deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		draft,
	)

	require.NoError(t, err)
	assert.NotContains(t, ports.calls, "model")
	spec, err := domain.DecodeSpec(resolved.SpecJSON())
	require.NoError(t, err)
	assert.True(t, spec.Runtime.ModelBinding.IsEmpty())
}

func TestResolverValidatesTypeConfigBeforeReferencesAndWorkspace(t *testing.T) {
	ports := newResolverPortsForTest()
	draft := validDraftForTest()
	draft.TypeConfig.Values["mode"] = "unsupported"

	resolved, err := NewResolver(ports.deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		draft,
	)

	require.ErrorContains(t, err, "invalid option")
	assert.Equal(t, ResolvedSnapshot{}, resolved)
	assert.Equal(t, []string{"worker-type", "runtime", "model"}, ports.calls)
}

func TestCreateSnapshotDoesNotPersistAfterAnyResolutionFailure(t *testing.T) {
	resolutionError := errors.New("resolution failed")
	tests := []struct {
		name      string
		failAt    string
		scope     Scope
		mutate    func(*Draft)
		wantError error
	}{
		{name: "cross scope", scope: Scope{OrgID: 78, UserID: 7}, wantError: errCrossScopeForTest},
		{name: "worker type", failAt: "worker-type", scope: validScopeForTest(), wantError: resolutionError},
		{name: "runtime", failAt: "runtime", scope: validScopeForTest(), wantError: resolutionError},
		{name: "model", failAt: "model", scope: validScopeForTest(), wantError: resolutionError},
		{name: "schema", scope: validScopeForTest(), mutate: func(draft *Draft) {
			draft.TypeConfig.SchemaVersion++
		}},
		{name: "secret", failAt: "secret", scope: validScopeForTest(), wantError: resolutionError},
		{name: "workspace", failAt: "workspace", scope: validScopeForTest(), wantError: resolutionError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ports := newResolverPortsForTest()
			ports.failAt = test.failAt
			ports.failure = resolutionError
			repository := &snapshotRepositoryForTest{}
			service := NewService(NewResolver(ports.deps()), repository)
			draft := validDraftForTest()
			if test.mutate != nil {
				test.mutate(&draft)
			}

			snapshot, err := service.CreateSnapshot(context.Background(), test.scope, draft)

			require.Error(t, err)
			if test.wantError != nil {
				assert.ErrorIs(t, err, test.wantError)
			}
			assert.Equal(t, domain.Snapshot{}, snapshot)
			assert.Zero(t, repository.createCalls)
		})
	}
}

func TestCreateSnapshotPersistsJSONDetachedFromDraft(t *testing.T) {
	ports := newResolverPortsForTest()
	repository := &snapshotRepositoryForTest{}
	service := NewService(NewResolver(ports.deps()), repository)
	draft := validDraftForTest()

	_, err := service.CreateSnapshot(context.Background(), validScopeForTest(), draft)
	require.NoError(t, err)
	require.Equal(t, 1, repository.createCalls)

	draft.TypeConfig.Values["mode"] = "changed"
	draft.TypeConfig.SecretRefs["api-token"] = domain.SecretReference{}
	draft.Workspace.SkillIDs[0] = 999
	*draft.Workspace.RepositoryID = 999

	specJSON := repository.captured.SpecJSON()
	specJSON[0] = ' '
	spec, err := domain.DecodeSpec(repository.captured.SpecJSON())
	require.NoError(t, err)
	assert.Equal(t, "careful", spec.TypeConfig.Values["mode"])
	assert.Equal(t, int64(81), spec.TypeConfig.SecretRefs["api-token"].ID)
	assert.Equal(t, []int64{3, 9}, spec.Workspace.SkillIDs)
	require.NotNil(t, spec.Workspace.RepositoryID)
	assert.Equal(t, int64(22), *spec.Workspace.RepositoryID)
}

func TestResolverRejectsRuntimeOrBindingSubstitution(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*resolverPortsForTest)
	}{
		{"worker type", func(ports *resolverPortsForTest) {
			ports.workerType.WorkerType.Slug = mustSlugForTest("other-worker")
		}},
		{"runtime image", func(ports *resolverPortsForTest) {
			ports.runtime.RuntimeImage.ID++
		}},
		{"model resource", func(ports *resolverPortsForTest) {
			ports.modelBinding.ResourceID++
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ports := newResolverPortsForTest()
			test.mutate(ports)

			resolved, err := NewResolver(ports.deps()).Resolve(
				context.Background(),
				validScopeForTest(),
				validDraftForTest(),
			)

			require.Error(t, err)
			assert.Equal(t, ResolvedSnapshot{}, resolved)
		})
	}
}

func validResolvedRuntimeForTest() workerruntime.Resolved {
	return workerruntime.Resolved{
		RuntimeImage: domain.RuntimeImage{
			ID:     41,
			Digest: "sha256:" + strings.Repeat("b", 64),
		},
		Placement: validPlacementForTest(),
	}
}
