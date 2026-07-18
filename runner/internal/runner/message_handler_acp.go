package runner

import (
	"encoding/json"
	"fmt"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/policy"
)

// wireAndStartACPPod creates the ACPClient with Relay-forwarding callbacks,
// wires it into the pod, and starts the subprocess.
func (h *RunnerMessageHandler) wireAndStartACPPod(pod *Pod, cmd *runnerv1.CreatePodCommand, cols, rows int) error {
	log := logger.Pod()
	podKey := cmd.PodKey
	conn := h.conn
	bridge := newAcpBackendBridge()
	workbenchForwarder, err := newACPWorkbenchForwarder(
		podKey,
		cmd.AdapterId,
		pod.WorkDir,
		conn,
	)
	if err != nil {
		h.abortACPPodStartup(cmd.PodKey, nil, pod.SandboxPath)
		h.sendPodError(
			cmd.PodKey,
			fmt.Sprintf("failed to initialize workbench artifacts: %v", err),
		)
		return fmt.Errorf("initialize workbench artifacts: %w", err)
	}
	pod.workbenchForwarder = workbenchForwarder

	// Pre-declare so callbacks can capture it (NewClient returns the same pointer).
	var acpClient *acp.ACPClient

	// Create ACPClient with event callbacks that forward via Relay.
	acpClient = acp.NewClient(acp.ClientConfig{
		Command:                 pod.LaunchCommand,
		Args:                    pod.LaunchArgs,
		WorkDir:                 pod.WorkDir,
		Env:                     pod.LaunchEnv,
		Logger:                  log.With("pod_key", podKey),
		TransportType:           cmd.AdapterId,
		ResumeExternalSessionID: resumeExternalSessionFromEnv(pod.LaunchEnv),
		Callbacks: acp.EventCallbacks{
			OnContentChunk: func(sessionID string, chunk acp.ContentChunk) {
				workbenchForwarder.content(sessionID, chunk)
				sendAcpViaRelay(pod, "contentChunk", sessionID, chunk)
				payload, _ := json.Marshal(chunk)
				_ = conn.SendAcpSessionEvent(podKey, "contentChunk", string(payload))
			},
			OnToolCallUpdate: func(sessionID string, update acp.ToolCallUpdate) {
				workbenchForwarder.toolUpdate(sessionID, update)
				sendAcpViaRelay(pod, "toolCallUpdate", sessionID, update)
				payload, _ := json.Marshal(update)
				_ = conn.SendAcpSessionEvent(podKey, "toolCallUpdate", string(payload))
			},
			OnToolCallResult: func(sessionID string, result acp.ToolCallResult) {
				workbenchForwarder.toolResult(sessionID, result)
				sendAcpViaRelay(pod, "toolCallResult", sessionID, result)
				payload, _ := json.Marshal(result)
				_ = conn.SendAcpSessionEvent(podKey, "toolCallResult", string(payload))
			},
			OnPlanUpdate: func(sessionID string, update acp.PlanUpdate) {
				workbenchForwarder.plan(sessionID, update)
				sendAcpViaRelay(pod, "planUpdate", sessionID, update)
			},
			OnThinkingUpdate: func(sessionID string, update acp.ThinkingUpdate) {
				workbenchForwarder.thinking(sessionID, update)
				sendAcpViaRelay(pod, "thinkingUpdate", sessionID, update)
				payload, _ := json.Marshal(update)
				_ = conn.SendAcpSessionEvent(podKey, "thinkingUpdate", string(payload))
			},
			OnPermissionRequest: func(req acp.PermissionRequest) {
				path := permissionPathFromArgs(req.ArgumentsJSON)
				switch policy.Evaluate(pod.PolicyRules, req.ToolName, path) {
				case policy.VerdictAllow:
					_ = acpClient.RespondToPermission(req.RequestID, true, nil)
					return
				case policy.VerdictDeny:
					_ = acpClient.RespondToPermission(req.RequestID, false, nil)
					return
				}
				acpClient.AddPendingPermission(req)
				workbenchForwarder.permission(req)
				sendAcpViaRelay(pod, "permissionRequest", req.SessionID, req)
				payload, _ := json.Marshal(req)
				_ = conn.SendAcpSessionEvent(podKey, "permissionRequest", string(payload))
			},
			OnUsage: func(_ string, usage acp.TurnUsage) {
				bridge.onUsage(podKey, usage)
			},
			OnStateChange: func(newState string) {
				workbenchForwarder.state(newState)
				backendStatus := mapACPState(newState)
				_ = conn.SendAgentStatus(podKey, backendStatus)
				payload, _ := json.Marshal(map[string]string{"state": newState})
				_ = conn.SendAcpSessionEvent(podKey, "sessionState", string(payload))
				if newState == acp.StateIdle {
					bridge.onStateIdle(h, conn, podKey, "")
				}
				sendAcpViaRelay(pod, "sessionState", "", map[string]string{"state": newState})
				// Notify PodIO subscribers (e.g. Autopilot StateDetectorCoordinator).
				if sa, ok := pod.IO.(SessionAccess); ok {
					sa.NotifyStateChange(newState)
				}
			},
			OnLog: func(level, message string) {
				workbenchForwarder.log(level, message)
				entry := map[string]string{"level": level, "message": message}
				sendAcpViaRelay(pod, "log", "", entry)
				payload, _ := json.Marshal(entry)
				_ = conn.SendAcpSessionEvent(podKey, "log", string(payload))
			},
			OnConfigChange: func(sessionID string, update acp.ConfigUpdate) {
				workbenchForwarder.configurationChanged(update)
				sendAcpViaRelay(pod, "configChanged", sessionID, update)
			},
			OnLoopalExt: func(sessionID, kind string, data json.RawMessage) {
				sendAcpViaRelay(pod, "loopal."+kind, sessionID, data)
			},
			OnSessionID: func(sessionID string) {
				if sessionID != "" {
					workbenchForwarder.sessionID(sessionID)
					_ = conn.SendExternalSessionCaptured(podKey, sessionID)
				}
			},
			OnExit: func(exitCode int) {
				h.handleACPExit(podKey, exitCode)
			},
		},
	})

	// Seed configuration from launch_args so the first snapshot reflects the
	// resolved AgentFile values (e.g. --permission-mode bypassPermissions),
	// not an empty placeholder. Falls back silently when args don't carry
	// these flags (codex/gemini variants).
	acpClient.SeedConfiguration(parseClaudeInitialConfig(pod.LaunchArgs))

	// Wire client into pod
	pod.IO = NewACPPodIO(acpClient, podKey)
	pod.Relay = NewACPPodRelay(podKey, acpClient, func(payload []byte) {
		h.handleAcpRelayCommand(pod, payload)
	})

	// Start the ACP client (launches subprocess, performs initialize handshake)
	if err := acpClient.Start(); err != nil {
		h.abortACPPodStartup(cmd.PodKey, acpClient, pod.SandboxPath)
		h.sendPodError(cmd.PodKey, fmt.Sprintf("failed to start ACP agent: %v", err))
		return fmt.Errorf("failed to start ACP agent: %w", err)
	}
	if len(cmd.DeclaredCapabilities) > 0 {
		acpClient.CalibrateDeclaredCapabilities(podKey, cmd.DeclaredCapabilities)
	}
	workbenchForwarder.sessionInitialized(acpClient.Configuration())

	pod.SetStatus(PodStatusRunning)

	// Broadcast seeded configuration so late-joining subscribers (and the
	// snapshot path) see the initial mode/model rather than empty defaults.
	if initialCfg := acpClient.Configuration(); initialCfg.PermissionMode != "" || initialCfg.Model != "" {
		sendAcpViaRelay(pod, "configChanged", "", acp.ConfigUpdate{
			PermissionMode: initialCfg.PermissionMode,
			Model:          initialCfg.Model,
		})
	}

	// Create a new ACP session with MCP servers config
	mcpPort := h.runner.GetConfig().GetMCPPort()
	mcpServers := acp.BuildMCPServersConfig(mcpPort, podKey)
	if err := acpClient.NewSession(mcpServers); err != nil {
		h.abortACPPodStartup(cmd.PodKey, acpClient, pod.SandboxPath)
		h.sendPodError(cmd.PodKey, fmt.Sprintf("failed to create ACP session: %v", err))
		return fmt.Errorf("failed to create ACP session: %w", err)
	}
	if sid := acpClient.SessionID(); sid != "" {
		_ = conn.SendExternalSessionCaptured(podKey, sid)
	}

	// Send prompt if provided.
	// Claude: sessionID is empty (first message triggers system/init asynchronously).
	// ACP/Codex: sessionID is already set by NewSession().
	// ACPClient.SendPrompt checks State() == Idle (guaranteed after Handshake).
	if cmd.Prompt != "" {
		// Echo user message so it appears in chat on all connected devices.
		sendAcpViaRelay(pod, "contentChunk", "", map[string]string{
			"text": cmd.Prompt, "role": "user",
		})
		workbenchForwarder.content("", acp.ContentChunk{Text: cmd.Prompt, Role: "user"})
		if err := acpSendPromptWhenReady(pod, cmd.Prompt); err != nil {
			log.Error("Failed to send prompt", "pod_key", podKey, "error", err)
		}
	}

	h.sendPodCreated(cmd.PodKey, 0, pod.SandboxPath, pod.Branch, uint16(cols), uint16(rows))
	log.Info("Pod created (ACP)", "pod_key", cmd.PodKey, "sandbox", pod.SandboxPath)
	return nil
}

func (h *RunnerMessageHandler) abortACPPodStartup(podKey string, acpClient *acp.ACPClient, sandboxPath string) {
	h.podStore.Delete(podKey)
	if acpClient != nil {
		acpClient.Stop()
	}
	if sandboxPath != "" {
		h.removePodSandbox(sandboxPath)
	}
}

// handleACPExit handles ACP subprocess exit.
func (h *RunnerMessageHandler) handleACPExit(podKey string, exitCode int) {
	if pod, ok := h.podStore.Get(podKey); ok && pod != nil {
		if acpIO, ok := pod.IO.(*ACPPodIO); ok {
			acpIO.ForceIdleIfBusy()
		}
	}
	logger.Pod().Info("ACP process exited", "pod_key", podKey, "exit_code", exitCode)
	h.cleanupPodExit(podKey, exitCode, false)
}
