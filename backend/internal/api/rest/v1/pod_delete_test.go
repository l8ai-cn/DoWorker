package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeletePod_DeletesTerminalPod(t *testing.T) {
	pod := &agentpod.Pod{
		PodKey: "pod-completed", OrganizationID: 1, CreatedByID: 10, Status: agentpod.StatusCompleted,
	}
	deleted := false
	handler := &PodHandler{podService: &mockPodService{
		getPodFn: func(context.Context, string) (*agentpod.Pod, error) { return pod, nil },
		deletePodFn: func(_ context.Context, key string) error {
			deleted = key == pod.PodKey
			return nil
		},
	}}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/pods/pod-completed", nil)
	c.Params = gin.Params{{Key: "key", Value: pod.PodKey}}
	setPodTenantContext(c, 1, 10)

	handler.DeletePod(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, deleted)
}
