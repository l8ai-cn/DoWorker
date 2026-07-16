package orchestrationresource

import "errors"

var (
	ErrDuplicateJSONKey = errors.New("duplicate JSON key")
	ErrUnknownJSONField = errors.New("unknown JSON field")
)
