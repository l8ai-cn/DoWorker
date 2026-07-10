package airesource

import (
	"context"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixedProvidersRejectArbitraryBaseURL(t *testing.T) {
	f := newFixture()
	_, err := f.service.CreateConnection(context.Background(), actor(1), CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI", BaseURL: "https://attacker.example/v1", Credentials: map[string]string{"api_key": "secret"}})
	assert.ErrorIs(t, err, ErrInvalidEndpoint)
}

func TestCustomProviderRequiresExplicitValidatedBaseURL(t *testing.T) {
	f := newFixture()
	input := CreateConnectionInput{OwnerScope: domain.OwnerScopeUser, OwnerID: 1, Identifier: "custom-main", ProviderKey: "custom-openai-compatible", Name: "Custom", Credentials: map[string]string{"api_key": "secret"}}
	_, err := f.service.CreateConnection(context.Background(), actor(1), input)
	assert.ErrorIs(t, err, ErrInvalidEndpoint)
	input.BaseURL = "https://custom.example/v1"
	created, err := f.service.CreateConnection(context.Background(), actor(1), input)
	require.NoError(t, err)
	assert.Equal(t, input.BaseURL, created.BaseURL)
}
