package sessionapi

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSessionReportsPodCompensationFailure(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	lifecycle.terminateErr = errors.New("termination unavailable")
	require.NoError(t, failSessionInserts(db))

	response := createSessionCompensationRequest(deps, `{"agent_id":"codex-cli"}`)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{
		"error":"session creation cleanup failed",
		"code":"session_compensation_failed"
	}`, response.Body.String())
}

func TestCreateSessionTerminatesPodWhenSessionPersistenceFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	require.NoError(t, failSessionInserts(db))

	response := createSessionCompensationRequest(deps, `{"agent_id":"codex-cli"}`)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}
