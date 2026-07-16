package v1

import (
	"sync"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
)

// PodHandler handles pod-related requests.
// Pod creation is delegated to PodOrchestrator (service layer).
// This handler remains responsible for CRUD and HTTP protocol adaptation.
type PodHandler struct {
	podService         PodServiceForHandler       // Pod CRUD operations (ListPods, GetPod, TerminatePod, etc.)
	runnerService      *runner.Service            // Runner management
	podCoordinator     *runner.PodCoordinator     // Pod coordination (TerminatePod, terminal routing)
	orchestrator       *agentpod.PodOrchestrator  // Unified Pod creation logic
	commandSender      runner.RunnerCommandSender // Unified command sender (PTY + ACP)
	grantService       *grantservice.Service      // Resource grant/sharing service
	pendingQueue       pendingQueueReader
	sandboxFs          podWorkspaceSandbox
	workspaceArtifacts podWorkspaceArtifactTransfer
	artifactTransfers  sync.Map

	// Preview (Gateway HTTP data plane) dependencies.
	relaySelector       previewRelaySelector
	relayTokens         previewTokenGenerator
	previewPublicOrigin string
}

// PodHandlerOption is a functional option for configuring PodHandler
type PodHandlerOption func(*PodHandler)

// WithPodCoordinator sets the pod coordinator
func WithPodCoordinator(pc *runner.PodCoordinator) PodHandlerOption {
	return func(h *PodHandler) {
		h.podCoordinator = pc
	}
}

// WithPodService sets the pod service (for testing with mock implementations)
func WithPodService(ps PodServiceForHandler) PodHandlerOption {
	return func(h *PodHandler) {
		h.podService = ps
	}
}

// WithCommandSender sets the unified command sender for PTY and ACP commands
func WithCommandSender(sender runner.RunnerCommandSender) PodHandlerOption {
	return func(h *PodHandler) {
		h.commandSender = sender
	}
}

// WithGrantServiceForPod sets the grant service for resource sharing
func WithGrantServiceForPod(gs *grantservice.Service) PodHandlerOption {
	return func(h *PodHandler) {
		h.grantService = gs
	}
}

func WithPendingQueue(q pendingQueueReader) PodHandlerOption {
	return func(h *PodHandler) {
		h.pendingQueue = q
	}
}

func WithPodWorkspaceSandbox(sandbox podWorkspaceSandbox) PodHandlerOption {
	return func(h *PodHandler) {
		h.sandboxFs = sandbox
	}
}

func WithPodWorkspaceArtifactTransfer(transfer podWorkspaceArtifactTransfer) PodHandlerOption {
	return func(h *PodHandler) {
		h.workspaceArtifacts = transfer
	}
}

// WithRelayPreview wires the dependencies needed by the pod preview endpoint.
func WithRelayPreview(sel previewRelaySelector, tokens previewTokenGenerator, publicOrigin string) PodHandlerOption {
	return func(h *PodHandler) {
		h.relaySelector = sel
		h.relayTokens = tokens
		h.previewPublicOrigin = publicOrigin
	}
}

// NewPodHandler creates a new pod handler with required dependencies and optional configurations.
func NewPodHandler(
	podService *agentpod.PodService,
	runnerService *runner.Service,
	orchestrator *agentpod.PodOrchestrator,
	opts ...PodHandlerOption,
) *PodHandler {
	h := &PodHandler{
		podService:    podService,
		runnerService: runnerService,
		orchestrator:  orchestrator,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
