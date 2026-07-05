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
			result, err = h.sandboxFsForPod(cmd.PodKey, func(root string) (*runnerv1.SandboxFsResultEvent, error) {
				return h.sandboxFsMkdir(root, cmd.Path)
			})
		}
	default:
		result, err = h.sandboxFsForPod(cmd.PodKey, func(root string) (*runnerv1.SandboxFsResultEvent, error) {
			return h.dispatchSandboxFsOp(root, cmd)
		})
	}
	if err != nil {
		result = fsErrResult(err.Error())
	}
	result.RequestId = cmd.RequestId
	result.PodKey = cmd.PodKey
	return h.conn.SendSandboxFsResult(result)
}

func (h *RunnerMessageHandler) sandboxFsForPod(podKey string, fn func(string) (*runnerv1.SandboxFsResultEvent, error)) (*runnerv1.SandboxFsResultEvent, error) {
	pod, ok := h.podStore.Get(podKey)
	if !ok {
		return fsErrResult("pod not found"), nil
	}
	root, err := podWorkspaceRoot(pod)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	return fn(root)
}

func (h *RunnerMessageHandler) dispatchSandboxFsOp(root string, cmd *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error) {
	switch cmd.Op {
	case "list":
		return h.sandboxFsList(root, cmd.Path)
	case "read":
		return h.sandboxFsRead(root, cmd.Path)
	case "write":
		return h.sandboxFsWrite(root, cmd.Path, cmd.Payload)
	case "changes":
		return h.sandboxFsChanges(root)
	case "diff":
		return h.sandboxFsDiff(root, cmd.Path)
	case "search":
		return h.sandboxFsSearch(root, cmd.Payload, cmd.IncludeGlob, cmd.ExcludeGlob)
	default:
		return fsErrResult(fmt.Sprintf("unknown op %q", cmd.Op)), nil
	}
}
