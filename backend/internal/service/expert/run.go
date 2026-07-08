package expert

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/agentfile/serialize"
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

func knowledgeMountsForRun(expert *expertdom.Expert) []agentpodSvc.KnowledgeMountRequest {
	parsed := expertdom.ParseKnowledgeMounts(expert.KnowledgeMounts)
	out := make([]agentpodSvc.KnowledgeMountRequest, 0, len(parsed))
	for _, m := range parsed {
		if m.Slug == "" {
			continue
		}
		out = append(out, agentpodSvc.KnowledgeMountRequest{Slug: m.Slug, Mode: m.Mode})
	}
	return out
}

type RunExpertRequest struct {
	OrganizationID int64
	UserID         int64
	ExpertSlug     string
	Alias          *string
	PromptOverride *string
	RunnerID       *int64
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
	// Git is the source of truth: read agent.md from the repo first, refreshing
	// the DB cache when it lags. On miss/disabled/transient-error fall back to
	// the DB agentfile_layer cache / generator so Run never hard-fails.
	layer, fromGit := s.readAgentFileFromGit(ctx, expert)
	if fromGit {
		s.refreshAgentfileCache(ctx, expert, layer)
	} else {
		layer = s.buildAgentfileLayer(ctx, expert)
	}
	if req.PromptOverride != nil && strings.TrimSpace(*req.PromptOverride) != "" {
		override := fmt.Sprintf("PROMPT %s", serialize.QuoteString(strings.TrimSpace(*req.PromptOverride)))
		if layer == "" {
			layer = override
		} else {
			layer = override + "\n" + layer
		}
	}
	runnerID := int64(0)
	if req.RunnerID != nil {
		runnerID = *req.RunnerID
	} else if expert.RunnerID != nil {
		runnerID = *expert.RunnerID
	}
	orchReq := &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:  req.OrganizationID,
		UserID:          req.UserID,
		RunnerID:        runnerID,
		AgentSlug:       expert.AgentSlug,
		RepositoryID:    expert.RepositoryID,
		Alias:           req.Alias,
		AgentfileLayer:  &layer,
		Cols:            req.Cols,
		Rows:            req.Rows,
		Perpetual:       expert.Perpetual,
		KnowledgeMounts: knowledgeMountsForRun(expert),
	}
	result, err := s.dispatch.CreatePod(ctx, orchReq)
	if err != nil {
		return nil, err
	}
	_ = s.store.RecordRun(ctx, req.OrganizationID, expert.ID, time.Now())
	return &RunExpertResult{Pod: result.Pod, Warning: result.Warning}, nil
}
