package airesource

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestValidateProviderModelCapabilityRequiresSeedanceModelForVideoProviders(t *testing.T) {
	for _, provider := range []string{"doubao", "sub2api-seedance"} {
		t.Run(provider, func(t *testing.T) {
			require.NoError(t, ValidateProviderModelCapability(
				slugkit.MustNewForTest(provider),
				"doubao-seedance-2-0-260128",
				CapabilityVideoGeneration,
			))
			require.Error(t, ValidateProviderModelCapability(
				slugkit.MustNewForTest(provider),
				"doubao-seed-1-8-251228",
				CapabilityVideoGeneration,
			))
		})
	}
}
