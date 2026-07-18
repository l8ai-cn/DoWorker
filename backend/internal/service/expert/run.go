package expert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/agentfile/serialize"
	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

var (
	ErrExpertRepublishRequired = errors.New(
		"expert must be republished from a workerspec-backed pod",
	)
	ErrExpertDispatchUnavailable = errors.New(
		"expert pod dispatcher is not configured",
	)
	ErrExpertResourceBindingCorrupt = errors.New(
		"expert orchestration resource binding is corrupt",
	)
)

func (s *Service) buildAgentfileLayer(ctx context.Context, expert *expertdom.Expert) string {
	if expert.AgentfileLayer != nil && strings.TrimSpace(*expert.AgentfileLayer) != "" {
		return strings.TrimSpace(*expert.AgentfileLayer)
	}

	var lines []string
	if expert.InteractionMode == expertdom.InteractionModeACP {
		lines = append(lines, "MODE acp")
	}
	if expert.Prompt != nil && strings.TrimSpace(*expert.Prompt) != "" {
		lines = append(lines, fmt.Sprintf("PROMPT %s", serialize.QuoteString(strings.TrimSpace(*expert.Prompt))))
	}
	for _, bundle := range expert.UsedEnvBundles {
		if bundle == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("USE_ENV_BUNDLE %s", serialize.QuoteString(bundle)))
	}
	var configOverrides map[string]interface{}
	if len(expert.ConfigOverrides) > 0 {
		_ = json.Unmarshal(expert.ConfigOverrides, &configOverrides)
	}
	for k, v := range configOverrides {
		lines = append(lines, fmt.Sprintf("CONFIG %s = %s", k, serialize.FormatValue(v)))
	}
	if expert.RepositoryID != nil && s.repos != nil {
		repo, err := s.repos.GetByID(ctx, *expert.RepositoryID)
		if err == nil && repo != nil {
			lines = append(lines, fmt.Sprintf(`REPO "%s"`, repo.Slug))
			branch := ""
			if expert.BranchName != nil {
				branch = strings.TrimSpace(*expert.BranchName)
			}
			if branch == "" && repo.DefaultBranch != "" {
				branch = repo.DefaultBranch
			}
			if branch != "" {
				lines = append(lines, fmt.Sprintf(`BRANCH "%s"`, branch))
			}
		}
	}
	if len(expert.SkillSlugs) > 0 {
		lines = append(lines, fmt.Sprintf("SKILLS %s", strings.Join([]string(expert.SkillSlugs), ", ")))
	}
	mounts := expertdom.ParseKnowledgeMounts(expert.KnowledgeMounts)
	if len(mounts) > 0 {
		refs := make([]string, 0, len(mounts))
		for _, m := range mounts {
			if m.Slug == "" {
				continue
			}
			if strings.EqualFold(m.Mode, "rw") {
				refs = append(refs, fmt.Sprintf("%s [rw]", m.Slug))
			} else {
				refs = append(refs, m.Slug)
			}
		}
		if len(refs) > 0 {
			lines = append(lines, fmt.Sprintf("KNOWLEDGE %s", strings.Join(refs, ", ")))
		}
	}
	return strings.Join(lines, "\n")
}

type RunExpertRequest struct {
	OrganizationID int64
	UserID         int64
	ExpertSlug     string
	Alias          *string
	PromptOverride *string
	Cols           int32
	Rows           int32
}

type RunExpertResult struct {
	Pod     *agentpod.Pod
	Warning string
}

func (s *Service) Run(ctx context.Context, req *RunExpertRequest) (*RunExpertResult, error) {
	expert, err := s.store.GetBySlug(ctx, req.OrganizationID, req.ExpertSlug)
	if err != nil {
		return nil, err
	}
	promptOverride, err := resourceManagedPromptOverride(
		expert,
		req.PromptOverride,
	)
	if err != nil {
		return nil, err
	}
	if expert.WorkerSpecSnapshotID == nil ||
		*expert.WorkerSpecSnapshotID <= 0 {
		return nil, ErrExpertRepublishRequired
	}
	if s.dispatch == nil {
		return nil, ErrExpertDispatchUnavailable
	}
	prepareSession, err := s.prepareRunInitialMessage(
		ctx,
		req.OrganizationID,
		*expert.WorkerSpecSnapshotID,
		req.PromptOverride,
	)
	if err != nil {
		return nil, err
	}
	orchReq := &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:           req.OrganizationID,
		UserID:                   req.UserID,
		Alias:                    req.Alias,
		WorkerSpecSnapshotID:     workerSpecSnapshotPointer(*expert.WorkerSpecSnapshotID),
		WorkerSpecPromptOverride: promptOverride,
		Cols:                     req.Cols,
		Rows:                     req.Rows,
		SessionProvision:         &sessionDomain.ProvisionSpec{},
		PrepareSession:           prepareSession,
	}
	result, err := s.dispatch.CreatePod(ctx, orchReq)
	if err != nil {
		return nil, err
	}
	_ = s.store.RecordRun(ctx, req.OrganizationID, expert.ID, time.Now())
	return &RunExpertResult{Pod: result.Pod, Warning: result.Warning}, nil
}

func workerSpecSnapshotPointer(value int64) *int64 {
	return &value
}

func resourceManagedPromptOverride(
	expert *expertdom.Expert,
	requestOverride *string,
) (*string, error) {
	resourceManaged := expert.OrchestrationResourceID != nil ||
		expert.OrchestrationResourceRevision != nil
	if !resourceManaged {
		return requestOverride, nil
	}
	if expert.OrchestrationResourceID == nil ||
		*expert.OrchestrationResourceID <= 0 ||
		expert.OrchestrationResourceRevision == nil ||
		*expert.OrchestrationResourceRevision <= 0 ||
		expert.WorkerSpecSnapshotID == nil ||
		*expert.WorkerSpecSnapshotID <= 0 {
		return nil, ErrExpertResourceBindingCorrupt
	}
	if requestOverride != nil {
		return requestOverride, nil
	}
	return expert.Prompt, nil
}
