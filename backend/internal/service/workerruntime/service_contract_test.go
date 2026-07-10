package workerruntime

import (
	"context"
	"errors"
	"fmt"
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
	assert.Equal(t, []string{
		"image:77:41",
		"target:77:52",
		"profile:77:63",
		"image-compatible:77:codex-cli:41",
		"deployment-compatible:77:52:dedicated",
		"profile-compatible:77:52:63",
	}, repo.calls)
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
	calls           []string
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

func (repo *runtimeRepoForTest) GetRuntimeImageByIDForOrganization(
	_ context.Context,
	organizationID, imageID int64,
) (*domain.RuntimeImage, error) {
	repo.calls = append(repo.calls, fmt.Sprintf("image:%d:%d", organizationID, imageID))
	if organizationID != 77 || imageID != 41 || repo.image == nil {
		return nil, domain.ErrNotFound
	}
	image := *repo.image
	return &image, nil
}

func (repo *runtimeRepoForTest) GetComputeTargetByIDForOrganization(
	_ context.Context,
	organizationID, targetID int64,
) (*domain.ComputeTarget, error) {
	repo.calls = append(repo.calls, fmt.Sprintf("target:%d:%d", organizationID, targetID))
	if organizationID != 77 || targetID != 52 || repo.target == nil {
		return nil, domain.ErrNotFound
	}
	target := *repo.target
	return &target, nil
}

func (repo *runtimeRepoForTest) GetResourceProfileByIDForOrganization(
	_ context.Context,
	organizationID, profileID int64,
) (*domain.ResourceProfile, error) {
	repo.calls = append(repo.calls, fmt.Sprintf("profile:%d:%d", organizationID, profileID))
	if organizationID != 77 || profileID != 63 || repo.profile == nil {
		return nil, domain.ErrNotFound
	}
	profile := *repo.profile
	return &profile, nil
}

func (repo *runtimeRepoForTest) IsRuntimeImageCompatibleWithWorkerType(
	_ context.Context,
	organizationID int64,
	workerType slugkit.Slug,
	imageID int64,
) (bool, error) {
	repo.calls = append(repo.calls, fmt.Sprintf(
		"image-compatible:%d:%s:%d",
		organizationID,
		workerType.String(),
		imageID,
	))
	if organizationID != 77 || workerType.String() != "codex-cli" || imageID != 41 {
		return false, domain.ErrNotFound
	}
	return repo.imageOK, nil
}

func (repo *runtimeRepoForTest) IsComputeTargetCompatibleWithDeployment(
	_ context.Context,
	organizationID, targetID int64,
	mode workerspec.DeploymentMode,
) (bool, error) {
	repo.calls = append(repo.calls, fmt.Sprintf(
		"deployment-compatible:%d:%d:%s",
		organizationID,
		targetID,
		mode,
	))
	if organizationID != 77 || targetID != 52 || mode != workerspec.DeploymentModeDedicated {
		return false, domain.ErrNotFound
	}
	return repo.deploymentOK, nil
}

func (repo *runtimeRepoForTest) IsComputeTargetCompatibleWithResourceProfile(
	_ context.Context,
	organizationID, targetID, profileID int64,
) (bool, error) {
	repo.calls = append(repo.calls, fmt.Sprintf(
		"profile-compatible:%d:%d:%d",
		organizationID,
		targetID,
		profileID,
	))
	if organizationID != 77 || targetID != 52 || profileID != 63 {
		return false, domain.ErrNotFound
	}
	return repo.targetProfileOK, nil
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
