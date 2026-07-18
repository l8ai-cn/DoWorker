package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentpodsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
)

func TestRunExpertMapsSnapshotMismatchToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	snapshotID := int64(42)
	service := expertsvc.NewService(expertsvc.Deps{
		Store: &runExpertStore{row: &expertdom.Expert{
			ID:                   1,
			OrganizationID:       7,
			Slug:                 "review",
			Name:                 "Review",
			WorkerSpecSnapshotID: &snapshotID,
		}},
		Dispatch:    &runExpertDispatcher{},
		WorkerSpecs: &runExpertSnapshotLoader{err: specdomain.ErrNotFound},
	})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 7,
		UserID:         5,
	})
	ctx.Params = gin.Params{{Key: "expertSlug", Value: "review"}}
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/experts/review/run",
		strings.NewReader("{}"),
	)

	NewExpertHandler(service).RunExpert(ctx)

	assert.Equal(t, http.StatusConflict, recorder.Code)
}

func TestRunExpertMapsSnapshotServiceFailureToUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	snapshotID := int64(42)
	service := expertsvc.NewService(expertsvc.Deps{
		Store: &runExpertStore{row: &expertdom.Expert{
			ID:                   1,
			OrganizationID:       7,
			Slug:                 "review",
			Name:                 "Review",
			WorkerSpecSnapshotID: &snapshotID,
		}},
		Dispatch: &runExpertDispatcher{},
		WorkerSpecs: &runExpertSnapshotLoader{
			err: errors.New("database unavailable"),
		},
	})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 7,
		UserID:         5,
	})
	ctx.Params = gin.Params{{Key: "expertSlug", Value: "review"}}
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/experts/review/run",
		strings.NewReader("{}"),
	)

	NewExpertHandler(service).RunExpert(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestRunExpertMapsOrchestratorSnapshotErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "mismatch",
			err:    agentpodsvc.ErrWorkerSpecSnapshotMismatch,
			status: http.StatusConflict,
		},
		{
			name:   "unavailable",
			err:    agentpodsvc.ErrWorkerSpecSnapshotUnavailable,
			status: http.StatusServiceUnavailable,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			(&ExpertHandler{}).runError(ctx, test.err)

			assert.Equal(t, test.status, recorder.Code)
		})
	}
}

type runExpertStore struct {
	row *expertdom.Expert
}

func (*runExpertStore) Create(context.Context, *expertdom.Expert) error {
	return nil
}

func (*runExpertStore) Update(context.Context, *expertdom.Expert) error {
	return nil
}

func (*runExpertStore) Delete(context.Context, int64, int64) error {
	return nil
}

func (store *runExpertStore) GetByID(context.Context, int64, int64) (*expertdom.Expert, error) {
	return store.row, nil
}

func (store *runExpertStore) GetBySlug(context.Context, int64, string) (*expertdom.Expert, error) {
	return store.row, nil
}

func (*runExpertStore) SlugExists(context.Context, int64, string, int64) (bool, error) {
	return false, nil
}

func (*runExpertStore) List(context.Context, int64, int, int) ([]expertdom.Expert, int64, error) {
	return nil, 0, nil
}

func (*runExpertStore) RecordRun(context.Context, int64, int64, time.Time) error {
	return nil
}

type runExpertDispatcher struct{}

func (*runExpertDispatcher) CreatePod(
	context.Context,
	*agentpodsvc.OrchestrateCreatePodRequest,
) (*agentpodsvc.OrchestrateCreatePodResult, error) {
	return &agentpodsvc.OrchestrateCreatePodResult{}, nil
}

type runExpertSnapshotLoader struct {
	err error
}

func (loader *runExpertSnapshotLoader) GetByID(
	context.Context,
	int64,
	int64,
) (specdomain.Snapshot, error) {
	return specdomain.Snapshot{}, loader.err
}
