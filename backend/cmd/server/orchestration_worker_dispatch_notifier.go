package main

import workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"

type workerDispatchQueue interface {
	TriggerDrain(int64)
	SealPayload([]byte) ([]byte, error)
}

type pendingPayloadSealer interface {
	SealPayload([]byte) ([]byte, error)
}

type orchestrationWorkerDispatchNotifier struct {
	queue workerDispatchQueue
}

func newOrchestrationWorkerDispatchNotifier(
	queue workerDispatchQueue,
) workerplanner.WorkerDispatchNotifier {
	if queue == nil {
		return nil
	}
	return &orchestrationWorkerDispatchNotifier{queue: queue}
}

func (notifier *orchestrationWorkerDispatchNotifier) TriggerWorkerDispatch(
	runnerID int64,
) {
	notifier.queue.TriggerDrain(runnerID)
}
