package orchestrationcontrol

import (
	"context"
	"errors"
	"testing"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceCapabilitiesAllowCreatorToViewReferenceAndPlan(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.repository.resourceSequence = []resourceRead{{head: orchestrationServiceHead()}}

	result, err := fixture.service(t).GetResourceCapabilities(
		context.Background(),
		fixture.scope,
		orchestrationServiceTarget(),
	)

	require.NoError(t, err)
	assert.Equal(t, ResourceCapabilities{
		Exists:        true,
		CanViewSource: true,
		CanReference:  true,
		CanPlan:       true,
	}, result)
}

func TestResourceCapabilitiesExposeExplicitUpdateDenial(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.repository.resourceSequence = []resourceRead{{head: orchestrationServiceHead()}}
	fixture.authorizer.updateErr = ErrForbidden

	result, err := fixture.service(t).GetResourceCapabilities(
		context.Background(),
		fixture.scope,
		orchestrationServiceTarget(),
	)

	require.NoError(t, err)
	assert.True(t, result.Exists)
	assert.True(t, result.CanViewSource)
	assert.True(t, result.CanReference)
	assert.False(t, result.CanPlan)
}

func TestResourceCapabilitiesAuthorizeCreateForMissingTarget(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)

	result, err := fixture.service(t).GetResourceCapabilities(
		context.Background(),
		fixture.scope,
		orchestrationServiceTarget(),
	)

	require.NoError(t, err)
	assert.Equal(t, ResourceCapabilities{CanPlan: true}, result)
	assert.Equal(t, 1, fixture.authorizer.createCalls)
	assert.Zero(t, fixture.authorizer.referenceCalls)
	assert.Zero(t, fixture.authorizer.updateCalls)
}

func TestResourceCapabilitiesDoNotMaskInfrastructureFailures(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	expected := errors.New("member lookup failed")
	fixture.repository.resourceSequence = []resourceRead{{head: orchestrationServiceHead()}}
	fixture.authorizer.referenceErr = expected

	_, err := fixture.service(t).GetResourceCapabilities(
		context.Background(),
		fixture.scope,
		orchestrationServiceTarget(),
	)

	assert.ErrorIs(t, err, expected)
}

func TestResourceCapabilitiesRejectCorruptHeadBeforeAuthorization(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	head := orchestrationServiceHead()
	head.Identity.Name = slugkit.MustNewForTest("other-resource")
	fixture.repository.resourceSequence = []resourceRead{{head: head}}

	_, err := fixture.service(t).GetResourceCapabilities(
		context.Background(),
		fixture.scope,
		orchestrationServiceTarget(),
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
	assert.Zero(t, fixture.authorizer.referenceCalls)
	assert.Zero(t, fixture.authorizer.updateCalls)
}
