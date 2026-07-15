package orchestrationresourceconnect

import (
	"errors"

	"connectrpc.com/connect"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func mapServiceError(err error) error {
	if err == nil {
		return nil
	}
	code := connect.CodeInternal
	message := "orchestration resource operation failed"
	switch {
	case errors.Is(err, control.ErrInvalid):
		code = connect.CodeInvalidArgument
		message = "invalid orchestration resource request"
	case errors.Is(err, service.ErrForbidden):
		code = connect.CodePermissionDenied
		message = "orchestration resource access forbidden"
	case errors.Is(err, control.ErrNotFound):
		code = connect.CodeNotFound
		message = "orchestration resource not found"
	case errors.Is(err, control.ErrConflict),
		errors.Is(err, control.ErrStale),
		errors.Is(err, control.ErrExpired),
		errors.Is(err, control.ErrConsumed),
		errors.Is(err, service.ErrStaleOptions):
		code = connect.CodeAborted
		message = "orchestration resource state changed"
	case errors.Is(err, service.ErrUnavailable):
		code = connect.CodeUnavailable
		message = "orchestration resource service unavailable"
	}
	return connect.NewError(code, errors.New(message))
}
