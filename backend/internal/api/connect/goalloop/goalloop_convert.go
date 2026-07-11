package goalloopconnect

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

func toProto(loop *domain.GoalLoop) (*goalloopv1.GoalLoop, error) {
	var criteria []string
	if err := json.Unmarshal(loop.AcceptanceCriteria, &criteria); err != nil {
		return nil, fmt.Errorf("decode acceptance criteria: %w", err)
	}
	return &goalloopv1.GoalLoop{
		Id:                          loop.ID,
		Slug:                        loop.Slug,
		Name:                        loop.Name,
		Description:                 loop.Description,
		WorkerSpecSnapshotId:        loop.WorkerSpecSnapshotID,
		Objective:                   loop.Objective,
		AcceptanceCriteria:          criteria,
		VerificationCommand:         loop.VerificationCommand,
		Status:                      loop.Status,
		PodKey:                      loop.PodKey,
		AutopilotControllerKey:      loop.AutopilotControllerKey,
		MaxIterations:               int32(loop.MaxIterations),
		TokenBudget:                 loop.TokenBudget,
		TimeoutMinutes:              int32(loop.TimeoutMinutes),
		NoProgressLimit:             int32(loop.NoProgressLimit),
		SameErrorLimit:              int32(loop.SameErrorLimit),
		EscalationPolicy:            loop.EscalationPolicy,
		VerificationExitCode:        int32Pointer(loop.VerificationExitCode),
		VerificationOutput:          loop.VerificationOutput,
		VerificationOutputTruncated: loop.VerificationOutputTruncated,
		VerificationError:           loop.VerificationError,
		StartedAt:                   formatTime(loop.StartedAt),
		VerifiedAt:                  formatTime(loop.VerifiedAt),
		CompletedAt:                 formatTime(loop.CompletedAt),
		CreatedAt:                   loop.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                   loop.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func optionalInt(value *int32) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

func formatTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	out := value.UTC().Format(time.RFC3339)
	return &out
}

func int32Pointer(value *int) *int32 {
	if value == nil {
		return nil
	}
	out := int32(*value)
	return &out
}
