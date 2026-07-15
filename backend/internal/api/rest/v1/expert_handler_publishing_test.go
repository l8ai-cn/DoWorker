package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInstallMarketApplicationRequiresModelResourceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{`{}`, `{"model_resource_id":0}`} {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(
			http.MethodPost,
			"/marketplace/experts/video/install",
			strings.NewReader(body),
		)
		context.Request.Header.Set("Content-Type", "application/json")
		context.Params = gin.Params{{Key: "marketSlug", Value: "video"}}

		(&ExpertHandler{}).InstallMarketApplication(context)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
	}
}

func TestInstallMarketErrorMapping(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{expertsvc.ErrMarketApplicationNotFound, http.StatusNotFound},
		{expertsvc.ErrMarketUnavailable, http.StatusServiceUnavailable},
		{specservice.ErrInvalidDraft, http.StatusBadRequest},
		{expertsvc.ErrMarketSnapshotInvalid, http.StatusConflict},
		{expertmarket.ErrConflict, http.StatusConflict},
		{expertdom.ErrConflict, http.StatusConflict},
		{expertdom.ErrNotFound, http.StatusNotFound},
		{errors.New("database offline"), http.StatusInternalServerError},
	}
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		(&ExpertHandler{}).installMarketError(context, test.err)

		require.Equal(t, test.status, recorder.Code)
	}
}
