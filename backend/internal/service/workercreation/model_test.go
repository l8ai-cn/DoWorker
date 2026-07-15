package workercreation

import (
	"context"
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	resourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelResolverBindsExactResourceAndConnectionRevisions(t *testing.T) {
	resources := &modelResourceService{
		resolved: &resourceservice.ResolvedResource{
			Connection: resourcedomain.Connection{
				ID:          201,
				ProviderKey: slugkit.MustNewForTest("openai"),
				Revision:    9,
			},
			Resource: resourcedomain.ModelResource{
				ID:                   101,
				ProviderConnectionID: 201,
				ModelID:              "gpt-5",
				Revision:             7,
			},
			Credentials: map[string]string{"api_key": "must-not-leak"},
		},
	}
	resolver := newModelResolver(resources)

	binding, err := resolver.ResolveModel(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		101,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(101), binding.ResourceID)
	assert.Equal(t, int64(7), binding.ResourceRevision)
	assert.Equal(t, int64(201), binding.ConnectionID)
	assert.Equal(t, int64(9), binding.ConnectionRevision)
	assert.Equal(t, slugkit.MustNewForTest("openai"), binding.ProviderKey)
	assert.Equal(t, "gpt-5", binding.ModelID)
	assert.Equal(t, resourceservice.Actor{UserID: 7}, resources.actor)
	assert.Equal(t, int64(77), resources.orgID)
	assert.Equal(t, int64(101), resources.resourceID)
	assert.Equal(t, resourcedomain.ModalityChat, resources.requirements.Modality)
	assert.Equal(t, resourcedomain.CapabilityTextGeneration, resources.requirements.Capability)
	assert.Equal(t, []string{"openai-compatible"}, resources.requirements.AllowedProtocolAdapters)
}

func TestModelResolverUsesWorkerSpecificProtocolAdapters(t *testing.T) {
	tests := []struct {
		workerType string
		adapter    string
	}{
		{workerType: "codex-cli", adapter: "openai-compatible"},
		{workerType: "video-studio", adapter: "openai-compatible"},
		{workerType: "claude-code", adapter: "anthropic"},
		{workerType: "gemini-cli", adapter: "gemini"},
	}

	for _, test := range tests {
		t.Run(test.workerType, func(t *testing.T) {
			resources := validModelResourceService()

			_, err := newModelResolver(resources).ResolveModel(
				context.Background(),
				specservice.Scope{OrgID: 77, UserID: 7},
				slugkit.MustNewForTest(test.workerType),
				101,
			)

			require.NoError(t, err)
			assert.Equal(t, []string{test.adapter}, resources.requirements.AllowedProtocolAdapters)
		})
	}
}

func TestModelResolverRejectsUnsupportedOrInvalidSelections(t *testing.T) {
	t.Run("unsupported worker type", func(t *testing.T) {
		resources := validModelResourceService()

		_, err := newModelResolver(resources).ResolveModel(
			context.Background(),
			specservice.Scope{OrgID: 77, UserID: 7},
			slugkit.MustNewForTest("do-agent"),
			101,
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
		assert.Zero(t, resources.calls)
	})

	t.Run("forbidden model resource", func(t *testing.T) {
		resources := validModelResourceService()
		resources.err = resourceservice.ErrForbidden

		_, err := newModelResolver(resources).ResolveModel(
			context.Background(),
			specservice.Scope{OrgID: 77, UserID: 7},
			slugkit.MustNewForTest("codex-cli"),
			101,
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
		assert.ErrorIs(t, err, resourceservice.ErrForbidden)
	})

	t.Run("infrastructure error is not disguised", func(t *testing.T) {
		resources := validModelResourceService()
		resources.err = assert.AnError

		_, err := newModelResolver(resources).ResolveModel(
			context.Background(),
			specservice.Scope{OrgID: 77, UserID: 7},
			slugkit.MustNewForTest("codex-cli"),
			101,
		)

		assert.ErrorIs(t, err, assert.AnError)
		assert.NotErrorIs(t, err, specservice.ErrInvalidDraft)
	})
}

type modelResourceService struct {
	resolved     *resourceservice.ResolvedResource
	err          error
	calls        int
	actor        resourceservice.Actor
	orgID        int64
	resourceID   int64
	requirements resourceservice.ResolutionRequirements
}

func (service *modelResourceService) ResolveExact(
	_ context.Context,
	actor resourceservice.Actor,
	orgID, resourceID int64,
	requirements resourceservice.ResolutionRequirements,
) (*resourceservice.ResolvedResource, error) {
	service.calls++
	service.actor = actor
	service.orgID = orgID
	service.resourceID = resourceID
	service.requirements = requirements
	return service.resolved, service.err
}

func validModelResourceService() *modelResourceService {
	return &modelResourceService{
		resolved: &resourceservice.ResolvedResource{
			Connection: resourcedomain.Connection{
				ID:          201,
				ProviderKey: slugkit.MustNewForTest("openai"),
				Revision:    9,
			},
			Resource: resourcedomain.ModelResource{
				ID:                   101,
				ProviderConnectionID: 201,
				ModelID:              "gpt-5",
				Revision:             7,
			},
		},
	}
}
