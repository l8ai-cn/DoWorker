package workerspec

import (
	"context"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResolvedSnapshotRebuildsCanonicalAggregate(t *testing.T) {
	expected, err := NewResolver(newResolverPortsForTest().deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		validDraftForTest(),
	)
	require.NoError(t, err)
	spec, err := domain.DecodeSpec(expected.SpecJSON())
	require.NoError(t, err)

	actual, err := NewResolvedSnapshot(validScopeForTest().OrgID, spec)

	require.NoError(t, err)
	assert.Equal(t, expected.OrganizationID(), actual.OrganizationID())
	assert.Equal(t, expected.Version(), actual.Version())
	assert.Equal(t, expected.SpecJSON(), actual.SpecJSON())
	assert.Equal(t, expected.SummaryJSON(), actual.SummaryJSON())
}

func TestNewResolvedSnapshotRejectsInvalidScopeAndSpec(t *testing.T) {
	expected, err := NewResolver(newResolverPortsForTest().deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		validDraftForTest(),
	)
	require.NoError(t, err)
	spec, err := domain.DecodeSpec(expected.SpecJSON())
	require.NoError(t, err)

	_, err = NewResolvedSnapshot(0, spec)
	assert.Error(t, err)
	spec.Runtime.Image.ID = 0
	_, err = NewResolvedSnapshot(validScopeForTest().OrgID, spec)
	assert.Error(t, err)
}

func TestNewResolvedSnapshotReturnsDetachedDocuments(t *testing.T) {
	expected, err := NewResolver(newResolverPortsForTest().deps()).Resolve(
		context.Background(),
		validScopeForTest(),
		validDraftForTest(),
	)
	require.NoError(t, err)
	spec, err := domain.DecodeSpec(expected.SpecJSON())
	require.NoError(t, err)
	resolved, err := NewResolvedSnapshot(validScopeForTest().OrgID, spec)
	require.NoError(t, err)

	specJSON := resolved.SpecJSON()
	summaryJSON := resolved.SummaryJSON()
	specJSON[0] = '['
	summaryJSON[0] = '['

	assert.JSONEq(t, string(expected.SpecJSON()), string(resolved.SpecJSON()))
	assert.JSONEq(t, string(expected.SummaryJSON()), string(resolved.SummaryJSON()))
}
