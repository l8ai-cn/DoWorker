package workerspec

import "errors"

var (
	ErrInvalidScope                  = errors.New("workerspec scope requires positive org and user ids")
	ErrResolverUnavailable           = errors.New("workerspec resolver dependency is unavailable")
	ErrSnapshotRepositoryUnavailable = errors.New("workerspec snapshot repository is unavailable")
	ErrInvalidResolvedSnapshot       = errors.New("resolved workerspec snapshot is invalid")
)
