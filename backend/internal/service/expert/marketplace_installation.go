package expert

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/google/uuid"
)

var ErrMarketplaceInstallationInvalid = errors.New("invalid marketplace installation")

type MarketplaceInstallationRequest struct {
	InstallationID       string
	TargetOrganizationID int64
	ActorUserID          int64
	RuntimeSnapshot      json.RawMessage
}

type marketplaceExpertSnapshot struct {
	Name            string                     `json:"name"`
	Description     *string                    `json:"description"`
	AgentSlug       string                     `json:"agent_slug"`
	Prompt          *string                    `json:"prompt"`
	InteractionMode string                     `json:"interaction_mode"`
	AutomationLevel string                     `json:"automation_level"`
	Perpetual       bool                       `json:"perpetual"`
	UsedEnvBundles  []string                   `json:"used_env_bundles"`
	SkillSlugs      []string                   `json:"skill_slugs"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledge_mounts"`
	ConfigOverrides map[string]interface{}     `json:"config_overrides"`
	AgentfileLayer  *string                    `json:"agentfile_layer"`
}

func (s *Service) InstallMarketplaceExpert(
	ctx context.Context,
	request MarketplaceInstallationRequest,
) (*expertdom.Expert, bool, error) {
	slug, err := marketplaceInstallationSlug(request)
	if err != nil {
		return nil, false, err
	}
	existing, err := s.store.GetBySlug(ctx, request.TargetOrganizationID, slug)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, expertdom.ErrNotFound) {
		return nil, false, err
	}
	var snapshot marketplaceExpertSnapshot
	if json.Unmarshal(request.RuntimeSnapshot, &snapshot) != nil {
		return nil, false, ErrMarketplaceInstallationInvalid
	}
	row, err := s.Create(ctx, &CreateExpertRequest{
		OrganizationID:  request.TargetOrganizationID,
		UserID:          request.ActorUserID,
		Name:            snapshot.Name,
		Slug:            slug,
		Description:     snapshot.Description,
		AgentSlug:       snapshot.AgentSlug,
		Prompt:          snapshot.Prompt,
		InteractionMode: snapshot.InteractionMode,
		AutomationLevel: snapshot.AutomationLevel,
		Perpetual:       snapshot.Perpetual,
		UsedEnvBundles:  append([]string(nil), snapshot.UsedEnvBundles...),
		SkillSlugs:      append([]string(nil), snapshot.SkillSlugs...),
		KnowledgeMounts: append([]expertdom.KnowledgeMount(nil), snapshot.KnowledgeMounts...),
		ConfigOverrides: snapshot.ConfigOverrides,
		AgentfileLayer:  snapshot.AgentfileLayer,
	})
	if err != nil {
		existing, lookupErr := s.store.GetBySlug(
			ctx,
			request.TargetOrganizationID,
			slug,
		)
		if lookupErr == nil {
			return existing, true, nil
		}
	}
	return row, false, err
}

func marketplaceInstallationSlug(
	request MarketplaceInstallationRequest,
) (string, error) {
	if uuid.Validate(request.InstallationID) != nil ||
		request.TargetOrganizationID <= 0 ||
		request.ActorUserID <= 0 ||
		!json.Valid(request.RuntimeSnapshot) {
		return "", ErrMarketplaceInstallationInvalid
	}
	compact := strings.ReplaceAll(request.InstallationID, "-", "")
	return "market-" + compact, nil
}
