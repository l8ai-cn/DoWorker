package orchestrationresourceconnect

import (
	"testing"

	"connectrpc.com/connect"

	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetResourceCapabilitiesUsesResolvedScopeAndReturnsAuthority(t *testing.T) {
	stub := &serviceStub{
		capabilitiesResult: service.ResourceCapabilities{
			Exists:        true,
			CanViewSource: true,
			CanReference:  true,
			CanPlan:       false,
		},
	}
	server := newTestServer(stub, testOrganizations())

	response, err := server.GetResourceCapabilities(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.GetResourceCapabilitiesRequest{
			OrgSlug: "acme",
			Target: &resourcev1.ResourceTarget{
				TypeMeta: &resourcev1.TypeMeta{
					ApiVersion: "agentsmesh.io/v1alpha1",
					Kind:       "Expert",
				},
				Namespace: "acme",
				Name:      "reviewer",
			},
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, int64(81), stub.capabilitiesScope.OrganizationID)
	assert.Equal(t, int64(42), stub.capabilitiesScope.ActorID)
	assert.Equal(t, "Expert", stub.capabilitiesTarget.Kind)
	require.NotNil(t, response.Msg.Capabilities)
	assert.True(t, response.Msg.Capabilities.Exists)
	assert.True(t, response.Msg.Capabilities.CanViewSource)
	assert.True(t, response.Msg.Capabilities.CanReference)
	assert.False(t, response.Msg.Capabilities.CanPlan)
}
