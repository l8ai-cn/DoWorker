package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	orchestrationresourceconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/orchestration_resource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
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
