package goalloop

import (
	"encoding/json"
	"strings"
)

const PendingPodCleanupErrorPrefix = "pod-cleanup-pending:"

type PendingPodCleanup struct {
	TargetStatus string `json:"target_status"`
	Reason       string `json:"reason"`
	StopError    string `json:"stop_error"`
}

func EncodePendingPodCleanup(targetStatus, reason, stopError string) string {
	state := PendingPodCleanup{
		TargetStatus: targetStatus,
		Reason:       strings.TrimSpace(reason),
		StopError:    strings.TrimSpace(stopError),
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return PendingPodCleanupErrorPrefix + string(encoded)
}

func ParsePendingPodCleanup(value *string) (PendingPodCleanup, bool) {
	if value == nil || !strings.HasPrefix(*value, PendingPodCleanupErrorPrefix) {
		return PendingPodCleanup{}, false
	}
	var state PendingPodCleanup
	if err := json.Unmarshal(
		[]byte(strings.TrimPrefix(*value, PendingPodCleanupErrorPrefix)),
		&state,
	); err != nil {
		return PendingPodCleanup{}, false
	}
	if !validCleanupTarget(state.TargetStatus) || strings.TrimSpace(state.Reason) == "" {
		return PendingPodCleanup{}, false
	}
	return state, true
}

func (l *GoalLoop) PendingPodCleanup() (PendingPodCleanup, bool) {
	if l.PodKey == nil {
		return PendingPodCleanup{}, false
	}
	return ParsePendingPodCleanup(l.VerificationError)
}

func validCleanupTarget(status string) bool {
	return status == StatusPaused ||
		status == StatusCompleted ||
		status == StatusFailed ||
		status == StatusCancelled
}
