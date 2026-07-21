package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	orchestrationresourceconnect "github.com/l8ai-cn/agentcloud/backend/internal/api/connect/orchestration_resource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/organization"
	"github.com/stretchr/testify/assert"
)

func TestMountOrchestrationResourceServiceRegistersControlPlane(t *testing.T) {
	mux := http.NewServeMux()
	mountOrchestrationResourceService(
		mux,
		&serviceContainer{
			orchestration:       &controlservice.Service{},
			bindingApply:        new(workerplanner.BindingApplyService),
			workerTemplateApply: new(workerplanner.WorkerTemplateApplyService),
			workerApply:         new(workerplanner.WorkerApplyService),
			promptApply:         new(workerplanner.PromptApplyService),
			expertApply:         new(workerplanner.ExpertApplyService),
			workflowApply:       new(workerplanner.WorkflowApplyService),
			goalLoopApply:       new(workerplanner.GoalLoopApplyService),
			org:                 organization.NewService(nil),
		},
		nil,
	)

	_, pattern := mux.Handler(httptest.NewRequest(
		http.MethodPost,
		orchestrationresourceconnect.PlanResourceProcedure,
		nil,
	))
	assert.Equal(t, orchestrationresourceconnect.PlanResourceProcedure, pattern)
}
