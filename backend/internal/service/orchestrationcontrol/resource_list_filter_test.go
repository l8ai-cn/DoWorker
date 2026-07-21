package orchestrationcontrol

import (
	"testing"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceListFilterAcceptsEnvironmentBundleReferenceContext(
	t *testing.T,
) {
	filter := ResourceListFilter{
		Kind:   resource.KindEnvironmentBundle,
		Limit:  100,
		Offset: 0,
		EnvironmentBundle: &EnvironmentBundleReferenceFilter{
			Purpose:    EnvironmentBundlePurposeCredential,
			WorkerType: slugkit.Slug("cursor-cli"),
			TargetName: "CURSOR_API_KEY",
		},
	}

	require.NoError(t, filter.Validate(orchestrationServiceScope()))
}

func TestResourceListFilterRejectsInvalidEnvironmentBundleContext(
	t *testing.T,
) {
	valid := EnvironmentBundleReferenceFilter{
		Purpose:    EnvironmentBundlePurposeConfig,
		WorkerType: slugkit.Slug("do-agent"),
	}
	tests := []ResourceListFilter{
		{Kind: resource.KindPrompt, Limit: 100, EnvironmentBundle: &valid},
		{
			Kind: resource.KindEnvironmentBundle, Limit: 100,
			EnvironmentBundle: &EnvironmentBundleReferenceFilter{
				Purpose: "unknown", WorkerType: slugkit.Slug("do-agent"),
			},
		},
		{
			Kind: resource.KindEnvironmentBundle, Limit: 100,
			EnvironmentBundle: &EnvironmentBundleReferenceFilter{
				Purpose:    EnvironmentBundlePurposeRuntime,
				WorkerType: slugkit.Slug("Cursor_CLI"),
			},
		},
		{
			Kind: resource.KindEnvironmentBundle, Limit: 100,
			EnvironmentBundle: &EnvironmentBundleReferenceFilter{
				Purpose:    EnvironmentBundlePurposeCredential,
				WorkerType: slugkit.Slug("cursor-cli"),
			},
		},
		{
			Kind: resource.KindEnvironmentBundle, Limit: 100,
			EnvironmentBundle: &EnvironmentBundleReferenceFilter{
				Purpose:    EnvironmentBundlePurposeConfig,
				WorkerType: slugkit.Slug("do-agent"),
				TargetName: "CONFIG_PATH",
			},
		},
		{
			Kind: resource.KindEnvironmentBundle, Limit: 100,
			EnvironmentBundle: &EnvironmentBundleReferenceFilter{
				Purpose:            EnvironmentBundlePurposeRuntime,
				WorkerType:         slugkit.Slug("do-agent"),
				ModelManagedFields: []string{"MODEL", "MODEL"},
			},
		},
	}

	for _, filter := range tests {
		assert.ErrorIs(
			t,
			filter.Validate(orchestrationServiceScope()),
			control.ErrInvalid,
		)
	}
}

func TestResourceListFilterRejectsModelBindingFilterOutsideModelBindingKind(
	t *testing.T,
) {
	filter := ResourceListFilter{
		Kind:  resource.KindPrompt,
		Limit: 100,
		ModelBinding: &ModelBindingReferenceFilter{
			WorkerType: slugkit.MustNewForTest("minimax-cli"),
		},
	}

	assert.ErrorIs(t, filter.Validate(orchestrationServiceScope()), control.ErrInvalid)
}
