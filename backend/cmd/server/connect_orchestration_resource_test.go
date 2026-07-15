package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	orchestrationresourceconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/orchestration_resource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/stretchr/testify/assert"
)

func TestMountOrchestrationResourceServiceRegistersPhase2AControlPlane(
	t *testing.T,
) {
	mux := http.NewServeMux()
	mountOrchestrationResourceService(
		mux,
		&serviceContainer{
			orchestration: &controlservice.Service{},
			org:           organization.NewService(nil),
		},
		nil,
	)

	for _, procedure := range []string{
		orchestrationresourceconnect.ValidateResourceProcedure,
		orchestrationresourceconnect.PlanResourceProcedure,
		orchestrationresourceconnect.GetResourceProcedure,
		orchestrationresourceconnect.ListResourcesProcedure,
		orchestrationresourceconnect.ExportResourceProcedure,
		orchestrationresourceconnect.GetResourcePlanProcedure,
	} {
		_, pattern := mux.Handler(httptest.NewRequest(
			http.MethodPost,
			procedure,
			nil,
		))
		assert.Equal(t, procedure, pattern)
	}

	_, pattern := mux.Handler(httptest.NewRequest(
		http.MethodPost,
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyWorkerTemplatePlan",
		nil,
	))
	assert.Empty(t, pattern)
}
