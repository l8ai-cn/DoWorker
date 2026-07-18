package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/organization"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

const (
	testResourceJSON = `{
		"apiVersion":"agentsmesh.io/v1alpha1",
		"kind":"WorkerTemplate",
		"metadata":{"name":"worker-one","namespace":"team-alpha"},
		"spec":{
			"promptRef":{"kind":"Prompt","name":"system-prompt"},
			"modelRef":{"kind":"ModelBinding","name":"coding-primary"}
		}
	}`
	testResourceYAML = `
kind: WorkerTemplate
metadata:
  namespace: team-alpha
  name: worker-one
spec:
  modelRef: {kind: ModelBinding, name: coding-primary}
  promptRef: {kind: Prompt, name: system-prompt}
apiVersion: agentsmesh.io/v1alpha1
`
)

type orchestrationTestSpec struct {
	ModelRef  orchestrationresource.Reference `json:"modelRef"`
	PromptRef orchestrationresource.Reference `json:"promptRef"`
}

type orchestrationServiceFixture struct {
	scope      orchestrationcontrol.Scope
	now        time.Time
	planID     string
	repository *orchestrationRepositoryStub
	authorizer *orchestrationAuthorizerStub
	references *orchestrationReferenceResolverStub
	planner    *orchestrationPlannerStub
	deps       ServiceDeps
}

func newOrchestrationServiceFixture(t *testing.T) *orchestrationServiceFixture {
	t.Helper()
	meta := orchestrationServiceTarget().TypeMeta
	registry := orchestrationresource.NewRegistry()
	require.NoError(t, registry.Register(meta, orchestrationresource.Schema{
		NewSpec: func() any { return &orchestrationTestSpec{} },
		Validate: func(metadata orchestrationresource.Metadata, value any) error {
			spec := value.(*orchestrationTestSpec)
			for _, ref := range []orchestrationresource.Reference{
				spec.ModelRef,
				spec.PromptRef,
			} {
				if err := ref.ValidateDraft(metadata.Namespace.String()); err != nil {
					return err
				}
			}
			return nil
		},
	}))
	repository := &orchestrationRepositoryStub{
		revisions: make(map[int64]orchestrationcontrol.ResourceRevision),
	}
	authorizer := &orchestrationAuthorizerStub{}
	references := &orchestrationReferenceResolverStub{
		errByKind: make(map[string]error),
	}
	planner := &orchestrationPlannerStub{
		meta: meta,
		output: TargetPlanOutput{
			ArtifactKind:    "WorkerSpec",
			ArtifactJSON:    json.RawMessage(`{"snapshotVersion":2}`),
			OptionsRevision: "runtime-catalog-3",
		},
	}
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	fixture := &orchestrationServiceFixture{
		scope: orchestrationServiceScope(), now: now,
		planID:     "11111111-1111-4111-8111-111111111111",
		repository: repository, authorizer: authorizer,
		references: references, planner: planner,
	}
	fixture.deps = ServiceDeps{
		Registry: registry, Repository: repository,
		Authorizer: authorizer, References: references,
		WorkerDefinitions: workerDefinitionPolicyStub{},
		Planners:          []TargetPlanner{planner}, RequiredTypes: []orchestrationresource.TypeMeta{meta},
		Clock: func() time.Time { return now },
		IDGenerator: func() string {
			return fixture.planID
		},
		PlanTTL: 15 * time.Minute,
	}
	return fixture
}

func (fixture *orchestrationServiceFixture) service(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(fixture.deps)
	require.NoError(t, err)
	return service
}

func orchestrationServiceScope() orchestrationcontrol.Scope {
	return orchestrationcontrol.Scope{
		OrganizationID: 42, OrganizationSlug: slugkit.MustNewForTest("team-alpha"),
		ActorID: 7,
	}
}

func orchestrationServiceTarget() orchestrationcontrol.ResourceTarget {
	return orchestrationcontrol.ResourceTarget{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		},
		Namespace: slugkit.MustNewForTest("team-alpha"),
		Name:      slugkit.MustNewForTest("worker-one"),
	}
}

