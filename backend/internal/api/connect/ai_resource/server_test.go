package airesourceconnect

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	service "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	aiv1 "github.com/anthropics/agentsmesh/proto/gen/go/ai_resource/v1"
)

type fakeService struct {
	Service
	catalog      []domain.ProviderDefinition
	connections  []service.ConnectionView
	effective    []service.EffectiveResourceView
	createInput  service.CreateConnectionInput
	createActor  service.Actor
	listScope    domain.OwnerScope
	listOwnerID  int64
	effectiveOrg int64
	validationID int64
	err          error
}

func (f *fakeService) Catalog() []domain.ProviderDefinition { return f.catalog }

func (f *fakeService) ListOwnerConnections(_ context.Context, _ service.Actor, scope domain.OwnerScope, ownerID int64) ([]service.ConnectionView, error) {
	f.listScope, f.listOwnerID = scope, ownerID
	return f.connections, f.err
}

func (f *fakeService) ListEffective(_ context.Context, _ service.Actor, orgID int64, _ []domain.Modality) ([]service.EffectiveResourceView, error) {
	f.effectiveOrg = orgID
	return f.effective, f.err
}

func (f *fakeService) CreateConnection(_ context.Context, actor service.Actor, input service.CreateConnectionInput) (service.ConnectionView, error) {
	f.createActor, f.createInput = actor, input
	if f.err != nil {
		return service.ConnectionView{}, f.err
	}
	return connectionFixture(input.OwnerScope, input.OwnerID), nil
}

func (f *fakeService) ValidateConnection(_ context.Context, _ service.Actor, connectionID int64) error {
	f.validationID = connectionID
	return f.err
}

type fakeOrg struct{ id int64 }

func (f fakeOrg) GetID() int64  { return f.id }
func (fakeOrg) GetSlug() string { return "acme" }
func (fakeOrg) GetName() string { return "Acme" }

type fakeOrgService struct {
	id     int64
	member bool
	role   string
}

func (f fakeOrgService) GetBySlug(context.Context, string) (middleware.OrganizationGetter, error) {
	return fakeOrg{id: f.id}, nil
}
func (f fakeOrgService) IsMember(context.Context, int64, int64) (bool, error) { return f.member, nil }
func (f fakeOrgService) GetMemberRole(context.Context, int64, int64) (string, error) {
	if !f.member {
		return "", errors.New("not a member")
	}
	return f.role, nil
}

func TestGetCatalogReturnsSafeProviderMetadata(t *testing.T) {
	svc := &fakeService{catalog: []domain.ProviderDefinition{{
		Key: slugkit.Slug("openai"), DisplayName: "OpenAI", DefaultBaseURL: "https://api.openai.com/v1",
		ProtocolAdapter: "openai-compatible", Modalities: []domain.Modality{domain.ModalityChat},
		CredentialFields: []domain.CredentialField{{Key: "api_key", Label: "API key", Secret: true, Required: true}},
		ConnectionCheck:  domain.ConnectionCheck{Method: "GET", Path: "/models", CredentialKey: "api_key"},
	}}}
	srv := NewServer(svc, fakeOrgService{})
	response, err := srv.GetCatalog(userContext(7), connect.NewRequest(&aiv1.GetCatalogRequest{}))
	require.NoError(t, err)
	require.Len(t, response.Msg.Providers, 1)
	encoded, err := json.Marshal(response.Msg.Providers[0])
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "connection_check")
	assert.NotContains(t, string(encoded), "/models")
}

func TestPersonalCreateDerivesOwnerFromAuthenticatedUser(t *testing.T) {
	svc := &fakeService{}
	srv := NewServer(svc, fakeOrgService{})
	ctx := trace.ContextWithSpanContext(userContext(42), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1, 2, 3}, SpanID: trace.SpanID{4, 5, 6}, TraceFlags: trace.FlagsSampled,
	}))
	_, err := srv.CreatePersonalConnection(ctx, connect.NewRequest(&aiv1.CreatePersonalConnectionRequest{
		Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI",
		Credentials: map[string]string{"api_key": "secret"},
	}))
	require.NoError(t, err)
	assert.Equal(t, domain.OwnerScopeUser, svc.createInput.OwnerScope)
	assert.EqualValues(t, 42, svc.createInput.OwnerID)
	assert.EqualValues(t, 42, svc.createActor.UserID)
	assert.Equal(t, "01020300000000000000000000000000", svc.createActor.CorrelationID)
}

