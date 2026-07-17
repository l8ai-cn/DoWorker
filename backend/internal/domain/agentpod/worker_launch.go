package agentpod

import "errors"

var ErrWorkerLaunchPodAlreadyExists = errors.New(
	"orchestration worker launch already has a pod",
)
