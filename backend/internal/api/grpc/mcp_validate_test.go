package grpc

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
)

// MCP adapters short-circuit on invalid identifier/display payload before
// touching downstream services. Drive them with a bare *GRPCRunnerAdapter
// and nil dependencies; if validation regresses (caller-trust mode), the
// test panics on nil podOrchestrator/channelService/ticketService.

func TestMcpCreatePodRejectsInvalidPlanID(t *testing.T) {
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{
		OrganizationID: 1, OrganizationSlug: "team-alpha", UserID: 1,
	}
	payload := []byte(`{"plan_id":"not-a-uuid"}`)

	_, mcpErr := a.mcpCreatePod(context.Background(), tc, payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for invalid plan_id, got %v", mcpErr)
	}
}

func TestMcpCreatePodRejectsLegacyRuntimeFields(t *testing.T) {
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{
		OrganizationID: 1, OrganizationSlug: "team-alpha", UserID: 1,
	}
	payload := []byte(
		`{"plan_id":"11111111-1111-4111-8111-111111111111","agent_slug":"codex-cli"}`,
	)

	_, mcpErr := a.mcpCreatePod(context.Background(), tc, payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for legacy runtime fields, got %v", mcpErr)
	}
}

func TestMcpCreateWorkflowRequiresResourceApplyBeforeNameValidation(t *testing.T) {
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":"  ","prompt_template":"do work"}`)

	_, mcpErr := a.mcpCreateWorkflow(context.Background(), tc, "1-standalone-abcd0003", payload)
	if mcpErr == nil || mcpErr.code != 409 {
		t.Fatalf("expected 409 resource apply gate, got %v", mcpErr)
	}
}

func TestMcpCreateWorkflowRequiresResourceApplyBeforePromptValidation(t *testing.T) {
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":"daily-review","prompt_template":""}`)

	_, mcpErr := a.mcpCreateWorkflow(context.Background(), tc, "1-standalone-abcd0004", payload)
	if mcpErr == nil || mcpErr.code != 409 {
		t.Fatalf("expected 409 resource apply gate, got %v", mcpErr)
	}
}

func TestMcpCreateChannel_RejectsEmptyName(t *testing.T) {
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":""}`)

	_, mcpErr := a.mcpCreateChannel(context.Background(), tc, "1-standalone-abcd0001", payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for empty name, got %v", mcpErr)
	}
}

func TestMcpCreateChannel_RejectsZeroWidthOnlyName(t *testing.T) {
	// "​" (ZWSP) sanitizes to empty → displaykit returns ErrEmpty.
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":"​​"}`)

	_, mcpErr := a.mcpCreateChannel(context.Background(), tc, "1-standalone-abcd0002", payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for zero-width-only name, got %v", mcpErr)
	}
}

func TestMcpCreateTicket_RejectsEmptyTitle(t *testing.T) {
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"title":""}`)

	_, mcpErr := a.mcpCreateTicket(context.Background(), tc, payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for empty title, got %v", mcpErr)
	}
}

func TestMcpCreateTicket_RejectsRTLOverrideOnlyTitle(t *testing.T) {
	// "‮" (RTL override) sanitizes to empty.
	a := &GRPCRunnerAdapter{}
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"title":"‮"}`)

	_, mcpErr := a.mcpCreateTicket(context.Background(), tc, payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for RTL-only title, got %v", mcpErr)
	}
}
