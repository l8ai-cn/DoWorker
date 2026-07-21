package consumer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	actorapi "github.com/l8ai-cn/agentcloud/marketplace/internal/api/actor"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCreatePlanUsesAuthenticatedUser(t *testing.T) {
	orchestration := &installationOrchestratorStub{
		plan: service.InstallationPlanResult{
			InstallationID:   "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			OperationID:      "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
			PlanID:           "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
			PlanDigest:       strings.Repeat("d", 64),
			ListingVersionID: 301, EstimatedCredits: 20_000_000,
			ExpiresAt:   time.Date(2026, 7, 12, 8, 15, 0, 0, time.UTC),
			Permissions: []byte(`["repository.write"]`),
		},
	}
	router := authenticatedInstallationRouter(orchestration)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/marketplace/v1/markets/commerce-market/listings/listing-optimizer/plans",
		strings.NewReader(`{
		  "listing_version_id":"301",
		  "target_platform_organization_id":"9",
		  "requested_configuration":{"model_resource_id":"18"}
		}`),
	)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code)
	require.Equal(t, int64(14), orchestration.createCommand.ActorUserID)
	require.Contains(t, response.Body.String(), `"estimated_credits_micro":"20000000"`)
}

func TestGetOperationUsesAuthenticatedUser(t *testing.T) {
	orchestration := &installationOrchestratorStub{
		apply: service.ApplyResult{
			OperationID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
			Status:      service.ApplySucceeded,
		},
	}
	router := authenticatedInstallationRouter(orchestration)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/installation-operations/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		nil,
	)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, int64(14), orchestration.operationActorUserID)
}

func TestApplyRequiresIdempotencyKey(t *testing.T) {
	orchestration := &installationOrchestratorStub{
		applyErr: service.ErrInvalidInstallationRequest,
	}
	router := authenticatedInstallationRouter(orchestration)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/marketplace/v1/installation-operations/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb/apply",
		strings.NewReader(`{
		  "plan_id":"cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		  "plan_digest":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
		}`),
	)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{
	  "error":{"code":"INVALID_INSTALLATION_REQUEST","message":"安装请求无效"}
	}`, response.Body.String())
	require.Empty(t, orchestration.applyCommand.IdempotencyKey)
}

func authenticatedInstallationRouter(
	orchestration InstallationOrchestrator,
) *gin.Engine {
	router := gin.New()
	group := router.Group("/api/marketplace/v1")
	group.Use(actorapi.Middleware(tokenVerifierStub{}))
	NewInstallationHandler(orchestration).RegisterRoutes(group)
	return router
}

type tokenVerifierStub struct{}

func (tokenVerifierStub) Verify(context.Context, string) (*authpkg.Claims, error) {
	return &authpkg.Claims{
		UserID: 14, OrganizationID: 9, Email: "user@example.com",
	}, nil
}

type installationOrchestratorStub struct {
	createCommand        service.CreateInstallationPlanCommand
	applyCommand         service.ApplyInstallationCommand
	plan                 service.InstallationPlanResult
	apply                service.ApplyResult
	applyErr             error
	operationActorUserID int64
}

func (s *installationOrchestratorStub) CreatePlan(
	_ context.Context,
	command service.CreateInstallationPlanCommand,
) (service.InstallationPlanResult, error) {
	s.createCommand = command
	return s.plan, nil
}

func (s *installationOrchestratorStub) Apply(
	_ context.Context,
	command service.ApplyInstallationCommand,
) (service.ApplyResult, error) {
	s.applyCommand = command
	return s.apply, s.applyErr
}

func (s *installationOrchestratorStub) GetOperation(
	_ context.Context,
	_ string,
	actorUserID int64,
) (service.ApplyResult, error) {
	s.operationActorUserID = actorUserID
	return s.apply, nil
}
