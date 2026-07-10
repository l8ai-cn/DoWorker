package workerruntime

import "errors"

var (
	ErrInvalidRequest        = errors.New("invalid runtime resource request")
	ErrNotFound              = errors.New("runtime resource not found")
	ErrDisabled              = errors.New("runtime resource is disabled")
	ErrIncompatible          = errors.New("runtime resource combination is incompatible")
	ErrInvalidResolvedValue  = errors.New("invalid resolved runtime resource value")
	ErrRepositoryUnavailable = errors.New("runtime resource repository is unavailable")
)
