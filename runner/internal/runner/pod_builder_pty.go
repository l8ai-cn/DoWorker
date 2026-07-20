package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/poddaemon"
	"github.com/anthropics/agentsmesh/runner/internal/terminal"
	"github.com/anthropics/agentsmesh/runner/internal/terminal/aggregator"
	"github.com/anthropics/agentsmesh/runner/internal/terminal/vt"
)

func (b *PodBuilder) buildPTYPod(
	ctx context.Context,
	sandboxRoot, workingDir, branchName string,
	resolvedArgs []string,
	envVars map[string]string,
	launchCommand string,
	workspace *sandboxWorkspace,
) (*Pod, error) {
	b.sendProgress("starting_pty", 80, "Starting terminal...")
	unsetEnv := gitProcessIsolationUnsetEnv(b.cmd.GetSandboxConfig().GetCredentialType())
	capturedEnv := buildMergedEnvWithout(envVars, unsetEnv)
	injectTraceparent(ctx, envVars)
	if traceparent, ok := envVars["TRACEPARENT"]; ok {
		capturedEnv = append(capturedEnv, "TRACEPARENT="+traceparent)
	}

	ptyFactory := b.podDaemonPTYFactory(
		sandboxRoot,
		workingDir,
		branchName,
		launchCommand,
	)
	term, err := terminal.New(terminal.Options{
		Command: launchCommand, Args: resolvedArgs, WorkDir: workingDir,
		Env: envVars, Rows: b.rows, Cols: b.cols, Label: b.cmd.PodKey,
		UnsetEnv: unsetEnv, PTYFactory: ptyFactory,
	})
	if err != nil {
		workspace.Close()
		podErr := &client.PodError{
			Code: client.ErrCodeCommandStart, Message: fmt.Sprintf("failed to create terminal: %v", err),
		}
		return nil, errors.Join(podErr, b.cleanupSandbox(ctx, sandboxRoot, "terminal creation error"))
	}

	virtualTerm := vt.NewVirtualTerminal(b.cols, b.rows, b.vtHistoryLimit)
	if b.oscHandler != nil {
		virtualTerm.SetOSCHandler(b.oscHandler)
	}
	agg := aggregator.NewSmartAggregator(nil, aggregator.WithFullRedrawThrottling())
	ptyLogger := b.newPTYLogger(ctx, agg)
	pod := &Pod{
		ID: b.cmd.PodKey, PodKey: b.cmd.PodKey, Agent: launchCommand,
		InteractionMode: InteractionModePTY, Branch: branchName,
		SandboxPath: sandboxRoot, LaunchCommand: launchCommand,
		LaunchArgs: resolvedArgs, WorkDir: workingDir, LaunchEnv: capturedEnv,
		Perpetual: b.cmd.Perpetual, StartedAt: time.Now(),
		Status: PodStatusInitializing, workspace: workspace,
		vtProvider: func() *vt.VirtualTerminal { return virtualTerm },
	}
	comps := &PTYComponents{
		Terminal: term, VirtualTerminal: virtualTerm, Aggregator: agg, PTYLogger: ptyLogger,
	}
	pod.IO = NewPTYPodIO(b.cmd.PodKey, comps, PTYPodIODeps{
		GetOrCreateDetector: pod.GetOrCreateStateDetector,
		SubscribeState:      pod.SubscribeStateChange, UnsubscribeState: pod.UnsubscribeStateChange,
		GetPTYError: pod.GetPTYError,
	})
	pod.Relay = NewPTYPodRelay(b.cmd.PodKey, pod.IO, comps)
	term.SetOutputHandler(NewPTYOutputHandler(b.cmd.PodKey, comps, pod.NotifyStateDetectorWithScreen))
	logger.Pod().InfoContext(ctx, "Pod built (PTY)", "pod_key", b.cmd.PodKey, "working_dir", workingDir)
	b.sendProgress("ready", 100, "Pod is ready")
	return pod, nil
}

func (b *PodBuilder) podDaemonPTYFactory(
	sandboxRoot, workingDir, branchName, launchCommand string,
) terminal.PTYFactory {
	if b.deps.PodDaemonManager == nil || sandboxRoot == "" {
		return nil
	}
	manager := b.deps.PodDaemonManager
	options := poddaemon.CreateOpts{
		PodKey: b.cmd.PodKey, Agent: launchCommand, SandboxPath: sandboxRoot,
		WorkDir: workingDir, RepositoryURL: b.cmd.GetSandboxConfig().GetHttpCloneUrl(),
		Branch: branchName, TicketSlug: b.cmd.GetSandboxConfig().GetTicketSlug(),
		VTHistoryLimit: b.vtHistoryLimit, Perpetual: b.cmd.Perpetual,
	}
	return func(command string, args []string, _ string, env []string, cols, rows int) (terminal.PtyProcess, error) {
		options.Command, options.Args, options.Env = command, args, env
		options.Cols, options.Rows = cols, rows
		process, _, err := manager.CreateSession(options)
		return process, err
	}
}

func (b *PodBuilder) newPTYLogger(
	ctx context.Context,
	agg *aggregator.SmartAggregator,
) *aggregator.PTYLogger {
	if !b.enablePTYLogging || b.ptyLogDir == "" {
		return nil
	}
	ptyLogger, err := aggregator.NewPTYLogger(b.ptyLogDir, b.cmd.PodKey)
	if err != nil {
		logger.Pod().WarnContext(ctx, "Failed to create PTY logger", "pod_key", b.cmd.PodKey, "error", err)
		return nil
	}
	agg.SetPTYLogger(ptyLogger)
	return ptyLogger
}
