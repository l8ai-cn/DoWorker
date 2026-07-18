package grpc

import (
	"errors"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
)

func TestMapResourceControlErrorUsesStableMessages(t *testing.T) {
	tests := []struct {
		err  error
		code int32
	}{
		{control.ErrInvalid, 400},
		{controlservice.ErrForbidden, 403},
		{control.ErrNotFound, 404},
		{control.ErrStale, 409},
		{controlservice.ErrUnavailable, 503},
		{errors.New("SQLSTATE secret"), 500},
	}
	for _, test := range tests {
		mapped := mapResourceControlError(test.err)
		assert.Equal(t, test.code, mapped.code)
		assert.NotContains(t, mapped.message, "SQLSTATE")
		assert.NotContains(t, mapped.message, "secret")
	}
}
