package main

import (
	"fmt"

	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
)

func attachOrchestrationWorkerApply(
	services *serviceContainer,
	orchestrator workerPodOrchestrator,
	queue workerApplyDispatchQueue,
) error {
	if services == nil || orchestrator == nil || queue == nil {
		return fmt.Errorf(
			"orchestration Worker apply runtime dependencies are incomplete",
		)
	}
	runtime := services.workerApplyRuntime
	workerApply, err := workerplanner.NewWorkerApplyService(
		runtime.registry,
		runtime.repository,
		runtime.resolver,
		newOrchestrationWorkerPodLauncher(orchestrator, queue),
		newOrchestrationWorkerDispatchNotifier(queue),
	)
	if err != nil {
		return err
	}
	services.workerApply = workerApply
	return nil
}
