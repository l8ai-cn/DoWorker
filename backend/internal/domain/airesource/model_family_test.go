package airesource

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestValidateProviderModelCapabilityRequiresProviderSpecificVideoModel(t *testing.T) {
	doubao := slugkit.MustNewForTest("doubao")
	require.NoError(t, ValidateProviderModelCapability(
		doubao, "doubao-seedance-2-0-260128", CapabilityVideoGeneration,
	))
	require.Error(t, ValidateProviderModelCapability(
		doubao, "creative-video", CapabilityVideoGeneration,
	))

	sub2api := slugkit.MustNewForTest("sub2api-seedance")
	require.NoError(t, ValidateProviderModelCapability(
		sub2api, "creative-video", CapabilityVideoGeneration,
	))
	require.Error(t, ValidateProviderModelCapability(
		sub2api, "doubao-seedance-2-0-260128", CapabilityVideoGeneration,
	))
}