func orchestrationServiceHead() orchestrationcontrol.ResourceHead {
	at := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	return orchestrationcontrol.ResourceHead{
		ID: 101, OrganizationID: 42,
		Identity: orchestrationcontrol.ResourceIdentity{
			ResourceTarget: orchestrationServiceTarget(),
			UID:            "22222222-2222-4222-8222-222222222222",
		},
		Labels: map[string]string{}, Status: json.RawMessage(`{}`),
		Revision: 2, Generation: 2, ResourceVersion: 3,
		CreatedByID: 7, UpdatedByID: 7, CreatedAt: at, UpdatedAt: at,
	}
}

func orchestrationServiceRevision(
	t *testing.T,
	head orchestrationcontrol.ResourceHead,
) orchestrationcontrol.ResourceRevision {
	t.Helper()
	manifest, err := orchestrationcontrol.CanonicalJSONObject(orchestrationresource.Manifest{
		TypeMeta: head.Identity.TypeMeta,
		Metadata: orchestrationresource.Metadata{
			Name: head.Identity.Name, Namespace: head.Identity.Namespace,
			UID: head.Identity.UID, ResourceVersion: "3", Generation: 2,
		},
		Spec: json.RawMessage(`{
			"modelRef":{"kind":"ModelBinding","name":"old-model"},
			"promptRef":{"kind":"Prompt","name":"system-prompt"}
		}`),
		Status: json.RawMessage(`{}`),
	})
	require.NoError(t, err)
	spec, err := orchestrationcontrol.CanonicalJSONObject(json.RawMessage(`{
		"modelRef":{"kind":"ModelBinding","name":"old-model"},
		"promptRef":{"kind":"Prompt","name":"system-prompt"}
	}`))
	require.NoError(t, err)
	digest, err := orchestrationcontrol.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	return orchestrationcontrol.ResourceRevision{
		OrganizationID: 42, ResourceID: head.ID, Identity: head.Identity,
		Revision: head.Revision, Generation: head.Generation,
		ResourceVersion: head.ResourceVersion, CanonicalManifest: manifest,
		CanonicalSpec: spec, ResolvedReferences: []orchestrationcontrol.ResolvedReference{},
		Digest: digest, ActorID: 7, CreatedAt: head.UpdatedAt,
	}
}

type resourceRead struct {
	head orchestrationcontrol.ResourceHead
	err  error
}

type orchestrationRepositoryStub struct {
	resourceSequence []resourceRead
	revisions        map[int64]orchestrationcontrol.ResourceRevision
	plan             orchestrationcontrol.Plan
	planErr          error
	createdPlans     []orchestrationcontrol.Plan
	createPlanCalls  int
}

func (stub *orchestrationRepositoryStub) GetResource(
	context.Context,
	orchestrationcontrol.Scope,
	orchestrationcontrol.ResourceTarget,
) (orchestrationcontrol.ResourceHead, error) {
	if len(stub.resourceSequence) == 0 {
		return orchestrationcontrol.ResourceHead{}, orchestrationcontrol.ErrNotFound
	}
	read := stub.resourceSequence[0]
	stub.resourceSequence = stub.resourceSequence[1:]
	return read.head, read.err
}

func (*orchestrationRepositoryStub) ListResources(
	context.Context,
	orchestrationcontrol.Scope,
	ResourceListFilter,
) (ResourceListPage, error) {
	return ResourceListPage{}, nil
}

func (stub *orchestrationRepositoryStub) GetRevision(
	_ context.Context,
	_ orchestrationcontrol.Scope,
	_ int64,
	revision int64,
) (orchestrationcontrol.ResourceRevision, error) {
	value, exists := stub.revisions[revision]
	if !exists {
		return orchestrationcontrol.ResourceRevision{}, orchestrationcontrol.ErrNotFound
	}
	return value, nil
}

func (*orchestrationRepositoryStub) ListRevisions(
	context.Context,
	orchestrationcontrol.Scope,
	int64,
	int,
	int,
) ([]orchestrationcontrol.ResourceRevision, error) {
	return nil, nil
}

func (stub *orchestrationRepositoryStub) CreatePlan(
	_ context.Context,
	plan orchestrationcontrol.Plan,
) error {
	stub.createPlanCalls++
	stub.createdPlans = append(stub.createdPlans, plan)
	return nil
}

