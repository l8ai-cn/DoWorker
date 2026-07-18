package workerdefinition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildEnvironmentBundlePolicySeparatesManagedAndCredentialFields(
	t *testing.T,
) {
	policy := BuildEnvironmentBundlePolicy(Definition{
		ModelRequirement: ModelRequirement{Required: true},
		CredentialBindings: []CredentialBinding{
			{
				Source: CredentialSource{Kind: "credential_bundle"},
				Target: CredentialTarget{Name: "SECOND_API_KEY"},
			},
			{
				Source: CredentialSource{Kind: "model_resource"},
				Target: CredentialTarget{Name: "OPENAI_BASE_URL"},
			},
			{
				Source: CredentialSource{Kind: "credential_bundle"},
				Target: CredentialTarget{Name: "FIRST_API_KEY"},
			},
		},
	})

	assert.Equal(
		t,
		[]string{"OPENAI_BASE_URL", "model"},
		policy.ModelManagedFields,
	)
	assert.Equal(
		t,
		[]string{"FIRST_API_KEY", "SECOND_API_KEY"},
		policy.CredentialBundleFields,
	)
}
