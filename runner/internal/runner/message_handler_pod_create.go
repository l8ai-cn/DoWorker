package runner

import (
	"errors"
	"fmt"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

func (h *RunnerMessageHandler) OnCreatePod(cmd *runnerv1.CreatePodCommand) (err error) {
	log := logger.Pod()
	log.Info("Creating pod", "pod_key", cmd.PodKey, "command", cmd.LaunchCommand,
		"args", cmd.LaunchArgs)

	if existing, ok := h.podStore.Get(cmd.PodKey); ok && existing != nil {
		status := existing.GetStatus()
		if status != PodStatusStopped && status != PodStatusFailed {
			log.Info("duplicate create_pod absorbed", "pod_key", cmd.PodKey, "status", status)
			if status == PodStatusRunning {
				pid := 0
				if existing.IO != nil {
					pid = existing.IO.GetPID()
				}
				h.sendPodCreated(cmd.PodKey, pid, existing.SandboxPath, existing.Branch, uint16(cmd.Cols), uint16(cmd.Rows))
			}
			return nil
		}
	}

	ctx := h.runner.GetRunContext()
	h.podStore.Put(cmd.PodKey, &Pod{
		PodKey: cmd.PodKey,
		Status: PodStatusInitializing,
	})
	var registeredMCP MCPServer
	defer func() {
		if recovered := recover(); recovered != nil {
			h.podStore.Delete(cmd.PodKey)
			err = fmt.Errorf("pod build panic: %v", recovered)
			h.sendPodError(cmd.PodKey, err.Error())
		}
		if err != nil && registeredMCP != nil {
			registeredMCP.UnregisterPod(cmd.PodKey)
		}
	}()

	_ = h.conn.SendPodInitProgress(cmd.PodKey, "received", 1, "Pod command received by runner")

	cols, rows := podTerminalSize(cmd)
	cfg := h.runner.GetConfig()
	builder := h.runner.NewPodBuilder().
		WithCommand(cmd).
		WithPtySize(cols, rows).
		WithOSCHandler(h.createOSCHandler(cmd.PodKey))
	if cfg.LogPTY {
		builder.WithPTYLogging(cfg.GetLogPTYDir())
	}

	pod, err := builder.Build(ctx)
	if err != nil {
		h.podStore.Delete(cmd.PodKey)
		var podErr *client.PodError
		if errors.As(err, &podErr) {
			h.sendPodErrorWithCode(cmd.PodKey, podErr)
		} else {
			h.sendPodError(cmd.PodKey, fmt.Sprintf("failed to build pod: %v", err))
		}
		return fmt.Errorf("failed to build pod: %w", err)
	}

	if _, ok := h.podStore.Get(cmd.PodKey); !ok {
		log.InfoContext(ctx, "Pod was terminated during build, cleaning up", "pod_key", cmd.PodKey)
		pod.closeWorkspace()
		if pod.IO != nil {
			pod.IO.Teardown()
			pod.IO.Stop()
		}
		if pod.SandboxPath != "" {
			cleanupErr := h.removePodSandbox(cmd.PodKey, pod.SandboxPath)
			return errors.Join(fmt.Errorf("pod %s was terminated during build", cmd.PodKey), cleanupErr)
		}
		return fmt.Errorf("pod %s was terminated during build", cmd.PodKey)
	}

	h.podStore.Put(cmd.PodKey, pod)
	registeredMCP = h.runner.GetMCPServer()
	if registeredMCP != nil {
		registeredMCP.RegisterPod(
			cmd.PodKey,
			h.conn.GetOrgSlug(),
			nil,
			nil,
			cmd.LaunchCommand,
		)
	}
	if pod.IsACPMode() {
		if err := h.wireAndStartACPPod(pod, cmd, cols, rows); err != nil {
			return err
		}
	} else if err := h.wireAndStartPTYPod(pod, cmd, cols, rows); err != nil {
		return err
	}

	return nil
}

func podTerminalSize(cmd *runnerv1.CreatePodCommand) (int, int) {
	cols := int(cmd.Cols)
	rows := int(cmd.Rows)
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return cols, rows
}

func (h *RunnerMessageHandler) wireAndStartPTYPod(
	pod *Pod,
	cmd *runnerv1.CreatePodCommand,
	cols, rows int,
) error {
	pod.IO.SetExitHandler(h.createExitHandler(cmd.PodKey))
	pod.IO.SetIOErrorHandler(h.createPTYErrorHandler(cmd.PodKey, pod))

	if err := pod.IO.Start(); err != nil {
		h.podStore.Delete(cmd.PodKey)
		pod.closeWorkspace()
		if pod.IO != nil {
			pod.IO.Teardown()
		}
		if pod.SandboxPath != "" {
			err = errors.Join(err, h.removePodSandbox(cmd.PodKey, pod.SandboxPath))
		}
		h.sendPodError(cmd.PodKey, fmt.Sprintf("failed to start terminal: %v", err))
		return fmt.Errorf("failed to start terminal: %w", err)
	}

	pod.SetStatus(PodStatusRunning)
	if agentMon := h.runner.GetAgentMonitor(); agentMon != nil {
		agentMon.RegisterPod(cmd.PodKey, pod.IO.GetPID())
	}
	pod.SubscribeAgentStatusBridge(h.conn.SendAgentStatus)
	h.sendPodCreated(cmd.PodKey, pod.IO.GetPID(), pod.SandboxPath, pod.Branch, uint16(cols), uint16(rows))
	logger.Pod().Info(
		"Pod created (PTY)",
		"pod_key", cmd.PodKey,
		"pid", pod.IO.GetPID(),
		"sandbox", pod.SandboxPath,
	)
	return nil
}
