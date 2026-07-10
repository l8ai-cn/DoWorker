package agentpod

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func newInitialPodConfigRevision(req *CreatePodRequest) (*agentpod.PodConfigRevision, error) {
	summary, err := NewSafeConfigSummary(newInitialConfigReferences(req), nil)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &agentpod.PodConfigRevision{
		Revision:        1,
		AgentfileLayer:  req.AgentfileLayer,
		Status:          agentpod.ConfigRevisionStatusActive,
		ConfigSummary:   summary,
		ModelResourceID: req.ModelResourceID,
		CreatedByID:     req.CreatedByID,
		AppliedAt:       &now,
	}, nil
}

func newInitialConfigReferences(req *CreatePodRequest) map[string]ConfigReference {
	references := map[string]ConfigReference{}
	if req.RepositoryID != nil {
		references["repository"] = ConfigReference{ID: *req.RepositoryID, Available: true}
	}
	if req.ModelResourceID != nil {
		references["model_resource"] = ConfigReference{ID: *req.ModelResourceID, Available: true}
	}
	return references
}
