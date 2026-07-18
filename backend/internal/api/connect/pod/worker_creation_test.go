package podconnect

import (
	"context"
	"encoding/json"
	"testing"

	"connectrpc.com/connect"
	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerDraftProtoRoundTripPreservesAllFields(t *testing.T) {
	input := completeWorkerDraftProto()

	draft, err := workerDraftFromProto(input)
	require.NoError(t, err)
	output, err := workerDraftToProto(draft)
	require.NoError(t, err)
	roundTrip, err := workerDraftFromProto(output)
	require.NoError(t, err)

	assert.Equal(t, draft, roundTrip)
	assert.Equal(t, json.Number("0.2"), draft.WorkerSpec.TypeConfig.Values["temperature"])
}

func TestListWorkerCreateOptionsReturnsUnavailableWithoutService(t *testing.T) {
	server := NewServer(nil, &fakeOrgService{role: "admin"})

	_, err := server.ListWorkerCreateOptions(
		ctxAsUser(42),
		connect.NewRequest(&podv1.ListWorkerCreateOptionsRequest{OrgSlug: "acme"}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
}

func TestListWorkerCreateOptionsMapsServiceResult(t *testing.T) {
	service := &fakeWorkerCreationAPI{
		options: workercreation.CreateOptions{
			Revision: "runtime-catalog-1",
			WorkerTypes: []workercreation.WorkerTypeOption{{
				Slug:        "codex-cli",
				Name:        "Codex CLI",
				Description: "Codex worker",
				Schema: specdomain.TypeSchema{
					Version: 1,
					Fields: map[string]specdomain.TypeFieldSchema{
						"approval_mode": {
							Kind:    specdomain.TypeFieldSelect,
							Options: []string{"never"},
						},
					},
					SecretRequirementGroups: []specdomain.SecretRequirementGroup{{
						ID:    "provider-api-key",
						AnyOf: []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY"},
					}},
				},
				RequiresModelResource: true,
				ModelProtocolAdapters: []string{
					"openai-compatible",
					"anthropic",
				},
				Selectable: true,
			}},
			RuntimeImages: []workercreation.RuntimeImageOption{{
				Image: runtimedomain.CatalogRuntimeImage{
					ID:              1,
					Slug:            "codex-cli-stable",
					Name:            "Codex CLI",
					Reference:       "repo/image@sha256:abc",
					Digest:          "sha256:abc",
					WorkerTypeSlugs: []string{"codex-cli"},
				},
				Selectable: true,
			}},
			ComputeTargets: []workercreation.ComputeTargetOption{{
				Target: runtimedomain.CatalogComputeTarget{
					ID:             1,
					Slug:           "organization-runner-pool",
					Name:           "Organization runner pool",
					Kind:           specdomain.ComputeTargetKindRunnerPool,
					SupportsPooled: true,
				},
				Selectable: true,
			}},
			DeploymentModes: []workercreation.DeploymentModeOption{{
				Value:      specdomain.DeploymentModePooled,
				Name:       "Pooled",
				Selectable: true,
			}},
			ResourceProfiles: []workercreation.ResourceProfileOption{{
				Profile: runtimedomain.CatalogResourceProfile{
					ID:   1,
					Slug: "standard",
					Name: "Standard",
					Resources: specdomain.ResourceRequestsLimits{
						CPURequestMilliCPU: 200,
						CPULimitMilliCPU:   1000,
						MemoryRequestBytes: 256,
						MemoryLimitBytes:   1024,
					},
				},
				Selectable: true,
			}},
		},
	}
	server := NewServer(
		nil,
		&fakeOrgService{role: "admin"},
		WithWorkerCreation(service),
	)
	targetID := int64(1)

	response, err := server.ListWorkerCreateOptions(
		ctxAsUser(42),
		connect.NewRequest(&podv1.ListWorkerCreateOptionsRequest{
			OrgSlug:         "acme",
			WorkerTypeSlug:  stringPtr("codex-cli"),
			ComputeTargetId: &targetID,
			DeploymentMode:  stringPtr("pooled"),
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, "runtime-catalog-1", response.Msg.Revision)
	require.Len(t, response.Msg.WorkerTypes, 1)
	assert.Equal(t, "codex-cli", response.Msg.WorkerTypes[0].Slug)
	assert.True(t, response.Msg.WorkerTypes[0].RequiresModelResource)
	assert.Equal(
		t,
		[]string{"openai-compatible", "anthropic"},
		response.Msg.WorkerTypes[0].ModelProtocolAdapters,
	)
	assert.JSONEq(t, `{"version":1,"fields":{"approval_mode":{"kind":"select","options":["never"]}},"credential_requirement_groups":[{"id":"provider-api-key","any_of":["OPENAI_API_KEY","ANTHROPIC_API_KEY"]}]}`, response.Msg.WorkerTypes[0].ConfigSchemaJson)
	require.Len(t, response.Msg.RuntimeImages, 1)
	assert.Equal(t, int64(1), response.Msg.RuntimeImages[0].Id)
	require.Len(t, response.Msg.ResourceProfiles, 1)
	assert.Equal(t, uint32(200), response.Msg.ResourceProfiles[0].CpuRequestMillicpu)
	assert.Equal(t, specservice.Scope{OrgID: 7, UserID: 42}, service.optionsScope)
	assert.Equal(t, workercreation.OptionsFilter{
		WorkerTypeSlug:  "codex-cli",
		ComputeTargetID: &targetID,
		DeploymentMode:  specdomain.DeploymentModePooled,
	}, service.optionsFilter)
}

func TestPreflightWorkerDecodesDraftAndMapsIssues(t *testing.T) {
	service := &fakeWorkerCreationAPI{
		preflight: workercreation.PreflightResult{
			BlockingErrors: []workercreation.Issue{{
				Code: "invalid-draft", Field: "worker_spec.workspace.branch",
				Message: "branch is required", Severity: "blocking",
			}},
			Warnings: []workercreation.Issue{{
				Code: "large-profile", Field: "worker_spec.resource_profile_id",
				Message: "large profile", Severity: "warning",
			}},
			OptionsRevision: "runtime-catalog-1",
		},
	}
	server := NewServer(
		nil,
		&fakeOrgService{role: "admin"},
		WithWorkerCreation(service),
	)

	response, err := server.PreflightWorker(
		ctxAsUser(42),
		connect.NewRequest(&podv1.PreflightWorkerRequest{
			OrgSlug: "acme",
			Draft:   completeWorkerDraftProto(),
		}),
	)

	require.NoError(t, err)
	require.Len(t, response.Msg.Issues, 2)
	assert.Equal(t, "blocking", response.Msg.Issues[0].Severity)
	assert.Equal(t, "warning", response.Msg.Issues[1].Severity)
	assert.Equal(t, "runtime-catalog-1", response.Msg.OptionsRevision)
	assert.Equal(t, specservice.Scope{OrgID: 7, UserID: 42}, service.preflightScope)
	assert.Equal(t, int64(101), service.preflightDraft.WorkerSpec.ModelResourceID)
	assert.Equal(t, []int64{3, 4}, service.preflightDraft.WorkerSpec.Workspace.SkillIDs)
}

func TestPreflightWorkerRejectsMalformedConfigJSON(t *testing.T) {
	service := &fakeWorkerCreationAPI{}
	server := NewServer(
		nil,
		&fakeOrgService{role: "admin"},
		WithWorkerCreation(service),
	)
	draft := completeWorkerDraftProto()
	draft.TypeConfigValuesJson = `{"temperature":`

	_, err := server.PreflightWorker(
		ctxAsUser(42),
		connect.NewRequest(&podv1.PreflightWorkerRequest{
			OrgSlug: "acme",
			Draft:   draft,
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
	assert.Zero(t, service.preflightCalls)
}

func TestFillWorkerDraftUsesInjectedFiller(t *testing.T) {
	current, err := workerDraftFromProto(completeWorkerDraftProto())
	require.NoError(t, err)
	filler := &fakeWorkerDraftFiller{
		result: workercreation.FillResult{
			Draft: current,
			Issues: []workercreation.Issue{{
				Code: "clarify-task", Field: "worker_spec.initial_task",
				Message: "Task can be more specific", Severity: "warning",
			}},
		},
	}
	server := NewServer(
		nil,
		&fakeOrgService{role: "admin"},
		WithWorkerDraftFiller(filler),
	)

	response, err := server.FillWorkerDraft(
		ctxAsUser(42),
		connect.NewRequest(&podv1.FillWorkerDraftRequest{
			OrgSlug:      "acme",
			Prompt:       "Build a review worker",
			CurrentDraft: completeWorkerDraftProto(),
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, response.Msg.Draft)
	assert.Equal(t, int64(101), response.Msg.Draft.ModelResourceId)
	require.Len(t, response.Msg.Issues, 1)
	assert.Equal(t, "warning", response.Msg.Issues[0].Severity)
	assert.Equal(t, "Build a review worker", filler.prompt)
	assert.Equal(t, specservice.Scope{OrgID: 7, UserID: 42}, filler.scope)
	require.NotNil(t, filler.current)
}

func completeWorkerDraftProto() *podv1.WorkerSpecDraft {
	repositoryID := int64(22)
	sourceExpertID := int64(31)
	return &podv1.WorkerSpecDraft{
		ModelResourceId:      101,
		WorkerTypeSlug:       "codex-cli",
		RuntimeImageId:       1,
		PlacementPolicy:      "explicit",
		ComputeTargetId:      1,
		DeploymentMode:       "pooled",
		ResourceProfileId:    1,
		TypeSchemaVersion:    1,
		TypeConfigValuesJson: `{"approval_mode":"never","temperature":0.2}`,
		SecretRefs: []*podv1.WorkerSecretReference{{
			Field: "SIGNING_KEY", Kind: "env-bundle", Id: 6,
		}},
		InteractionMode: "acp",
		AutomationLevel: "autonomous",
		RepositoryId:    &repositoryID,
		Branch:          "main",
		SkillIds:        []int64{3, 4},
		KnowledgeMounts: []*podv1.WorkerKnowledgeMount{{KnowledgeBaseId: 5, Mode: "rw"}},
		EnvBundleIds:    []int64{7, 8},
		ConfigDocumentBindings: []*podv1.WorkerConfigDocumentBinding{{
			DocumentId: "settings", ConfigBundleId: 9,
		}},
		Instructions:       "Review before editing.",
		InitialTask:        "Fix the failing test.",
		TerminationPolicy:  "idle",
		IdleTimeoutMinutes: 30,
		Alias:              "review-worker",
		SourceExpertId:     &sourceExpertID,
		OptionsRevision:    "runtime-catalog-1",
	}
}

type fakeWorkerCreationAPI struct {
	options        workercreation.CreateOptions
	optionsErr     error
	optionsScope   specservice.Scope
	optionsFilter  workercreation.OptionsFilter
	preflight      workercreation.PreflightResult
	preflightErr   error
	preflightScope specservice.Scope
	preflightDraft workercreation.Draft
	preflightCalls int
}

func (service *fakeWorkerCreationAPI) ListOptions(
	_ context.Context,
	scope specservice.Scope,
	filter workercreation.OptionsFilter,
) (workercreation.CreateOptions, error) {
	service.optionsScope = scope
	service.optionsFilter = filter
	return service.options, service.optionsErr
}

func (service *fakeWorkerCreationAPI) Preflight(
	_ context.Context,
	scope specservice.Scope,
	draft workercreation.Draft,
) (workercreation.PreflightResult, error) {
	service.preflightCalls++
	service.preflightScope = scope
	service.preflightDraft = draft
	return service.preflight, service.preflightErr
}

type fakeWorkerDraftFiller struct {
	result  workercreation.FillResult
	err     error
	scope   specservice.Scope
	prompt  string
	current *workercreation.Draft
}

func (filler *fakeWorkerDraftFiller) Fill(
	_ context.Context,
	scope specservice.Scope,
	prompt string,
	current *workercreation.Draft,
) (workercreation.FillResult, error) {
	filler.scope = scope
	filler.prompt = prompt
	filler.current = current
	return filler.result, filler.err
}

func stringPtr(value string) *string {
	return &value
}
