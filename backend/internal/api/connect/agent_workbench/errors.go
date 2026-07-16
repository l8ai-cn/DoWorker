package agentworkbenchconnect

import (
	"errors"

	"connectrpc.com/connect"
)

func invalidArgument(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, errors.New(message))
}

func unauthenticated(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(message))
}

func notFound(message string) error {
	return connect.NewError(connect.CodeNotFound, errors.New(message))
}

func permissionDenied(message string) error {
	return connect.NewError(connect.CodePermissionDenied, errors.New(message))
}

func unavailable(message string) error {
	return connect.NewError(connect.CodeUnavailable, errors.New(message))
}

func failedPrecondition(message string) error {
	return connect.NewError(connect.CodeFailedPrecondition, errors.New(message))
}

func aborted(message string) error {
	return connect.NewError(connect.CodeAborted, errors.New(message))
}

func dataLoss(message string) error {
	return connect.NewError(connect.CodeDataLoss, errors.New(message))
}

func internalError(err error) error {
	return connect.NewError(connect.CodeInternal, err)
}

func canceled(err error) error {
	return connect.NewError(connect.CodeCanceled, err)
}
