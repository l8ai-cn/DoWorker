package agentpod

import (
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func planSkillsForArtifactTest(
	t *testing.T,
	scope control.Scope,
	ids []int64,
) []workerdependencyartifact.SkillResolution {
	t.Helper()
	out := make([]workerdependencyartifact.SkillResolution, 0, len(ids))
	for _, id := range ids {
		name := slugkit.MustNewForTest("skill-" + formatPlanID(id))
		ref := planResolvedReferenceForTest(scope, resource.KindSkill, name.String(), 6)
		out = append(out, workerdependencyartifact.SkillResolution{
			ResourceResolution: planResourceProjectionForTest(t, scope, ref, id),
			Slug:               name, Version: 1,
			ContentDigest: workerdependency.TextDigest(name.String()),
			StorageKey:    "skills/" + name.String() + ".tgz",
			PackageSize:   1,
		})
	}
	return out
}

func planKnowledgeForArtifactTest(
	t *testing.T,
	scope control.Scope,
	mounts []specdomain.KnowledgeMount,
) []workerdependencyartifact.KnowledgeBaseResolution {
	t.Helper()
	out := make([]workerdependencyartifact.KnowledgeBaseResolution, 0, len(mounts))
	for _, mount := range mounts {
		name := planKnowledgeSlugForArtifactTest(mount.KnowledgeBaseID)
		ref := planResolvedReferenceForTest(scope, resource.KindKnowledgeBase, name.String(), 7)
		out = append(out, workerdependencyartifact.KnowledgeBaseResolution{
			ResourceResolution: planResourceProjectionForTest(t, scope, ref, mount.KnowledgeBaseID),
			Slug:               name, HTTPCloneURL: "https://git.example.com/kb/" + name.String() + ".git",
			Branch: "main", CommitSHA: strings.Repeat("e", 40), Mode: mount.Mode,
		})
	}
	return out
}

func planKnowledgeSlugForArtifactTest(id int64) slugkit.Slug {
	switch id {
	case 101:
		return slugkit.MustNewForTest("team-docs")
	case 102:
		return slugkit.MustNewForTest("product-wiki")
	default:
		return slugkit.MustNewForTest("kb-" + formatPlanID(id))
	}
}

func planRuntimeBundlesForArtifactTest(
	t *testing.T,
	scope control.Scope,
	workspace specdomain.Workspace,
) []workerdependencyartifact.RuntimeBundleResolution {
	t.Helper()
	out := make([]workerdependencyartifact.RuntimeBundleResolution, 0, len(workspace.EnvBundleIDs)+len(workspace.ConfigDocumentBindings))
	for _, id := range workspace.EnvBundleIDs {
		name := "runtime"
		ref := planResolvedReferenceForTest(scope, resource.KindEnvironmentBundle, name, 8)
		values := []workerdependencyartifact.RuntimeValueResolution{
			{Name: "FEATURE_FLAG", Value: "enabled"},
			{Name: "TEST", Value: "1"},
		}
		out = append(out, workerdependencyartifact.RuntimeBundleResolution{
			ResourceResolution: planResourceProjectionForTest(t, scope, ref, int64(id)),
			Kind:               envbundle.KindRuntime,
			ContentDigest:      planRuntimeValuesDigestForTest(t, values),
			Values:             values,
		})
	}
	return out
}

func planRuntimeValuesDigestForTest(
	t *testing.T,
	values []workerdependencyartifact.RuntimeValueResolution,
) string {
	t.Helper()
	documentValues := make([]workerdependency.RuntimeValue, len(values))
	for index, value := range values {
		documentValues[index] = workerdependency.RuntimeValue{
			Name:  value.Name,
			Value: value.Value,
		}
	}
	digest, err := workerdependency.DigestRuntimeValues(documentValues)
	require.NoError(t, err)
	return digest
}

func planPlacementForArtifactTest(
	t *testing.T,
	scope control.Scope,
	spec specdomain.Spec,
) workerdependencyartifact.PlacementResolution {
	t.Helper()
	compute := planResolvedReferenceForTest(scope, resource.KindComputeTarget, "runner-pool", 9)
	profile := planResolvedReferenceForTest(scope, resource.KindResourceProfile, "standard", 10)
	profileResolution := planResourceProjectionForTest(t, scope, profile, spec.Placement.ResourceProfile.ID)
	return workerdependencyartifact.PlacementResolution{
		CatalogRevision: "test-options",
		RuntimeImageID:  spec.Runtime.Image.ID,
		ImageReference:  "registry.example.com/worker@" + spec.Runtime.Image.Digest,
		ImageDigest:     spec.Runtime.Image.Digest,
		ComputeTarget:   planResourceProjectionForTest(t, scope, compute, spec.Placement.ComputeTarget.ID),
		ResourceProfile: &profileResolution,
		Spec:            spec.Placement,
	}
}

func planReferencesForTest(
	deps workerdependencyartifact.ResolvedDependencies,
) []control.ResolvedReference {
	refs := []control.ResolvedReference{}
	if deps.PrimaryModel != nil {
		refs = append(refs, deps.PrimaryModel.ResolvedReference())
	}
	for _, model := range deps.ToolModels {
		refs = append(refs, model.Binding, model.Model.ResolvedReference())
	}
	if deps.Repository != nil {
		refs = append(refs, deps.Repository.ResolvedReference())
	}
	for _, skill := range deps.Skills {
		refs = append(refs, skill.ResolvedReference())
	}
	for _, kb := range deps.KnowledgeBases {
		refs = append(refs, kb.ResolvedReference())
	}
	for _, bundle := range deps.RuntimeBundles {
		refs = append(refs, bundle.ResolvedReference())
	}
	refs = append(refs, deps.Placement.ComputeTarget.ResolvedReference())
	if deps.Placement.ResourceProfile != nil {
		refs = append(refs, deps.Placement.ResourceProfile.ResolvedReference())
	}
	return refs
}

func planResourceProjectionForTest(
	t *testing.T,
	scope control.Scope,
	ref control.ResolvedReference,
	domainID int64,
) workerdependencyartifact.ResourceResolution {
	t.Helper()
	resolution, err := workerdependencyartifact.BindResourceProjection(scope, ref, domainID)
	require.NoError(t, err)
	return resolution
}

func planResolvedReferenceForTest(
	scope control.Scope,
	kind string,
	name string,
	revision int64,
) control.ResolvedReference {
	return control.ResolvedReference{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		Namespace: scope.OrganizationSlug,
		Name:      slugkit.MustNewForTest(name),
		UID:       uuid.NewString(),
		Revision:  revision,
		Digest:    workerdependency.TextDigest(kind + "/" + name),
	}
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intPtrValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func formatPlanID(id int64) string {
	if id < 0 {
		id = -id
	}
	if id == 0 {
		return "zero"
	}
	const digits = "0123456789"
	buf := [20]byte{}
	index := len(buf)
	for id > 0 {
		index--
		buf[index] = digits[id%10]
		id /= 10
	}
	return string(buf[index:])
}
