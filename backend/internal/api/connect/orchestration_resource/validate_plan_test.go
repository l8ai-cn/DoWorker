package orchestrationresourceconnect

import (
	"encoding/json"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	resourcev1 "github.com/l8ai-cn/agentcloud/proto/gen/go/orchestration_resource/v1"
)

func TestValidateResourceConvertsJSONAndYAMLWithResolvedActorScope(t *testing.T) {
	tests := []struct {
		name     string
		format   resourcev1.SourceFormat
		expected service.SourceFormat
		content  []byte
	}{
		{
			name:     "json",
			format:   resourcev1.SourceFormat_SOURCE_FORMAT_JSON,
			expected: service.SourceFormatJSON,
			content:  []byte(`{"apiVersion":"agentcloud.io/v1alpha1"}`),
		},
		{
			name:     "yaml",
			format:   resourcev1.SourceFormat_SOURCE_FORMAT_YAML,
			expected: service.SourceFormatYAML,
			content:  []byte("apiVersion: agentcloud.io/v1alpha1\n"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &serviceStub{validateResult: service.ValidationResult{
				Target:            testTarget(),
				Operation:         control.PlanOperationUpdate,
				CanonicalManifest: json.RawMessage(`{"canonical":true}`),
				Issues: []control.PlanIssue{{
					Severity: control.PlanIssueWarning,
					Path:     "/spec",
					Code:     "review.required",
					Message:  "Review required.",
				}},
			}}
			server := newTestServer(stub, testOrganizations())

			response, err := server.ValidateResource(
				authenticatedContext(42),
				connect.NewRequest(&resourcev1.ValidateResourceRequest{
					OrgSlug: "lookup-alias",
					Source: &resourcev1.ResourceSource{
						Format:  test.format,
						Content: test.content,
					},
				}),
			)

			require.NoError(t, err)
			assert.Equal(t, testScope(), stub.validateRequest.Scope)
			assert.Equal(t, test.expected, stub.validateRequest.Source.Format)
			assert.Equal(t, test.content, stub.validateRequest.Source.Content)
			assert.Equal(t, resourcev1.ResourceOperation_RESOURCE_OPERATION_UPDATE, response.Msg.Operation)
			assert.Equal(t, []byte(`{"canonical":true}`), response.Msg.CanonicalJson)
			require.Len(t, response.Msg.Issues, 1)
			assert.Equal(t, resourcev1.IssueSeverity_ISSUE_SEVERITY_WARNING, response.Msg.Issues[0].Severity)
		})
	}
}

func TestPlanResourceWithBlockingIssueReturnsNoPlan(t *testing.T) {
	plan := testPlan()
	stub := &serviceStub{planResult: service.PlanResult{
		ValidationResult: service.ValidationResult{
			Target:    testTarget(),
			Operation: control.PlanOperationCreate,
			Issues: []control.PlanIssue{{
				Severity: control.PlanIssueBlocking,
				Path:     "/spec",
				Code:     "invalid-draft",
				Message:  "The resource draft is invalid.",
			}},
		},
		Plan: &plan,
	}}
	server := newTestServer(stub, testOrganizations())

	response, err := server.PlanResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.PlanResourceRequest{
			OrgSlug: "acme",
			Source: &resourcev1.ResourceSource{
				Format:  resourcev1.SourceFormat_SOURCE_FORMAT_JSON,
				Content: []byte(`{}`),
			},
		}),
	)

	require.NoError(t, err)
	assert.Nil(t, response.Msg.Plan)
	require.Len(t, response.Msg.Issues, 1)
	assert.Equal(t, resourcev1.IssueSeverity_ISSUE_SEVERITY_BLOCKING, response.Msg.Issues[0].Severity)
}

func TestPlanResourceConvertsCompletePlanWithoutPrivatePayloads(t *testing.T) {
	plan := testPlan()
	stub := &serviceStub{planResult: service.PlanResult{
		ValidationResult: service.ValidationResult{
			Target:            testTarget(),
			Operation:         control.PlanOperationUpdate,
			CanonicalManifest: json.RawMessage(`{"canonical":true}`),
			Issues:            plan.Issues,
		},
		Plan: &plan,
	}}
	server := newTestServer(stub, testOrganizations())

	response, err := server.PlanResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.PlanResourceRequest{
			OrgSlug: "acme",
			Source: &resourcev1.ResourceSource{
				Format:  resourcev1.SourceFormat_SOURCE_FORMAT_JSON,
				Content: []byte(`{}`),
			},
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, response.Msg.Plan)
	assert.Equal(t, testPlanID, response.Msg.Plan.PlanId)
	assert.Equal(t, resourcev1.ResourceOperation_RESOURCE_OPERATION_UPDATE, response.Msg.Plan.Operation)
	assert.Equal(t, testResourceID, response.Msg.Plan.Base.Uid)
	assert.EqualValues(t, 7, response.Msg.Plan.BaseResourceVersion)
	require.Len(t, response.Msg.Plan.ResolvedReferences, 1)
	assert.Equal(t, "review-prompt", response.Msg.Plan.ResolvedReferences[0].Name)
	require.Len(t, response.Msg.Plan.SemanticDiff, 1)
	assert.Equal(
		t,
		resourcev1.SemanticChangeOperation_SEMANTIC_CHANGE_OPERATION_REPLACE,
		response.Msg.Plan.SemanticDiff[0].Operation,
	)
	assert.Equal(t, plan.SemanticChanges[0].Before.Digest, response.Msg.Plan.SemanticDiff[0].Before.GetDigest())
	assert.Equal(t, []byte(`{"value":"redacted"}`), response.Msg.Plan.SemanticDiff[0].After.GetRedactedJson())
	assert.Equal(t, resourcev1.PlanStatus_PLAN_STATUS_PENDING, response.Msg.Plan.Status)
	assert.Equal(t, "2026-07-14T00:30:00Z", response.Msg.Plan.CreatedAt)
	assert.Equal(t, "2026-07-14T00:45:00Z", response.Msg.Plan.ExpiresAt)

	encoded, err := protojson.Marshal(response.Msg.Plan)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "canonical-payload")
	assert.NotContains(t, string(encoded), "artifact-payload")
	assert.NotContains(t, string(encoded), "artifactJson")
	assert.NotContains(t, string(encoded), "canonicalManifest")
}

func TestValidateResourceRejectsUnsupportedSourceFormat(t *testing.T) {
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	_, err := server.ValidateResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ValidateResourceRequest{
			OrgSlug: "acme",
			Source: &resourcev1.ResourceSource{
				Format:  resourcev1.SourceFormat_SOURCE_FORMAT_UNSPECIFIED,
				Content: []byte(`{}`),
			},
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Zero(t, stub.validateCalls)
}

func TestValidateResourceRejectsCrossOrganizationMember(t *testing.T) {
	stub := &serviceStub{}
	organizations := testOrganizations()
	organizations.member = false
	organizations.roleErr = assert.AnError
	server := newTestServer(stub, organizations)

	_, err := server.ValidateResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ValidateResourceRequest{
			OrgSlug: "other-org",
			Source: &resourcev1.ResourceSource{
				Format:  resourcev1.SourceFormat_SOURCE_FORMAT_JSON,
				Content: []byte(`{}`),
			},
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	assert.Zero(t, stub.validateCalls)
}
