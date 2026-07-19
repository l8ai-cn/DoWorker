package podconnect

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

var (
	errWorkerSpecRequired = errors.New(
		"fresh pod creation requires worker_spec",
	)
	errLegacyPodRuntimeFields = errors.New(
		"legacy pod runtime fields are not supported; use worker_spec",
	)
	errResumeRuntimeOverrides = errors.New(
		"resume accepts source lineage only; runtime overrides are not supported",
	)
)

func buildCreatePodRequest(
	message *podv1.CreatePodRequest,
	tenant *middleware.TenantContext,
) (*agentpod.OrchestrateCreatePodRequest, error) {
	if message.GetSourcePodKey() != "" {
		if message.WorkerSpec != nil || hasLegacyPodRuntimeFields(message) {
			return nil, connect.NewError(
				connect.CodeFailedPrecondition,
				errResumeRuntimeOverrides,
			)
		}
		return buildResumePodRequest(message, tenant), nil
	}
	if message.WorkerSpec == nil {
		return nil, connect.NewError(
			connect.CodeFailedPrecondition,
			errWorkerSpecRequired,
		)
	}
	if hasLegacyPodRuntimeFields(message) {
		return nil, connect.NewError(
			connect.CodeFailedPrecondition,
			errLegacyPodRuntimeFields,
		)
	}
	draft, err := workerDraftFromProto(message.WorkerSpec)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	orgSlug, err := slugkit.NewFromTrusted(tenant.OrganizationSlug)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	draft.OrganizationSlug = orgSlug
	return buildWorkerSpecPodRequest(message, tenant, draft), nil
}

func buildWorkerSpecPodRequest(
	message *podv1.CreatePodRequest,
	tenant *middleware.TenantContext,
	draft workercreation.Draft,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:  tenant.OrganizationID,
		UserID:          tenant.UserID,
		TicketSlug:      optionalString(message.TicketSlug),
		Cols:            message.GetCols(),
		Rows:            message.GetRows(),
		WorkerSpecDraft: &draft,
	}
}

func buildResumePodRequest(
	message *podv1.CreatePodRequest,
	tenant *middleware.TenantContext,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:     tenant.OrganizationID,
		UserID:             tenant.UserID,
		TicketSlug:         optionalString(message.TicketSlug),
		Cols:               message.GetCols(),
		Rows:               message.GetRows(),
		SourcePodKey:       message.GetSourcePodKey(),
		ResumeAgentSession: optionalBool(message.ResumeAgentSession),
	}
}

func hasLegacyPodRuntimeFields(message *podv1.CreatePodRequest) bool {
	return message.GetAgentSlug() != "" ||
		message.RunnerId != nil ||
		message.RepositoryId != nil ||
		message.Alias != nil ||
		message.AgentfileLayer != nil ||
		message.GetAutomationLevel() != "" ||
		message.Perpetual != nil ||
		len(message.GetKnowledgeMounts()) > 0 ||
		message.ModelResourceId != nil ||
		message.TokenBudget != nil
}
