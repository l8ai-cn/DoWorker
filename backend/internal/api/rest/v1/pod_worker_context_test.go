package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	workerspecdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	workerspecservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

type workerContextSnapshots struct {
	snapshot workerspecdomain.Snapshot
}

func (f workerContextSnapshots) Create(context.Context, workerspecservice.ResolvedSnapshot) (workerspecdomain.Snapshot, error) {
	return workerspecdomain.Snapshot{}, nil
}

func (f workerContextSnapshots) GetByID(context.Context, int64, int64) (workerspecdomain.Snapshot, error) {
	return f.snapshot, nil
}

func (f workerContextSnapshots) GetByIDs(context.Context, int64, []int64) ([]workerspecdomain.Snapshot, error) {
	return []workerspecdomain.Snapshot{f.snapshot}, nil
}

func (f workerContextSnapshots) ListByOrganization(context.Context, int64) ([]workerspecdomain.Snapshot, error) {
	return nil, nil
}

func (f workerContextSnapshots) Delete(context.Context, int64, int64) error {
	return nil
}

func TestGetPodWorkerContextUsesImmutableSnapshotSkills(t *testing.T) {
	snapshotID := int64(91)
	handler := &PodHandler{
		podService: &mockPodService{getPodFn: func(context.Context, string) (*agentpod.Pod, error) {
			return &agentpod.Pod{
				PodKey: "video-worker-1", OrganizationID: 1, CreatedByID: 10,
				PodResourceBindings: agentpod.PodResourceBindings{
					WorkerSpecSnapshotID: &snapshotID,
				},
			}, nil
		}},
		workerSpecs: workerContextSnapshots{snapshot: workerspecdomain.Snapshot{
			ID: snapshotID,
			Spec: workerspecdomain.Spec{
				Metadata: workerspecdomain.Metadata{Alias: "video-production-expert"},
				Workspace: workerspecdomain.Workspace{SkillPackages: []workerspecdomain.SkillPackageBinding{
					{Slug: "short-video-directing"},
					{Slug: "video-delivery-qa"},
				}},
			},
		}},
	}
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/pods/video-worker-1/worker-context", nil)
	ctx.Params = gin.Params{{Key: "key", Value: "video-worker-1"}}
	setPodTenantContext(ctx, 1, 10)

	handler.GetPodWorkerContext(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Worker podWorkerContext `json:"worker"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, snapshotID, body.Worker.SnapshotID)
	assert.Equal(t, "video-production-expert", body.Worker.Alias)
	assert.Equal(t, []string{"short-video-directing", "video-delivery-qa"}, body.Worker.SkillSlugs)
}
