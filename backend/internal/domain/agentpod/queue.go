package agentpod

import (
	"errors"
	"time"
)

var ErrPodQueued = errors.New("pod queued for runner dispatch")

type CreatePodQueueOpts struct {
	Queue bool
	TTL   time.Duration
	OrgID int64
}

func IsPodQueued(err error) bool {
	return errors.Is(err, ErrPodQueued)
}
