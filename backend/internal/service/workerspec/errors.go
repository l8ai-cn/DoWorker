package workerspec

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidScope                  = errors.New("workerspec scope requires positive org and user ids")
	ErrInvalidDraft                  = errors.New("workerspec draft is invalid")
	ErrResolverUnavailable           = errors.New("workerspec resolver dependency is unavailable")
	ErrSnapshotRepositoryUnavailable = errors.New("workerspec snapshot repository is unavailable")
	ErrInvalidResolvedSnapshot       = errors.New("resolved workerspec snapshot is invalid")
)

type InvalidDraftFieldError struct {
	Field  string
	Reason string
}

func (err *InvalidDraftFieldError) Error() string {
	return fmt.Sprintf("%s: %s: %s", ErrInvalidDraft, err.Field, err.Reason)
}

func (*InvalidDraftFieldError) Unwrap() error {
	return ErrInvalidDraft
}

func InvalidDraftField(err error) string {
	var fieldError *InvalidDraftFieldError
	if errors.As(err, &fieldError) {
		return fieldError.Field
	}
	return ""
}
