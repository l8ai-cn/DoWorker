package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLiveDoesNotDependOnDatabase(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready: func(context.Context) error { return errors.New("database unavailable") },
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"live"}`, response.Body.String())
}

func TestReadyReportsDatabaseFailure(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready: func(context.Context) error { return errors.New("database unavailable") },
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "SERVICE_NOT_READY",
			"message": "市场服务尚未就绪"
		}
	}`, response.Body.String())
}

func TestReadySucceedsAfterDatabaseProbe(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready: func(context.Context) error { return nil },
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"ready"}`, response.Body.String())
}