func (stub *orchestrationRepositoryStub) GetPlan(
	_ context.Context,
	_ orchestrationcontrol.Scope,
	_ string,
) (orchestrationcontrol.Plan, error) {
	if stub.planErr != nil {
		return orchestrationcontrol.Plan{}, stub.planErr
	}
	if stub.plan.ID == "" {
		return orchestrationcontrol.Plan{}, orchestrationcontrol.ErrNotFound
	}
	return stub.plan, nil
}

func (*orchestrationRepositoryStub) RunApplyTransaction(
	context.Context,
	orchestrationcontrol.Scope,
	string,
	ApplyBuilder,
) (orchestrationcontrol.ResourceHead, error) {
	return orchestrationcontrol.ResourceHead{}, errors.New("not implemented")
}

type orchestrationAuthorizerStub struct {
	createErr      error
	updateErr      error
	referenceErr   error
	listCalls      int
	createCalls    int
	updateCalls    int
	referenceCalls int
}

func (stub *orchestrationAuthorizerStub) AuthorizeList(
	context.Context,
	orchestrationcontrol.Scope,
) error {
	stub.listCalls++
	return nil
}

func (stub *orchestrationAuthorizerStub) AuthorizeCreate(
	context.Context,
	orchestrationcontrol.Scope,
	orchestrationcontrol.ResourceTarget,
) error {
	stub.createCalls++
	return stub.createErr
}

func (stub *orchestrationAuthorizerStub) AuthorizeUpdate(
	context.Context,
	orchestrationcontrol.Scope,
	orchestrationcontrol.ResourceHead,
) error {
	stub.updateCalls++
	return stub.updateErr
}

func (stub *orchestrationAuthorizerStub) AuthorizeReference(
	context.Context,
	orchestrationcontrol.Scope,
	orchestrationcontrol.ResourceHead,
) error {
	stub.referenceCalls++
	return stub.referenceErr
}

type orchestrationReferenceResolverStub struct {
	calls     int
	kinds     []string
	errByKind map[string]error
}

func (stub *orchestrationReferenceResolverStub) Resolve(
	_ context.Context,
	scope orchestrationcontrol.Scope,
	request DraftReference,
) (orchestrationcontrol.ResolvedReference, error) {
	stub.calls++
	stub.kinds = append(stub.kinds, request.Reference.Kind)
	if err := stub.errByKind[request.Reference.Kind]; err != nil {
		return orchestrationcontrol.ResolvedReference{}, err
	}
	revision := request.Reference.Revision
	if revision == 0 {
		revision = 4
	}
	return orchestrationcontrol.ResolvedReference{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       request.Reference.Kind,
		},
		Namespace: scope.OrganizationSlug,
		Name:      request.Reference.Name,
		UID: strings.Repeat(map[string]string{
			"ModelBinding": "3",
			"Prompt":       "4",
		}[request.Reference.Kind], 8) + "-3333-4333-8333-333333333333",
		Revision: revision,
		Digest: "sha256:" + strings.Repeat(map[string]string{
			"ModelBinding": "a",
			"Prompt":       "b",
		}[request.Reference.Kind], 64),
	}, nil
}

type orchestrationPlannerStub struct {
	meta           orchestrationresource.TypeMeta
	output         TargetPlanOutput
	planErr        error
	referenceCalls int
	planCalls      int
}

func (stub *orchestrationPlannerStub) TypeMeta() orchestrationresource.TypeMeta {
	return stub.meta
}

func (stub *orchestrationPlannerStub) References(value any) ([]DraftReference, error) {
	stub.referenceCalls++
	spec := value.(*orchestrationTestSpec)
	return []DraftReference{
		{Path: "/spec/promptRef", Reference: spec.PromptRef},
		{Path: "/spec/modelRef", Reference: spec.ModelRef},
	}, nil
}

func (stub *orchestrationPlannerStub) Plan(
	context.Context,
	TargetPlanInput,
) (TargetPlanOutput, error) {
	stub.planCalls++
	return stub.output, stub.planErr
}

type memberReaderStub struct {
	member *organization.Member
	err    error
}

func (stub *memberReaderStub) GetMember(
	context.Context,
	int64,
	int64,
) (*organization.Member, error) {
	return stub.member, stub.err
}
