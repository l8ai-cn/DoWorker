package runner

import (
	"sync"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/workbench"
)

type acpWorkbenchForwarder struct {
	podKey          string
	workDir         string
	mapper          *workbench.Mapper
	observer        *workbench.ArtifactObserver
	sender          client.ConnectionSender
	convertOffice   officePreviewConverter
	artifactMu      sync.Mutex
	previewMu       sync.Mutex
	converting      map[string]struct{}
	latestRevision  map[string]uint64
	activeCommandID string
}

func newACPWorkbenchForwarder(
	podKey, adapterID, workDir string,
	sender client.ConnectionSender,
) (*acpWorkbenchForwarder, error) {
	observer, err := workbench.NewArtifactObserver(workDir)
	if err != nil {
		return nil, err
	}
	return &acpWorkbenchForwarder{
		podKey:         podKey,
		workDir:        workDir,
		mapper:         workbench.NewMapper(podKey, adapterID),
		observer:       observer,
		sender:         sender,
		convertOffice:  convertOfficePreview,
		converting:     make(map[string]struct{}),
		latestRevision: make(map[string]uint64),
	}, nil
}

func (f *acpWorkbenchForwarder) content(sessionID string, chunk acp.ContentChunk) {
	f.send(f.mapper.ContentChunk(sessionID, chunk))
}

func (f *acpWorkbenchForwarder) toolUpdate(
	sessionID string,
	update acp.ToolCallUpdate,
) {
	f.send(f.mapper.ToolUpdate(sessionID, update))
}

func (f *acpWorkbenchForwarder) toolResult(
	sessionID string,
	result acp.ToolCallResult,
) {
	f.send(f.mapper.ToolResult(sessionID, result))
}

func (f *acpWorkbenchForwarder) plan(sessionID string, update acp.PlanUpdate) {
	f.send(f.mapper.Plan(sessionID, update))
}

func (f *acpWorkbenchForwarder) thinking(
	sessionID string,
	update acp.ThinkingUpdate,
) {
	f.send(f.mapper.Thinking(sessionID, update))
}

func (f *acpWorkbenchForwarder) permission(request acp.PermissionRequest) {
	f.send(f.mapper.Permission(request))
}

func (f *acpWorkbenchForwarder) sessionInitialized(configuration acp.Configuration) {
	f.send(f.mapper.SessionInitialized(configuration))
}

func (f *acpWorkbenchForwarder) configurationChanged(update acp.ConfigUpdate) {
	f.send(f.mapper.ConfigurationChanged(update))
}

func (f *acpWorkbenchForwarder) state(state string) {
	f.send(f.mapper.State(state))
	if state != acp.StateIdle {
		return
	}
	f.scanArtifacts()
}

func (f *acpWorkbenchForwarder) scanArtifacts() {
	f.artifactMu.Lock()
	defer f.artifactMu.Unlock()
	artifacts, err := f.observer.Scan()
	if err != nil {
		f.send(f.mapper.Unsupported("artifact.scan.error", map[string]string{
			"error": err.Error(),
		}))
		return
	}
	f.send(f.mapper.Artifacts(artifacts))
	for _, artifact := range artifacts {
		f.recordArtifactRevision(artifact)
		f.queueOfficePreview(artifact)
	}
}

func (f *acpWorkbenchForwarder) log(level, message string) {
	f.send(f.mapper.Log(level, message))
}

func (f *acpWorkbenchForwarder) sessionID(sessionID string) {
	f.mapper.SetExternalSessionID(sessionID)
}

func (f *acpWorkbenchForwarder) setActiveCommandID(commandID string) {
	f.previewMu.Lock()
	f.activeCommandID = commandID
	f.previewMu.Unlock()
}

func (f *acpWorkbenchForwarder) currentCommandID() string {
	f.previewMu.Lock()
	defer f.previewMu.Unlock()
	return f.activeCommandID
}

func (f *acpWorkbenchForwarder) send(
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) {
	if batch == nil || batch.GetPodKey() == "" {
		return
	}
	if err := f.sender.SendMessage(&runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_WorkbenchEvents{
			WorkbenchEvents: batch,
		},
	}); err != nil {
		logger.Pod().Error(
			"failed to send workbench events",
			"pod_key",
			f.podKey,
			"error",
			err,
		)
	}
}
