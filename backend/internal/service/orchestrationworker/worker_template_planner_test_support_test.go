package orchestrationworker

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
)

type bindingResolverStub struct {
	ids   map[string]int64
	calls int
}

func newBindingResolverStub() *bindingResolverStub {
	return &bindingResolverStub{ids: map[string]int64{
		resource.KindModelBinding + "/coding-primary":           101,
		resource.KindToolBinding + "/video-tool":                102,
		resource.KindComputeTarget + "/primary-pool":            103,
		resource.KindResourceProfile + "/balanced-profile":      104,
		resource.KindEnvironmentBundle + "/secret-bundle":       105,
		resource.KindRepository + "/agents-mesh":                106,
		resource.KindSkill + "/review-skill":                    107,
		resource.KindKnowledgeBase + "/engineering-docs":        108,
		resource.KindEnvironmentBundle + "/runtime-environment": 109,
		resource.KindEnvironmentBundle + "/config-environment":  110,
	}}
}

func (stub *bindingResolverStub) ResolveEntityID(
	_ context.Context,
	_ control.Scope,
	reference control.ResolvedReference,
) (int64, error) {
	stub.calls++
	id := stub.ids[reference.Kind+"/"+reference.Name.String()]
	if id <= 0 {
		return 0, fmt.Errorf("binding not found")
	}
	return id, nil
}

func (stub *bindingResolverStub) ResolveToolModelResourceID(
	_ context.Context,
	_ control.Scope,
	reference control.ResolvedReference,
) (int64, error) {
	stub.calls++
	id := stub.ids[reference.Kind+"/"+reference.Name.String()]
	if id <= 0 {
		return 0, fmt.Errorf("tool binding not found")
	}
	return id, nil
}

type workerCompilerStub struct {
	revision string
	artifact []byte
	issues   []control.PlanIssue
	draft    workercreation.Draft
	calls    int
}

func (stub *workerCompilerStub) Revision() string { return stub.revision }

func (stub *workerCompilerStub) Compile(
	_ context.Context,
	_ control.Scope,
	draft workercreation.Draft,
) (WorkerCompilation, error) {
	stub.calls++
	stub.draft = draft
	return WorkerCompilation{
		ArtifactJSON: stub.artifact,
		Issues:       stub.issues,
	}, nil
}

func workerTemplatePlannerForTest(t *testing.T) *WorkerTemplatePlanner {
	t.Helper()
	planner, err := NewWorkerTemplatePlanner(
		newBindingResolverStub(),
		&workerCompilerStub{
			revision: "runtime-catalog-7",
			artifact: []byte(`{"version":1}`),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return planner
}

func resolvedWorkerTemplateReferences(
	t *testing.T,
	planner *WorkerTemplatePlanner,
	spec resource.WorkerTemplateSpec,
) []control.ResolvedReference {
	t.Helper()
	drafts, err := planner.References(&spec)
	if err != nil {
		t.Fatal(err)
	}
	byIdentity := make(map[string]control.ResolvedReference, len(drafts))
	for index, draft := range drafts {
		apiVersion := draft.Reference.APIVersion
		if apiVersion == "" {
			apiVersion = resource.APIVersionV1Alpha1
		}
		namespace := draft.Reference.Namespace
		if namespace == "" {
			namespace = workerTemplateScope().OrganizationSlug
		}
		key := strings.Join([]string{
			apiVersion, draft.Reference.Kind, namespace.String(),
			draft.Reference.Name.String(),
		}, "/")
		digestCharacter := string(rune('a' + index%6))
		byIdentity[key] = control.ResolvedReference{
			TypeMeta: resource.TypeMeta{
				APIVersion: apiVersion, Kind: draft.Reference.Kind,
			},
			Namespace: namespace, Name: draft.Reference.Name,
			UID:      fmt.Sprintf("00000000-0000-4000-8000-%012d", index+1),
			Revision: 1, Digest: "sha256:" + strings.Repeat(digestCharacter, 64),
		}
	}
	resolved := make([]control.ResolvedReference, 0, len(byIdentity))
	for _, reference := range byIdentity {
		resolved = append(resolved, reference)
	}
	sort.Slice(resolved, func(left, right int) bool {
		return resolved[left].Kind+"/"+resolved[left].Name.String() <
			resolved[right].Kind+"/"+resolved[right].Name.String()
	})
	return resolved
}

func referencePaths(references []controlservice.DraftReference) []string {
	paths := make([]string, len(references))
	for index := range references {
		paths[index] = references[index].Path
	}
	sort.Strings(paths)
	return paths
}
