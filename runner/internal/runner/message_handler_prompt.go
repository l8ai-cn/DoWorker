package runner

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

const (
	ptyPromptReadyTimeout = 30 * time.Second
	ptySubmitGap          = 80 * time.Millisecond
)

var ptyPromptWaitSequence atomic.Uint64

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
	err = nil
	if pod.IsACPMode() {
		sendAcpViaRelay(pod, "contentChunk", "", map[string]string{
			"text": cmd.Prompt, "role": "user",
		})
		err = acpSendPromptWhenReady(pod, cmd.Prompt)
		if err != nil {
			err = errors.Join(err, h.releasePromptCommand(cmd))
		}
	} else {
		var safeToRetry bool
		safeToRetry, err = sendPTYPromptWhenReady(pod, cmd.Prompt)
		if err != nil && safeToRetry {
			err = errors.Join(err, h.releasePromptCommand(cmd))
		}
	}
	return err
}

func sendPTYPromptWhenReady(pod *Pod, prompt string) (bool, error) {
	if err := waitForPTYPromptReady(pod); err != nil {
		return true, err
	}
	if err := pod.IO.SendInput(prompt); err != nil {
		return true, err
	}
	if terminal, ok := pod.IO.(TerminalAccess); ok {
		time.Sleep(ptySubmitGap)
		return false, terminal.SendKeys([]string{"enter"})
	}
	return false, nil
}

func waitForPTYPromptReady(pod *Pod) error {
	if pod.IO.GetAgentStatus() == "waiting" {
		return nil
	}
	ready := make(chan struct{}, 1)
	subscriptionID := fmt.Sprintf("send-prompt-ready-%d", ptyPromptWaitSequence.Add(1))
	pod.IO.SubscribeStateChange(subscriptionID, func(status string) {
		if status == "waiting" {
			select {
			case ready <- struct{}{}:
			default:
			}
		}
	})
	defer pod.IO.UnsubscribeStateChange(subscriptionID)
	if pod.IO.GetAgentStatus() == "waiting" {
		return nil
	}
	timer := time.NewTimer(ptyPromptReadyTimeout)
	defer timer.Stop()
	select {
	case <-ready:
		return nil
	case <-timer.C:
		return fmt.Errorf(
			"timeout waiting for PTY pod %s to accept prompt; status=%s",
			pod.PodKey,
			pod.IO.GetAgentStatus(),
		)
	}
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
	if errors.Is(err, acp.ErrPromptNotReady) {
		return true
	}
	return strings.Contains(err.Error(), "cannot send prompt in state")
}
