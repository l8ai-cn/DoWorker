package agentpod

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

var (
	ErrPodNotFound       = errors.New("pod not found")
	ErrPodNotTerminal    = errors.New("pod is not in a terminal state")
	ErrNoAvailableRunner = errors.New("no available runner")
	ErrRunnerNotFound    = errors.New("runner not found")
	ErrRunnerOffline     = errors.New("runner is offline")
	// ErrSandboxAlreadyResumed is re-exported from domain for backward compatibility.
	ErrSandboxAlreadyResumed = agentpod.ErrSandboxAlreadyResumed
)

// PodService handles pod operations
type PodService struct {
	repo           agentpod.PodRepository
	eventPublisher EventPublisher
}

// SetEventPublisher sets the event publisher for the service
func (s *PodService) SetEventPublisher(publisher EventPublisher) {
	s.eventPublisher = publisher
}

// NewPodService creates a new pod service
func NewPodService(repo agentpod.PodRepository) *PodService {
	return &PodService{repo: repo}
}

// CreatePod creates a new pod
func (s *PodService) CreatePod(ctx context.Context, req *CreatePodRequest) (*agentpod.Pod, error) {
	previewPath, err := normalizeInitialPreviewPath(req)
	if err != nil {
		return nil, err
	}
	if err := validateAgentfileLayerSecrets(req.AgentfileLayer); err != nil {
		return nil, err
	}
	keyBytes := make([]byte, 4)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	randomSuffix := hex.EncodeToString(keyBytes)

	ticketPart := "standalone"
	if req.TicketID != nil {
		ticketPart = fmt.Sprintf("%d", *req.TicketID)
	}
	podKey := fmt.Sprintf("%d-%s-%s", req.CreatedByID, ticketPart, randomSuffix)

	var modelPtr *string
	if req.Model != "" {
		m := req.Model
		modelPtr = &m
	}

	var permissionModePtr *string
	if req.PermissionMode != "" {
		pm := req.PermissionMode
		permissionModePtr = &pm
	}
	// Handle session ID
	var sessionID *string
	if req.SessionID != "" {
		sessionID = &req.SessionID
	}

	// Handle source pod key for resume
	var sourcePodKey *string
	if req.SourcePodKey != "" {
		sourcePodKey = &req.SourcePodKey
	}

	// Resolve interaction mode with default
	interactionMode := req.InteractionMode
	if interactionMode == "" {
		interactionMode = agentpod.InteractionModePTY
	}

	initialStatus := req.InitialStatus
	if initialStatus == "" {
		initialStatus = agentpod.StatusInitializing
	}

	pod := &agentpod.Pod{
		OrganizationID:  req.OrganizationID,
		PodKey:          podKey,
		RunnerID:        req.RunnerID,
		ClusterID:       req.ClusterID,
		AgentSlug:       req.AgentSlug,
		RepositoryID:    req.RepositoryID,
		TicketID:        req.TicketID,
		CreatedByID:     req.CreatedByID,
		Status:          initialStatus,
		AgentStatus:     agentpod.AgentStatusIdle,
		Prompt:          req.Prompt,
		Alias:           req.Alias,
		BranchName:      req.BranchName,
		Model:           modelPtr,
		PermissionMode:  permissionModePtr,
		SessionID:       sessionID,
		SourcePodKey:    sourcePodKey,
		InteractionMode: interactionMode,
		AutomationLevel: agentpod.NormalizeAutomationLevel(req.AutomationLevel),
		Perpetual:       req.Perpetual,
		ResolvedConfig:  req.ResolvedConfig,
		PodResourceBindings: agentpod.PodResourceBindings{
			VirtualAPIKeyID:             req.VirtualAPIKeyID,
			ModelResourceID:             req.ModelResourceID,
			WorkerSpecSnapshotID:        req.WorkerSpecSnapshotID,
			OrchestrationWorkerLaunchID: req.OrchestrationWorkerLaunchID,
		},
		PreviewPort: req.PreviewPort,
		PreviewPath: previewPath,
	}

	revision, err := newInitialPodConfigRevision(req, previewPath)
	if err != nil {
		return nil, err
	}
	pod, err = s.persistOrReuseWorkerLaunchPod(ctx, req, pod, revision)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "pod created", "pod_key", pod.PodKey, "org_id", pod.OrganizationID, "agent_slug", pod.AgentSlug, "runner_id", pod.RunnerID)

	// NOTE: current_pods increment is handled by PodCoordinator.CreatePod(),
	// which runs only when a command is actually sent to Runner.
	// Do NOT increment here to avoid double-counting.

	return pod, nil
}

// CreatePodForTicket creates a pod with ticket context
func (s *PodService) CreatePodForTicket(ctx context.Context, req *CreatePodRequest) (*agentpod.Pod, error) {
	if req.TicketID == nil {
		return nil, errors.New("ticket_id is required")
	}

	slug, title, err := s.repo.GetTicketByID(ctx, *req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	if req.Prompt == "" {
		req.Prompt = fmt.Sprintf("Working on ticket: %s\nTitle: %s", slug, title)
	}

	return s.CreatePod(ctx, req)
}
