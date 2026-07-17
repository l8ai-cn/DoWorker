package infra

import (
	"math"
	"strconv"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
)

func validateAgentWorkbenchAppend(
	request agentworkbench.AppendRequest,
) error {
	projection := request.Projection
	if !validAgentWorkbenchText(request.SessionID, 100) ||
		projection.SessionID != request.SessionID ||
		!validAgentWorkbenchText(projection.StreamEpoch, 100) ||
		!validAgentWorkbenchDigest(projection.Digest) ||
		len(request.Events) == 0 ||
		request.ExpectedRevision == math.MaxUint64 ||
		projection.Revision != request.ExpectedRevision+1 {
		return agentworkbench.ErrInvalidArgument
	}
	if err := validateAgentWorkbenchEvents(request); err != nil {
		return err
	}
	if err := validateAgentWorkbenchSources(request); err != nil {
		return err
	}
	for _, receipt := range request.Receipts {
		if receipt.SessionID != request.SessionID ||
			validateAgentWorkbenchReceipt(receipt) != nil {
			return agentworkbench.ErrInvalidArgument
		}
	}
	return nil
}

func validateAgentWorkbenchEvents(
	request agentworkbench.AppendRequest,
) error {
	for index, event := range request.Events {
		if event.SessionID != request.SessionID ||
			event.StreamEpoch != request.Projection.StreamEpoch ||
			event.Revision != request.Projection.Revision ||
			event.Sequence == 0 ||
			!validAgentWorkbenchDigest(event.Digest) ||
			event.CreatedAt.IsZero() ||
			!validOptionalAgentWorkbenchText(event.CausationCommandID, 100) {
			return agentworkbench.ErrInvalidArgument
		}
		if index > 0 && event.Sequence != request.Events[index-1].Sequence+1 {
			return agentworkbench.ErrInvalidArgument
		}
	}
	if request.Projection.LatestSequence !=
		request.Events[len(request.Events)-1].Sequence {
		return agentworkbench.ErrInvalidArgument
	}
	return nil
}

func validateAgentWorkbenchSources(
	request agentworkbench.AppendRequest,
) error {
	stableIDs := make(map[string]struct{}, len(request.Sources))
	sequences := make(map[string]struct{}, len(request.Sources))
	for _, source := range request.Sources {
		if source.SessionID != request.SessionID ||
			!validAgentWorkbenchText(source.StableEventID, 200) ||
			!validAgentWorkbenchText(source.RunnerSessionEpoch, 100) ||
			source.SourceSequence == 0 ||
			!validAgentWorkbenchDigest(source.PayloadDigest) {
			return agentworkbench.ErrInvalidArgument
		}
		sequenceKey := source.RunnerSessionEpoch + "\x00" +
			strconv.FormatUint(source.SourceSequence, 10)
		if _, exists := stableIDs[source.StableEventID]; exists {
			return agentworkbench.ErrInvalidArgument
		}
		if _, exists := sequences[sequenceKey]; exists {
			return agentworkbench.ErrInvalidArgument
		}
		stableIDs[source.StableEventID] = struct{}{}
		sequences[sequenceKey] = struct{}{}
	}
	return nil
}

func validateAgentWorkbenchCurrent(
	current *agentWorkbenchStateRecord,
	request agentworkbench.AppendRequest,
) error {
	if current == nil {
		if request.ExpectedRevision != 0 {
			return agentworkbench.ErrRevisionConflict
		}
		if request.Events[0].Sequence != 1 {
			return agentworkbench.ErrEventConflict
		}
		return nil
	}
	if current.Revision != request.ExpectedRevision {
		return agentworkbench.ErrRevisionConflict
	}
	if current.StreamEpoch != request.Projection.StreamEpoch {
		return agentworkbench.ErrStreamConflict
	}
	if current.LatestSequence == math.MaxUint64 ||
		request.Events[0].Sequence != current.LatestSequence+1 {
		return agentworkbench.ErrEventConflict
	}
	return nil
}

func validateAgentWorkbenchReceipt(
	receipt agentworkbench.CommandReceipt,
) error {
	if !validAgentWorkbenchText(receipt.SessionID, 100) ||
		!validAgentWorkbenchText(receipt.CommandID, 100) ||
		!validAgentWorkbenchDigest(receipt.PayloadDigest) ||
		receipt.State < agentworkbench.ReceiptStateReceived ||
		receipt.State > agentworkbench.ReceiptStateCancelled {
		return agentworkbench.ErrInvalidArgument
	}
	return nil
}

func validAgentWorkbenchText(value string, limit int) bool {
	return value != "" && len(value) <= limit && value == strings.TrimSpace(value)
}

func validOptionalAgentWorkbenchText(value *string, limit int) bool {
	return value == nil || validAgentWorkbenchText(*value, limit)
}

func validAgentWorkbenchDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, char := range value[7:] {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}
