package orchestrationcontrol

import "errors"

var (
	ErrUnavailable  = errors.New("orchestration service unavailable")
	ErrForbidden    = errors.New("orchestration operation forbidden")
	ErrStaleOptions = errors.New("orchestration options are stale")
)
