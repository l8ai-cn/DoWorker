package adminconnect

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	adminservice "github.com/anthropics/agentsmesh/backend/internal/service/admin"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
)

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, adminservice.ErrUserNotFound),
		errors.Is(err, adminservice.ErrOrganizationNotFound),
		errors.Is(err, adminservice.ErrRunnerNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, adminservice.ErrUsernameAlreadyExists),
		errors.Is(err, adminservice.ErrEmailAlreadyExists),
		errors.Is(err, adminservice.ErrOrganizationHasActiveRunner),
		errors.Is(err, adminservice.ErrRunnerHasActivePods),
		errors.Is(err, adminservice.ErrRunnerHasWorkflowRefs):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, adminservice.ErrCannotRevokeOwnAdmin),
		errors.Is(err, adminservice.ErrCannotDisableSelf):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func mapExpertMarketError(err error) error {
	switch {
	case errors.Is(err, expertsvc.ErrMarketApplicationNotFound),
		errors.Is(err, expertmarket.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, expertsvc.ErrMarketUnavailable):
		return connect.NewError(connect.CodeUnavailable, err)
	case errors.Is(err, expertsvc.ErrMarketRejectionReasonRequired):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, expertsvc.ErrMarketInvalidTransition),
		errors.Is(err, expertmarket.ErrLifecycleStatusConflict),
		errors.Is(err, expertmarket.ErrConflict):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, expertmarket.ErrInvalidStatus):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
