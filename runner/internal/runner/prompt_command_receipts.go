package runner

import runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"

func (h *RunnerMessageHandler) isDuplicatePrompt(
	command *runnerv1.SendPromptCommand,
) (bool, error) {
	if command.CommandId == "" {
		return false, nil
	}
	h.promptDedupMu.Lock()
	defer h.promptDedupMu.Unlock()
	if h.promptDedup == nil {
		h.promptDedup = make(map[string]*promptDedupRing)
	}
	ring := h.promptDedup[command.PodKey]
	if ring == nil {
		ring = newPromptDedupRing(32)
		h.promptDedup[command.PodKey] = ring
	}
	if ring.seen(command.CommandId) {
		return true, nil
	}
	if h.receipts != nil {
		claimed, err := h.receipts.ClaimPrompt(command.PodKey, command.CommandId)
		if err != nil {
			return false, err
		}
		if !claimed {
			return true, nil
		}
	}
	ring.add(command.CommandId)
	return false, nil
}

func (h *RunnerMessageHandler) releasePromptCommand(
	command *runnerv1.SendPromptCommand,
) error {
	if command.CommandId == "" {
		return nil
	}
	h.promptDedupMu.Lock()
	if ring := h.promptDedup[command.PodKey]; ring != nil {
		ring.remove(command.CommandId)
	}
	h.promptDedupMu.Unlock()
	if h.receipts != nil {
		return h.receipts.ReleasePrompt(command.PodKey, command.CommandId)
	}
	return nil
}
