package runner

import (
	"fmt"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

type sandboxWorkspaceOperation func(
	*sandboxWorkspace,
) (*runnerv1.SandboxFsResultEvent, error)

func (pod *Pod) pinWorkspace() error {
	path, err := podWorkspaceRoot(pod)
	if err != nil {
		return err
	}
	workspace, err := openSandboxWorkspace(path)
	if err != nil {
		return err
	}
	if err := workspace.pinForPod(nil); err != nil {
		workspace.Close()
		return err
	}
	pod.workspaceMu.Lock()
	previous := pod.workspace
	pod.workspace = workspace
	pod.workspaceMu.Unlock()
	previous.Close()
	return nil
}

func (pod *Pod) withWorkspace(
	operation sandboxWorkspaceOperation,
) (*runnerv1.SandboxFsResultEvent, error) {
	if pod == nil {
		return nil, fmt.Errorf("pod not found")
	}
	pod.workspaceMu.RLock()
	defer pod.workspaceMu.RUnlock()
	if pod.workspace == nil {
		return nil, fmt.Errorf("workspace not configured")
	}
	workspace, closeAfter, err := pod.workspace.workspaceForOperation()
	if err != nil {
		return nil, err
	}
	if closeAfter {
		defer workspace.Close()
	}
	return operation(workspace)
}

func (pod *Pod) closeWorkspace() {
	if pod == nil {
		return
	}
	pod.workspaceMu.Lock()
	workspace := pod.workspace
	pod.workspace = nil
	pod.workspaceMu.Unlock()
	workspace.Close()
}
