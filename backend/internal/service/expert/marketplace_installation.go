package expert

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/google/uuid"
)

var ErrMarketplaceInstallationInvalid = errors.New("invalid marketplace installation")

type MarketplaceInstallationRequest struct {
	InstallationID            string
	TargetOrganizationID      int64
	TargetOrganizationSlug    string
	ActorUserID               int64
	ModelResourceID           int64
	ToolModelResourceIDs      map[string]int64
	SourceMarketApplicationID int64
	SourceMarketReleaseID     int64
	RuntimeSnapshot           json.RawMessage
}

type marketplaceRuntimeSnapshot struct {
	Version    int                  `json:"version"`
	Expert     marketExpertSnapshot `json:"expert"`
	WorkerSpec specdomain.Spec      `json:"worker_spec"`
}

func (s *Service) InstallMarketplaceExpert(
	ctx context.Context,
	request MarketplaceInstallationRequest,
) (*expertdom.Expert, bool, error) {
	slug, err := marketplaceInstallationSlug(request)
	if err != nil {
		return nil, false, err
	}
	if request.TargetOrganizationSlug == "" {
		return nil, false, ErrMarketplaceInstallationInvalid
	}
	if request.SourceMarketApplicationID > 0 {
		existing, lookupErr := s.store.GetByMarketApplication(
			ctx,
			request.TargetOrganizationID,
			request.SourceMarketApplicationID,
		)
		if lookupErr == nil {
			return existing, true, nil
		}
		if !errors.Is(lookupErr, expertdom.ErrNotFound) {
			return nil, false, lookupErr
		}
	}
	existing, err := s.store.GetBySlug(ctx, request.TargetOrganizationID, slug)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, expertdom.ErrNotFound) {
		return nil, false, err
	}
	if _, err := slugkit.NewFromTrusted(request.TargetOrganizationSlug); err != nil {
		return nil, false, ErrMarketplaceInstallationInvalid
	}
	var snapshot marketplaceRuntimeSnapshot
	if decodeStrictJSON(request.RuntimeSnapshot, &snapshot) != nil ||
		snapshot.Version != 1 ||
		snapshot.Expert.Version != 1 ||
		validateMarketExpertSnapshot(snapshot.Expert) != nil {
		return nil, false, ErrMarketplaceInstallationInvalid
	}
	workerSnapshotID, err := s.prepareMarketplaceWorkerSnapshot(
		ctx,
		request,
		snapshot,
	)
	if err != nil {
		return nil, false, err
	}
	expertSnapshot := snapshot.Expert
	var sourceApplicationID, sourceReleaseID *int64
	if request.SourceMarketApplicationID > 0 {
		sourceApplicationID = &request.SourceMarketApplicationID
		sourceReleaseID = &request.SourceMarketReleaseID
	}
	row, err := s.Create(ctx, &CreateExpertRequest{
		OrganizationID:            request.TargetOrganizationID,
		UserID:                    request.ActorUserID,
		Name:                      expertSnapshot.Name,
		Slug:                      slug,
		Description:               expertSnapshot.Description,
		AgentSlug:                 expertSnapshot.AgentSlug,
		Prompt:                    expertSnapshot.Prompt,
		InteractionMode:           expertSnapshot.InteractionMode,
		AutomationLevel:           expertSnapshot.AutomationLevel,
		Perpetual:                 expertSnapshot.Perpetual,
		UsedEnvBundles:            expertSnapshot.UsedEnvBundles,
		SkillSlugs:                expertSnapshot.SkillSlugs,
		KnowledgeMounts:           expertSnapshot.KnowledgeMounts,
		ConfigOverrides:           expertSnapshot.ConfigOverrides,
		WorkerSpecSnapshotID:      &workerSnapshotID,
		SourceMarketApplicationID: sourceApplicationID,
		SourceMarketReleaseID:     sourceReleaseID,
		Metadata:                  expertSnapshot.Metadata,
	})
	if err == nil {
		return row, false, nil
	}
	cleanupCtx, cancelCleanup := marketCleanupContext(ctx)
	defer cancelCleanup()
	if cleanupErr := s.removeUnusedMarketSnapshot(
		cleanupCtx,
		request.TargetOrganizationID,
		workerSnapshotID,
	); cleanupErr != nil {
		return nil, false, errors.Join(err, cleanupErr)
	}
	var lookupErr error
	if request.SourceMarketApplicationID > 0 {
		existing, lookupErr = s.store.GetByMarketApplication(
			ctx,
			request.TargetOrganizationID,
			request.SourceMarketApplicationID,
		)
	} else {
		existing, lookupErr = s.store.GetBySlug(
			ctx,
			request.TargetOrganizationID,
			slug,
		)
	}
	if lookupErr == nil {
		return existing, true, nil
	}
	return nil, false, err
}

func marketplaceInstallationSlug(
	request MarketplaceInstallationRequest,
) (string, error) {
	if uuid.Validate(request.InstallationID) != nil ||
		request.TargetOrganizationID <= 0 ||
		request.ActorUserID <= 0 ||
		request.ModelResourceID <= 0 ||
		(request.SourceMarketApplicationID == 0) !=
			(request.SourceMarketReleaseID == 0) ||
		request.SourceMarketApplicationID < 0 ||
		request.SourceMarketReleaseID < 0 ||
		!json.Valid(request.RuntimeSnapshot) {
		return "", ErrMarketplaceInstallationInvalid
	}
	compact := strings.ReplaceAll(request.InstallationID, "-", "")
	return "market-" + compact, nil
}
