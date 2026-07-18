package runner

import (
	"fmt"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

func (h *RunnerMessageHandler) OnSandboxFs(cmd *runnerv1.SandboxFsCommand) error {
	log := logger.Pod()
	log.Info("sandbox_fs", "request_id", cmd.RequestId, "pod_key", cmd.PodKey, "op", cmd.Op, "path", cmd.Path)

	var result *runnerv1.SandboxFsResultEvent
	var err error

	switch cmd.Op {
	case "list_host":
		cfg := h.runner.GetConfig()
		entries, listErr := listHostWorkspaceEntries(cfg.WorkspaceRoot, cmd.Path)
		if listErr != nil {
			result = fsErrResult(listErr.Error())
		} else {
			result = &runnerv1.SandboxFsResultEvent{Entries: entries, WorkspaceRoot: cfg.WorkspaceRoot}
		}
	case "mkdir":
		if cmd.PodKey == "" {
			root := h.runner.GetConfig().WorkspaceRoot
			result, err = h.sandboxFsMkdir(root, cmd.Path)
		} else {
			result, err = h.sandboxFsForPod(cmd.PodKey, func(workspace *sandboxWorkspace) (*runnerv1.SandboxFsResultEvent, error) {
				return h.sandboxFsMkdirWorkspace(workspace, cmd.Path)
			})
		}
	default:
		result, err = h.sandboxFsForPod(cmd.PodKey, func(workspace *sandboxWorkspace) (*runnerv1.SandboxFsResultEvent, error) {
			return h.dispatchSandboxFsOpWorkspace(workspace, cmd)
		})
	}
	if err != nil {
		result = fsErrResult(err.Error())
	}
	result.RequestId = cmd.RequestId
	result.PodKey = cmd.PodKey
	return h.conn.SendSandboxFsResult(result)
}

func (h *RunnerMessageHandler) sandboxFsForPod(
	podKey string,
	fn sandboxWorkspaceOperation,
) (*runnerv1.SandboxFsResultEvent, error) {
	pod, ok := h.podStore.Get(podKey)
	if !ok {
		workspace, err := openDetachedPodWorkspace(h.runner.GetConfig(), podKey)
		if err != nil {
			return fsErrResult(err.Error()), nil
		}
		defer workspace.Close()
		return fn(workspace)
	}
	return pod.withWorkspace(fn)
}

func (h *RunnerMessageHandler) dispatchSandboxFsOp(
	workspaceRoot string,
	cmd *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.dispatchSandboxFsOpWorkspace(workspace, cmd)
}

func (h *RunnerMessageHandler) dispatchSandboxFsOpWorkspace(
	workspace *sandboxWorkspace,
	cmd *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	switch cmd.Op {
	case "list":
		return h.sandboxFsListWorkspace(workspace, cmd.Path)
	case "stat":
		return h.sandboxFsStatWorkspace(workspace, cmd.Path)
	case "read":
		return h.sandboxFsReadWorkspace(workspace, cmd.Path)
	case "read_bytes":
		return h.sandboxFsReadBytesWorkspace(
			workspace,
			cmd.Path,
			cmd.Offset,
			cmd.Length,
		)
	case "read_verified_bytes":
		request, err := parseVerifiedArtifactRead(cmd.Payload)
		if err != nil {
			return fsErrResult(err.Error()), nil
		}
		return h.sandboxFsReadVerifiedArtifactWorkspace(
			workspace,
			cmd.Path,
			cmd.Offset,
			cmd.Length,
			request,
		)
	case "write":
		return h.sandboxFsWriteWorkspace(workspace, cmd.Path, cmd.Payload)
	case "download":
		return h.sandboxFsDownloadWorkspace(
			h.runner.GetRunContext(),
			workspace,
			cmd.Path,
			cmd.Payload,
		)
	case "upload":
		return h.sandboxFsUploadWorkspace(
			h.runner.GetRunContext(),
			workspace,
			cmd.Path,
			cmd.Payload,
		)
	case "changes":
		return h.sandboxFsChangesWorkspace(workspace)
	case "diff":
		return h.sandboxFsDiffWorkspace(workspace, cmd.Path)
	case "search":
		return h.sandboxFsSearchWorkspace(workspace, cmd.Payload, cmd.IncludeGlob, cmd.ExcludeGlob)
	case "skill_discover":
		return h.sandboxFsWorkerSkillDiscoverWorkspace(workspace, cmd.Path)
	default:
		return fsErrResult(fmt.Sprintf("unknown op %q", cmd.Op)), nil
	}
}
