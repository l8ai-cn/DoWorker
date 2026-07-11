package workercreation

import "errors"

var (
	ErrStaleOptions                = errors.New("worker create options revision is stale")
	ErrWorkerTypeDefinitionChanged = errors.New("worker type definition changed")
)
