package expert

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

var ErrPodAccessDenied = errors.New("pod access denied")

type PublishFromPodRequest struct {
	OrganizationID int64
	UserID         int64
	PodKey         string
	Name           string
	Slug           string
	Description    *string
	AgentfileLayer *string
	UsedEnvBundles []string
	SkillSlugs     []string
	KnowledgeMounts []expertdom.KnowledgeMount
}

func (s *Service) PublishFromPod(ctx context.Context, req *PublishFromPodRequest) (*expertdom.Expert, error) {
	if s.pods == nil {
		return nil, errors.New("pod loader not configured")
	}
	pod, err := s.pods.GetPod(ctx, req.PodKey)
	if err != nil {
		return nil, err
	}
	if pod.OrganizationID != req.OrganizationID {
		return nil, ErrPodAccessDenied
	}
	name := req.Name
	if name == "" && pod.Alias != nil {
		name = *pod.Alias
	}
	if name == "" {
		name = pod.PodKey
	}
	createReq := &CreateExpertRequest{
		OrganizationID:  req.OrganizationID,
		UserID:          req.UserID,
		Name:            name,
		Slug:            req.Slug,
		Description:     req.Description,
		AgentSlug:       pod.AgentSlug,
		RunnerID:        &pod.RunnerID,
		RepositoryID:    pod.RepositoryID,
		BranchName:      pod.BranchName,
		Prompt:          optionalString(pod.Prompt),
		InteractionMode: pod.InteractionMode,
		Perpetual:       pod.Perpetual,
		UsedEnvBundles:  req.UsedEnvBundles,
		SkillSlugs:      req.SkillSlugs,
		KnowledgeMounts: req.KnowledgeMounts,
		ConfigOverrides: mapFromConfigValues(pod.ResolvedConfig),
		AgentfileLayer:  req.AgentfileLayer,
		SourcePodKey:    &pod.PodKey,
	}
	return s.Create(ctx, createReq)
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func mapFromConfigValues(values agent.ConfigValues) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(values))
	for k, v := range values {
		out[k] = v
	}
	return out
}
