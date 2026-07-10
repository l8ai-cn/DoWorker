package workerruntime

import (
	"context"
	"errors"
	"strings"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveReturnsScopedRuntimePlacement(t *testing.T) {
	repo := newRuntimeRepoForTest()
	request := validRuntimeRequest()

	resolved, err := NewService(repo).Resolve(context.Background(), request)
	require.NoError(t, err)

	assert.Equal(t, workerspec.RuntimeImage{
		ID:     41,
		Digest: "sha256:" + strings.Repeat("b", 64),
	}, resolved.RuntimeImage)
	assert.Equal(t, workerspec.Placement{
		Policy: request.PlacementPolicy,
		ComputeTarget: workerspec.ComputeTarget{
			ID:   52,
			Kind: workerspec.ComputeTargetKindKubernetes,
		},
		DeploymentMode: request.DeploymentMode,
		ResourceProfile: workerspec.ResourceProfile{
			ID:        63,
			Resources: repo.profile.Resources,
		},
	}, resolved.Placement)
	assert.Equal(t, 1, repo.atomicCalls)
}

func TestResolveUsesSingleAtomicRepositorySelection(t *testing.T) {
	repo := newRuntimeRepoForTest()

	_, err := NewService(repo).Resolve(context.Background(), validRuntimeRequest())

	require.NoError(t, err)
	assert.Equal(t, 1, repo.atomicCalls)
}

func TestResolveFailsWithoutFallback(t *testing.T) {
	tests := []struct {
		name    string
		repo    *runtimeRepoForTest
		request domain.Request
		want    error
	}{
		{
			name:    "invalid request",
			repo:    newRuntimeRepoForTest(),
			request: domain.Request{},
			want:    domain.ErrInvalidRequest,
		},
		{
			name:    "missing image",
			repo:    newRuntimeRepoForTest().withoutImage(),
			request: validRuntimeRequest(),
			want:    domain.ErrNotFound,
		},
		{
			name:    "disabled target",
			repo:    newRuntimeRepoForTest().withDisabledTarget(),
			request: validRuntimeRequest(),
			want:    domain.ErrDisabled,
		},
		{
			name:    "id mismatch",
			repo:    newRuntimeRepoForTest().withImageID(99),
			request: validRuntimeRequest(),
			want:    domain.ErrInvalidResolvedValue,
		},
		{
			name:    "incompatible image",
			repo:    newRuntimeRepoForTest().withImageCompatibility(false),
			request: validRuntimeRequest(),
			want:    domain.ErrIncompatible,
		},
		{
			name:    "invalid resolved resources",
			repo:    newRuntimeRepoForTest().withCPURequest(2000),
			request: validRuntimeRequest(),
			want:    domain.ErrInvalidResolvedValue,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewService(test.repo).Resolve(context.Background(), test.request)
			require.Error(t, err)
			assert.True(t, errors.Is(err, test.want), "got %v", err)
		})
	}

	_, err := NewService(nil).Resolve(context.Background(), validRuntimeRequest())
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrRepositoryUnavailable))
}

type runtimeRepoForTest struct {
	image           *domain.RuntimeImage
	target          *domain.ComputeTarget
	profile         *domain.ResourceProfile
	imageOK         bool
	deploymentOK    bool
	targetProfileOK bool
	atomicCalls     int
}

func (repo *runtimeRepoForTest) ResolveSelection(
	_ context.Context,
	request domain.Request,
) (*domain.RepositorySelection, error) {
	repo.atomicCalls++
	if request != validRuntimeRequest() {
		return nil, domain.ErrNotFound
	}
	return &domain.RepositorySelection{
		RuntimeImage:              cloneRuntimeImage(repo.image),
		ComputeTarget:             cloneComputeTarget(repo.target),
		ResourceProfile:           cloneResourceProfile(repo.profile),
		ImageCompatible:           repo.imageOK,
		DeploymentCompatible:      repo.deploymentOK,
		ResourceProfileCompatible: repo.targetProfileOK,
	}, nil
}

func newRuntimeRepoForTest() *runtimeRepoForTest {
	return &runtimeRepoForTest{
		image: &domain.RuntimeImage{
			ID:      41,
			Digest:  "  sha256:" + strings.Repeat("b", 64) + "  ",
			Enabled: true,
		},
		target: &domain.ComputeTarget{
			ID:      52,
			Kind:    workerspec.ComputeTargetKindKubernetes,
			Enabled: true,
		},
		profile: &domain.ResourceProfile{
			ID: 63,
			Resources: workerspec.ResourceRequestsLimits{
				CPURequestMilliCPU: 500,
				CPULimitMilliCPU:   1000,
				MemoryRequestBytes: 512 << 20,
				MemoryLimitBytes:   1024 << 20,
			},
			Enabled: true,
		},
		imageOK:         true,
		deploymentOK:    true,
		targetProfileOK: true,
	}
}

func (repo *runtimeRepoForTest) withoutImage() *runtimeRepoForTest {
	repo.image = nil
	return repo
}

func (repo *runtimeRepoForTest) withDisabledTarget() *runtimeRepoForTest {
	repo.target.Enabled = false
	return repo
}

func (repo *runtimeRepoForTest) withImageID(id int64) *runtimeRepoForTest {
	repo.image.ID = id
	return repo
}

func (repo *runtimeRepoForTest) withImageCompatibility(ok bool) *runtimeRepoForTest {
	repo.imageOK = ok
	return repo
}

func (repo *runtimeRepoForTest) withCPURequest(cpu uint32) *runtimeRepoForTest {
	repo.profile.Resources.CPURequestMilliCPU = cpu
	return repo
}

func cloneRuntimeImage(image *domain.RuntimeImage) *domain.RuntimeImage {
	if image == nil {
		return nil
	}
	cloned := *image
	return &cloned
}

func cloneComputeTarget(target *domain.ComputeTarget) *domain.ComputeTarget {
	if target == nil {
		return nil
	}
	cloned := *target
	return &cloned
}

func cloneResourceProfile(profile *domain.ResourceProfile) *domain.ResourceProfile {
	if profile == nil {
		return nil
	}
	cloned := *profile
	return &cloned
}

func validRuntimeRequest() domain.Request {
	return domain.Request{
		OrganizationID:    77,
		WorkerTypeSlug:    slugkit.MustNewForTest("codex-cli"),
		RuntimeImageID:    41,
		PlacementPolicy:   workerspec.PlacementPolicyExplicit,
		ComputeTargetID:   52,
		DeploymentMode:    workerspec.DeploymentModeDedicated,
		ResourceProfileID: 63,
	}
}
