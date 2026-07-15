package runner

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// The gap lets the TUI process the prompt body before receiving Enter.
const ptySubmitGap = 80 * time.Millisecond

func (h *RunnerMessageHandler) OnSendPrompt(cmd *runnerv1.SendPromptCommand) error {
	log := logger.Pod()
	duplicate, err := h.isDuplicatePrompt(cmd)
	if err != nil {
		return fmt.Errorf("reserve prompt command: %w", err)
	}
	if duplicate {
		log.Info("duplicate send_prompt absorbed", "pod_key", cmd.PodKey, "command_id", cmd.CommandId)
		return nil
	}
	pod, ok := h.podStore.Get(cmd.PodKey)
	if !ok {
		log.Warn("Pod not found for send_prompt", "pod_key", cmd.PodKey)
		return errors.Join(
			fmt.Errorf("pod not found: %s", cmd.PodKey),
			h.releasePromptCommand(cmd),
		)
	}
	if pod.IO == nil {
		log.Warn("PodIO not available for send_prompt", "pod_key", cmd.PodKey)
		return errors.Join(
			fmt.Errorf("pod IO not available: %s", cmd.PodKey),
			h.releasePromptCommand(cmd),
		)
	}
	if pod.IsACPMode() {
		sendAcpViaRelay(pod, "contentChunk", "", map[string]string{
			"text": cmd.Prompt, "role": "user",
		})
		if err := acpSendPromptWhenReady(pod, cmd.Prompt); err != nil {
			return errors.Join(err, h.releasePromptCommand(cmd))
		}
		return nil
	}
	if err := pod.IO.SendInput(cmd.Prompt); err != nil {
		return errors.Join(err, h.releasePromptCommand(cmd))
	}
	if terminal, ok := pod.IO.(TerminalAccess); ok {
		time.Sleep(ptySubmitGap)
		return terminal.SendKeys([]string{"enter"})
	}
	return nil
}

func acpSendPromptWhenReady(pod *Pod, prompt string) error {
	const attempts = 100
	for i := 0; i < attempts; i++ {
		err := pod.IO.SendInput(prompt)
		if err == nil {
			return nil
		}
		if acpPromptRetryable(err) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("timeout waiting for ACP pod %s to accept prompt", pod.PodKey)
}

func acpPromptRetryable(err error) bool {
	return errors.Is(err, acp.ErrPromptNotReady) ||
		strings.Contains(err.Error(), "cannot send prompt in state")
}
