package tools

import (
	"encoding/json"
	"testing"
)

func TestPodCreateRequest(t *testing.T) {
	req := PodCreateRequest{
		PlanID: "11111111-1111-4111-8111-111111111111",
	}

	if req.PlanID != "11111111-1111-4111-8111-111111111111" {
		t.Errorf("PlanID: got %v", req.PlanID)
	}
}

func TestPodCreateResponse(t *testing.T) {
	resp := PodCreateResponse{
		PodKey:      "new-pod",
		Status:      "created",
		TerminalURL: "ws://localhost:8080/terminal",
	}

	if resp.PodKey != "new-pod" {
		t.Errorf("PodKey: got %v, want %v", resp.PodKey, "new-pod")
	}
	if resp.Status != "created" {
		t.Errorf("Status: got %v, want %v", resp.Status, "created")
	}
}

func TestAgentFieldUnmarshalJSONString(t *testing.T) {
	var field AgentField
	err := field.UnmarshalJSON([]byte(`"claude-code"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(field) != "claude-code" {
		t.Errorf("AgentField: got %v, want claude-code", field)
	}
}

func TestAgentFieldUnmarshalJSONObject(t *testing.T) {
	var field AgentField
	err := field.UnmarshalJSON([]byte(`{"id": 1, "slug": "aider", "name": "Aider"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(field) != "aider" {
		t.Errorf("AgentField: got %v, want aider", field)
	}
}

func TestAgentFieldUnmarshalJSONInvalid(t *testing.T) {
	var field AgentField
	// Invalid JSON should not cause error, just ignore
	err := field.UnmarshalJSON([]byte(`invalid json`))
	if err != nil {
		t.Errorf("expected no error for invalid JSON, got: %v", err)
	}
}

func TestAvailablePodGetUsername(t *testing.T) {
	// Test with CreatedBy set
	pod := AvailablePod{
		PodKey: "test-pod",
		CreatedBy: &PodCreator{
			ID:       1,
			Username: "testuser",
			Name:     "Test User",
		},
	}
	if pod.GetUsername() != "testuser" {
		t.Errorf("GetUsername: got %v, want testuser", pod.GetUsername())
	}

	// Test with CreatedBy nil
	pod2 := AvailablePod{
		PodKey: "test-pod-2",
	}
	if pod2.GetUsername() != "" {
		t.Errorf("GetUsername: got %v, want empty string", pod2.GetUsername())
	}
}

func TestAvailablePodGetTicketTitle(t *testing.T) {
	// Test with Ticket set
	ticketID := 123
	pod := AvailablePod{
		PodKey:   "test-pod",
		TicketID: &ticketID,
		Ticket: &PodTicket{
			ID:    123,
			Slug:  "AM-123",
			Title: "Test Ticket Title",
		},
	}
	if pod.GetTicketTitle() != "Test Ticket Title" {
		t.Errorf("GetTicketTitle: got %v, want Test Ticket Title", pod.GetTicketTitle())
	}

	// Test with Ticket nil
	pod2 := AvailablePod{
		PodKey: "test-pod-2",
	}
	if pod2.GetTicketTitle() != "" {
		t.Errorf("GetTicketTitle: got %v, want empty string", pod2.GetTicketTitle())
	}
}