func TestOrganizationCreateResolvesSlugAndNeverAcceptsOwnerID(t *testing.T) {
	svc := &fakeService{}
	srv := NewServer(svc, fakeOrgService{id: 81, member: true, role: "admin"})
	request := &aiv1.CreateOrganizationConnectionRequest{
		OrgSlug: "acme", Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI",
		Credentials: map[string]string{"api_key": "secret"},
	}
	_, err := srv.CreateOrganizationConnection(userContext(42), connect.NewRequest(request))
	require.NoError(t, err)
	assert.Equal(t, domain.OwnerScopeOrg, svc.createInput.OwnerScope)
	assert.EqualValues(t, 81, svc.createInput.OwnerID)
	assert.Nil(t, request.ProtoReflect().Descriptor().Fields().ByName("owner_id"))
}

func TestOrganizationEffectiveListUsesResolvedOrganization(t *testing.T) {
	svc := &fakeService{}
	srv := NewServer(svc, fakeOrgService{id: 81, member: true, role: "member"})
	_, err := srv.ListOrganizationEffectiveResources(userContext(42), connect.NewRequest(&aiv1.ListOrganizationEffectiveResourcesRequest{
		OrgSlug: "acme", Modalities: []string{"chat"},
	}))
	require.NoError(t, err)
	assert.EqualValues(t, 81, svc.effectiveOrg)
}

func TestIDScopedValidationPropagatesOrganizationMemberDenialSafely(t *testing.T) {
	svc := &fakeService{err: service.ErrForbidden}
	srv := NewServer(svc, fakeOrgService{})
	_, err := srv.ValidateConnection(userContext(42), connect.NewRequest(&aiv1.ValidateConnectionRequest{
		ConnectionId: 9,
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, errorCode(t, err))
	assert.Equal(t, "AI resource access forbidden", errors.Unwrap(err).Error())
}

func TestMapServiceErrorUsesStableCodesAndMessages(t *testing.T) {
	cases := []struct {
		name    string
		input   error
		code    connect.Code
		message string
	}{
		{"not_found", service.ErrNotFound, connect.CodeNotFound, "AI resource not found"},
		{"forbidden", service.ErrForbidden, connect.CodePermissionDenied, "AI resource access forbidden"},
		{"conflict", service.ErrConflict, connect.CodeAlreadyExists, "AI resource conflict"},
		{"unsupported_before_validation", errors.Join(service.ErrValidation, service.ErrProbeUnsupported), connect.CodeUnimplemented, "AI resource provider validation unsupported"},
		{"credentials_before_validation", errors.Join(service.ErrValidation, service.ErrInvalidCredentials), connect.CodeInvalidArgument, "invalid AI resource credentials"},
		{"endpoint_before_validation", errors.Join(service.ErrValidation, service.ErrInvalidEndpoint), connect.CodeInvalidArgument, "invalid AI resource request"},
		{"validation", service.ErrValidation, connect.CodeFailedPrecondition, "AI resource connection validation failed"},
		{"internal", errors.New("database secret"), connect.CodeInternal, "AI resource operation failed"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := mapServiceError(test.input)
			assert.Equal(t, test.code, errorCode(t, err))
			assert.Equal(t, test.message, errors.Unwrap(err).Error())
			assert.NotContains(t, err.Error(), "database secret")
		})
	}
}

func userContext(userID int64) context.Context {
	return middleware.SetTenant(context.Background(), &middleware.TenantContext{UserID: userID})
}

func errorCode(t *testing.T, err error) connect.Code {
	t.Helper()
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	return connectErr.Code()
}

func connectionFixture(scope domain.OwnerScope, ownerID int64) service.ConnectionView {
	return service.ConnectionView{ID: 1, OwnerScope: scope, OwnerID: ownerID, Identifier: "openai-main", ProviderKey: "openai", Name: "OpenAI"}
}
