package sessionapi

import (
	"context"
	"net/http/httptest"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingAIModelRepository struct {
	byIDModel       *domain.AIModel
	visibleModel    *domain.AIModel
	getByIDCalls    int
	getVisibleCalls int
	modelID         int64
	userID          int64
	organizationID  int64
}

func (r *recordingAIModelRepository) GetByID(_ context.Context, id int64) (*domain.AIModel, error) {
	r.getByIDCalls++
	r.modelID = id
	return r.byIDModel, nil
}

func (r *recordingAIModelRepository) GetVisibleByID(_ context.Context, id, userID, organizationID int64) (*domain.AIModel, error) {
	r.getVisibleCalls++
	r.modelID = id
	r.userID = userID
	r.organizationID = organizationID
	return r.visibleModel, nil
}

func (r *recordingAIModelRepository) Create(context.Context, *domain.AIModel) error { return nil }
func (r *recordingAIModelRepository) Save(context.Context, *domain.AIModel) error   { return nil }
func (r *recordingAIModelRepository) Delete(context.Context, int64) error           { return nil }
func (r *recordingAIModelRepository) ListVisible(context.Context, int64, int64) ([]*domain.AIModel, error) {
	return nil, nil
}
func (r *recordingAIModelRepository) DefaultVisible(context.Context, int64, int64) (*domain.AIModel, error) {
	return nil, nil
}
func (r *recordingAIModelRepository) ClearDefaults(context.Context, int64, int64) error { return nil }
func (r *recordingAIModelRepository) CountOrg(context.Context, int64) (int64, error)    { return 0, nil }
func (r *recordingAIModelRepository) FirstVisibleByProvider(context.Context, int64, int64, string) (*domain.AIModel, error) {
	return nil, nil
}

func TestResolveWorkerModel_ExplicitModelUsesCallerScope(t *testing.T) {
	model := sessionModelFixture(5)
	repo := &recordingAIModelRepository{byIDModel: model, visibleModel: model}
	deps := &Deps{AIModels: aimodelsvc.NewService(repo, nil)}
	ctx := workerModelTestContext()
	var layer *string

	mount, err := deps.resolveWorkerModel(ctx, 11, 21, createSessionBody{
		AgentID:       "codex-cli",
		ModelConfigID: sessionModelIDPointer(5),
	}, &layer)

	require.NoError(t, err)
	require.NotNil(t, mount)
	assert.Equal(t, 1, repo.getVisibleCalls)
	assert.Zero(t, repo.getByIDCalls)
	assert.Equal(t, int64(5), repo.modelID)
	assert.Equal(t, int64(11), repo.userID)
	assert.Equal(t, int64(21), repo.organizationID)
}

func TestResolveWorkerModel_InvisibleExplicitModelDoesNotMount(t *testing.T) {
	repo := &recordingAIModelRepository{byIDModel: sessionModelFixture(5)}
	deps := &Deps{AIModels: aimodelsvc.NewService(repo, nil)}
	ctx := workerModelTestContext()
	var layer *string

	mount, err := deps.resolveWorkerModel(ctx, 11, 21, createSessionBody{
		AgentID:       "codex-cli",
		ModelConfigID: sessionModelIDPointer(5),
	}, &layer)

	require.EqualError(t, err, "selected model not found")
	assert.Nil(t, mount)
	assert.Nil(t, layer)
	assert.Equal(t, 1, repo.getVisibleCalls)
	assert.Zero(t, repo.getByIDCalls)
}

func workerModelTestContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/api/sessions", nil)
	return c
}

func sessionModelFixture(id int64) *domain.AIModel {
	return &domain.AIModel{
		ID:                   id,
		ProviderType:         domain.ProviderTypeOpenAI,
		Model:                "gpt-5",
		EncryptedCredentials: `{"api_key":"sk-test"}`,
		IsEnabled:            true,
	}
}

func sessionModelIDPointer(value int64) *int64 { return &value }
