package orchestrationworker

import (
	"context"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceBindingResolverReturnsPinnedEntityIDs(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	cases := []struct {
		kind string
		name string
		spec any
		want int64
	}{
		{resource.KindModelBinding, "primary-model", resource.ModelBindingSpec{ResourceID: 101}, 101},
		{resource.KindRepository, "source-repository", resource.RepositoryBindingSpec{RepositoryID: 102}, 102},
		{resource.KindSkill, "review-skill", resource.SkillBindingSpec{SkillID: 103}, 103},
		{resource.KindKnowledgeBase, "engineering-docs", resource.KnowledgeBaseBindingSpec{KnowledgeBaseID: 104}, 104},
		{resource.KindEnvironmentBundle, "runtime-env", resource.EnvironmentBundleBindingSpec{EnvironmentBundleID: 105}, 105},
		{resource.KindComputeTarget, "primary-pool", resource.ComputeTargetBindingSpec{ComputeTargetID: 106}, 106},
		{resource.KindResourceProfile, "balanced-profile", resource.ResourceProfileBindingSpec{ResourceProfileID: 107}, 107},
	}

	for _, test := range cases {
		t.Run(test.kind, func(t *testing.T) {
			pinned := fixture.addBinding(t, test.kind, test.name, test.spec, nil)

			id, err := fixture.resolver.ResolveEntityID(
				context.Background(),
				fixture.scope,
				pinned,
			)

			require.NoError(t, err)
			assert.Equal(t, test.want, id)
			assert.Equal(t, int64(1), fixture.repository.lastRevision)
		})
	}
}

func TestResourceBindingResolverAuthorizesBeforeReadingRevision(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	pinned := fixture.addBinding(
		t,
		resource.KindModelBinding,
		"private-model",
		resource.ModelBindingSpec{ResourceID: 101},
		nil,
	)
	fixture.authorizer.err = controlservice.ErrForbidden

	_, err := fixture.resolver.ResolveEntityID(
		context.Background(),
		fixture.scope,
		pinned,
	)

	assert.ErrorIs(t, err, controlservice.ErrForbidden)
	assert.Zero(t, fixture.repository.revisionCalls)
}

func TestResourceBindingResolverRejectsSimpleBindingWithNestedReferences(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	nested := resolvedBindingReference(
		fixture.scope,
		resource.KindModelBinding,
		"nested-model",
	)
	pinned := fixture.addBinding(
		t,
		resource.KindRepository,
		"source-repository",
		resource.RepositoryBindingSpec{RepositoryID: 102},
		[]control.ResolvedReference{nested},
	)

	_, err := fixture.resolver.ResolveEntityID(
		context.Background(),
		fixture.scope,
		pinned,
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestResourceBindingResolverRejectsTamperedPinnedDigest(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	pinned := fixture.addBinding(
		t,
		resource.KindModelBinding,
		"primary-model",
		resource.ModelBindingSpec{ResourceID: 101},
		nil,
	)
	pinned.Digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	_, err := fixture.resolver.ResolveEntityID(
		context.Background(),
		fixture.scope,
		pinned,
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestResourceBindingResolverUsesToolRevisionModelPin(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	modelPin := fixture.addBinding(
		t,
		resource.KindModelBinding,
		"tool-model",
		resource.ModelBindingSpec{ResourceID: 501},
		nil,
	)
	toolPin := fixture.addBinding(
		t,
		resource.KindToolBinding,
		"web-search",
		resource.ToolBindingSpec{ModelRef: resource.Reference{
			Kind: resource.KindModelBinding,
			Name: modelPin.Name,
		}},
		[]control.ResolvedReference{modelPin},
	)

	id, err := fixture.resolver.ResolveToolModelResourceID(
		context.Background(),
		fixture.scope,
		toolPin,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(501), id)
	assert.Equal(t, 2, fixture.authorizer.referenceCalls)
	assert.Equal(t, 2, fixture.repository.revisionCalls)
}

func TestResourceBindingResolverRejectsToolWithoutOwnModelPin(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	toolPin := fixture.addBinding(
		t,
		resource.KindToolBinding,
		"web-search",
		resource.ToolBindingSpec{ModelRef: resource.Reference{
			Kind: resource.KindModelBinding,
			Name: resourceBindingName("missing-model"),
		}},
		nil,
	)

	_, err := fixture.resolver.ResolveToolModelResourceID(
		context.Background(),
		fixture.scope,
		toolPin,
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestNewResourceBindingResolverRequiresRegisteredSchemas(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)

	_, err := NewResourceBindingResolver(
		resource.NewRegistry(),
		fixture.repository,
		fixture.authorizer,
	)

	assert.ErrorIs(t, err, controlservice.ErrUnavailable)
}

func TestResourceBindingResolverReturnsPinnedWorkerTemplateSnapshot(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	pinned := fixture.addBinding(
		t,
		resource.KindWorkerTemplate,
		"reviewer",
		workerTemplateSpecForTest(),
		nil,
	)
	fixture.setSnapshotID(pinned, 901)

	snapshotID, err := fixture.resolver.ResolveWorkerSpecSnapshotID(
		context.Background(),
		fixture.scope,
		pinned,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(901), snapshotID)
	assert.Equal(t, pinned.Revision, fixture.repository.lastRevision)
}

func TestResourceBindingResolverReturnsPinnedPromptSpec(t *testing.T) {
	fixture := newResourceBindingResolverFixture(t)
	defaultValue := "main"
	pinned := fixture.addBinding(
		t,
		resource.KindPrompt,
		"review-task",
		resource.PromptSpec{
			Content: "Review {{branch}}",
			Variables: map[string]resource.PromptVariableSpec{
				"branch": {Required: true, Default: &defaultValue},
			},
		},
		nil,
	)

	prompt, err := fixture.resolver.ResolvePromptSpec(
		context.Background(),
		fixture.scope,
		pinned,
	)

	require.NoError(t, err)
	assert.Equal(t, "Review {{branch}}", prompt.Content)
	require.NotNil(t, prompt.Variables["branch"].Default)
	assert.Equal(t, "main", *prompt.Variables["branch"].Default)
}
