package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	expertSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/expert"
)

func TestUpdateExpertReturnsConflictForResourceManagedDefinition(t *testing.T) {
	handler, store := newResourceManagedExpertHandler()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPatch,
		"/experts/review",
		strings.NewReader(`{"name":"Changed"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "expertSlug", Value: "review"}}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 7, UserID: 5})

	handler.UpdateExpert(ctx)

	assertResourceManagedDefinitionConflict(t, response)
	assert.False(t, store.updated)
}

func TestDeleteExpertReturnsConflictForResourceManagedDefinition(t *testing.T) {
	handler, store := newResourceManagedExpertHandler()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/experts/review", nil)
	ctx.Params = gin.Params{{Key: "expertSlug", Value: "review"}}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 7, UserID: 5})

	handler.DeleteExpert(ctx)

	assertResourceManagedDefinitionConflict(t, response)
	assert.False(t, store.deleted)
}

func assertResourceManagedDefinitionConflict(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusConflict, response.Code)
	var body struct {
		Code  string `json:"code"`
		Error string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "EXPERT_MANAGED_BY_RESOURCE_APPLY", body.Code)
	assert.Equal(t, "Expert definition changes must go through resource validate-plan-apply", body.Error)
}

func newResourceManagedExpertHandler() (*ExpertHandler, *expertHandlerStore) {
	snapshotID := int64(42)
	resourceID := int64(90)
	resourceRevision := int64(3)
	store := &expertHandlerStore{row: &expertdom.Expert{
		ID:                            8,
		OrganizationID:                7,
		Slug:                          "review",
		Name:                          "Review",
		AgentSlug:                     "resource-native",
		WorkerSpecSnapshotID:          &snapshotID,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
	}}
	return NewExpertHandler(expertSvc.NewService(expertSvc.Deps{Store: store})), store
}

type expertHandlerStore struct {
	row     *expertdom.Expert
	updated bool
	deleted bool
}

func (s *expertHandlerStore) Create(context.Context, *expertdom.Expert) error {
	return expertdom.ErrNotFound
}

func (s *expertHandlerStore) Update(_ context.Context, row *expertdom.Expert) error {
	s.updated = true
	s.row = cloneHandlerExpert(row)
	return nil
}

func (s *expertHandlerStore) Delete(_ context.Context, _ int64, _ int64) error {
	s.deleted = true
	return nil
}

func (s *expertHandlerStore) GetByID(_ context.Context, orgID, id int64) (*expertdom.Expert, error) {
	if s.row == nil || s.row.OrganizationID != orgID || s.row.ID != id {
		return nil, expertdom.ErrNotFound
	}
	return cloneHandlerExpert(s.row), nil
}

func (s *expertHandlerStore) GetBySlug(_ context.Context, orgID int64, slug string) (*expertdom.Expert, error) {
	if s.row == nil || s.row.OrganizationID != orgID || s.row.Slug != slug {
		return nil, expertdom.ErrNotFound
	}
	return cloneHandlerExpert(s.row), nil
}

func (s *expertHandlerStore) GetByMarketApplication(context.Context, int64, int64) (*expertdom.Expert, error) {
	return nil, expertdom.ErrNotFound
}

func (s *expertHandlerStore) UpdateMarketRelease(
	context.Context,
	int64,
	int64,
	int64,
	expertdom.MarketReleaseUpdate,
) error {
	return expertdom.ErrNotFound
}

func (s *expertHandlerStore) SlugExists(context.Context, int64, string, int64) (bool, error) {
	return false, nil
}

func (s *expertHandlerStore) List(context.Context, int64, int, int) ([]expertdom.Expert, int64, error) {
	return nil, 0, nil
}

func (s *expertHandlerStore) RecordRun(context.Context, int64, int64, time.Time) error {
	return expertdom.ErrNotFound
}

func cloneHandlerExpert(row *expertdom.Expert) *expertdom.Expert {
	copy := *row
	return &copy
}
