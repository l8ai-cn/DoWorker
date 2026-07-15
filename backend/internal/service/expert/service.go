package expert

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	itemservice "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

type PodLoader interface {
	GetPod(ctx context.Context, podKey string) (*agentpod.Pod, error)
}

type PodDispatcher interface {
	CreatePod(ctx context.Context, req *agentpodSvc.OrchestrateCreatePodRequest) (*agentpodSvc.OrchestrateCreatePodResult, error)
}

type RepoResolver interface {
	GetByID(ctx context.Context, id int64) (*gitprovider.Repository, error)
}

type WorkerSpecSnapshotLoader interface {
	GetByID(
		context.Context,
		int64,
		int64,
	) (specdomain.Snapshot, error)
}

type Service struct {
	store       expertdom.Repository
	pods        PodLoader
	dispatch    PodDispatcher
	repos       RepoResolver
	workerSpecs WorkerSpecSnapshotLoader
	items       itemservice.PositionedAppender
	gitops      gitops.Service
	logger      *slog.Logger
}

type Deps struct {
	Store       expertdom.Repository
	Pods        PodLoader
	Dispatch    PodDispatcher
	Repos       RepoResolver
	WorkerSpecs WorkerSpecSnapshotLoader
	Items       itemservice.PositionedAppender
	// Gitops is the git-backing choke point (namespace am-experts). It may be
	// nil, in which case the service runs in DB-only mode (identical to the
	// pre-git-backing behavior).
	Gitops gitops.Service
	Logger *slog.Logger
}

func NewService(deps Deps) *Service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		store:       deps.Store,
		pods:        deps.Pods,
		dispatch:    deps.Dispatch,
		repos:       deps.Repos,
		workerSpecs: deps.WorkerSpecs,
		items:       deps.Items,
		gitops:      deps.Gitops,
		logger:      logger.With("component", "expert"),
	}
}

func encodeKnowledgeMounts(mounts []expertdom.KnowledgeMount) json.RawMessage {
	if len(mounts) == 0 {
		return json.RawMessage("[]")
	}
	b, err := json.Marshal(mounts)
	if err != nil {
		return json.RawMessage("[]")
	}
	return b
}

func encodeConfigOverrides(values map[string]interface{}) json.RawMessage {
	if len(values) == 0 {
		return json.RawMessage("{}")
	}
	b, err := json.Marshal(values)
	if err != nil {
		return json.RawMessage("{}")
	}
	return b
}
