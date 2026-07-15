package orchestrationresourceconnect

import (
	"net/http"

	"connectrpc.com/connect"
)

const serviceName = "proto.orchestration_resource.v1.OrchestrationResourceService"

const (
	ValidateResourceProcedure         = "/" + serviceName + "/ValidateResource"
	PlanResourceProcedure             = "/" + serviceName + "/PlanResource"
	GetResourceProcedure              = "/" + serviceName + "/GetResource"
	ListResourcesProcedure            = "/" + serviceName + "/ListResources"
	ExportResourceProcedure           = "/" + serviceName + "/ExportResource"
	GetResourcePlanProcedure          = "/" + serviceName + "/GetResourcePlan"
	ApplyBindingResourcePlanProcedure = "/" + serviceName +
		"/ApplyBindingResourcePlan"
	ApplyWorkerTemplatePlanProcedure = "/" + serviceName +
		"/ApplyWorkerTemplatePlan"
	CreateWorkerFromPlanProcedure = "/" + serviceName +
		"/CreateWorkerFromPlan"
	ApplyPromptPlanProcedure        = "/" + serviceName + "/ApplyPromptPlan"
	ApplyExpertPlanProcedure        = "/" + serviceName + "/ApplyExpertPlan"
	ApplyWorkflowPlanProcedure      = "/" + serviceName + "/ApplyWorkflowPlan"
	CreateGoalLoopFromPlanProcedure = "/" + serviceName +
		"/CreateGoalLoopFromPlan"
)

func Mount(mux *http.ServeMux, server *Server, options ...connect.HandlerOption) {
	mux.Handle(ValidateResourceProcedure, connect.NewUnaryHandler(
		ValidateResourceProcedure,
		server.ValidateResource,
		options...,
	))
	mux.Handle(PlanResourceProcedure, connect.NewUnaryHandler(
		PlanResourceProcedure,
		server.PlanResource,
		options...,
	))
	mux.Handle(GetResourceProcedure, connect.NewUnaryHandler(
		GetResourceProcedure,
		server.GetResource,
		options...,
	))
	mux.Handle(ListResourcesProcedure, connect.NewUnaryHandler(
		ListResourcesProcedure,
		server.ListResources,
		options...,
	))
	mux.Handle(ExportResourceProcedure, connect.NewUnaryHandler(
		ExportResourceProcedure,
		server.ExportResource,
		options...,
	))
	mux.Handle(GetResourcePlanProcedure, connect.NewUnaryHandler(
		GetResourcePlanProcedure,
		server.GetResourcePlan,
		options...,
	))
	mux.Handle(ApplyBindingResourcePlanProcedure, connect.NewUnaryHandler(
		ApplyBindingResourcePlanProcedure,
		server.ApplyBindingResourcePlan,
		options...,
	))
	mux.Handle(ApplyWorkerTemplatePlanProcedure, connect.NewUnaryHandler(
		ApplyWorkerTemplatePlanProcedure,
		server.ApplyWorkerTemplatePlan,
		options...,
	))
	mux.Handle(CreateWorkerFromPlanProcedure, connect.NewUnaryHandler(
		CreateWorkerFromPlanProcedure,
		server.CreateWorkerFromPlan,
		options...,
	))
	mux.Handle(ApplyPromptPlanProcedure, connect.NewUnaryHandler(
		ApplyPromptPlanProcedure,
		server.ApplyPromptPlan,
		options...,
	))
	mux.Handle(ApplyExpertPlanProcedure, connect.NewUnaryHandler(
		ApplyExpertPlanProcedure,
		server.ApplyExpertPlan,
		options...,
	))
	mux.Handle(ApplyWorkflowPlanProcedure, connect.NewUnaryHandler(
		ApplyWorkflowPlanProcedure,
		server.ApplyWorkflowPlan,
		options...,
	))
	mux.Handle(CreateGoalLoopFromPlanProcedure, connect.NewUnaryHandler(
		CreateGoalLoopFromPlanProcedure,
		server.CreateGoalLoopFromPlan,
		options...,
	))
}
