package agentworkbenchconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	workbenchsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentworkbench"
)

func commandExecutionError(err error) error {
	if connect.CodeOf(err) != connect.CodeUnknown {
		return err
	}
	switch {
	case errors.Is(err, workbenchsvc.ErrInvalidCommand),
		errors.Is(err, workbenchdomain.ErrInvalidArgument):
		return invalidArgument("agent workbench command is invalid")
	case errors.Is(err, workbenchsvc.ErrCommandConflict),
		errors.Is(err, workbenchdomain.ErrCommandIDConflict),
		errors.Is(err, workbenchdomain.ErrReceiptConflict):
		return aborted("agent workbench command conflicts with persisted state")
	case errors.Is(err, workbenchsvc.ErrCommandUnavailable):
		return unavailable("agent workbench command delivery is unavailable")
	case errors.Is(err, workbenchdomain.ErrRevisionConflict),
		errors.Is(err, workbenchdomain.ErrStreamConflict):
		return failedPrecondition("agent workbench command state changed; resync required")
	case errors.Is(err, context.Canceled):
		return canceled(err)
	default:
		return internalError(err)
	}
}
