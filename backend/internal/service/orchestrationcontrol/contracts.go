package orchestrationcontrol

import (
	"context"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

type ResourceListFilter struct {
	Kind              string
	Limit             int
	Offset            int
	EnvironmentBundle *EnvironmentBundleReferenceFilter
}

type ResourceListPage struct {
	Items []orchestrationcontrol.ResourceHead
	Total int64
}

type LockedApplyState struct {
	Plan             orchestrationcontrol.Plan
	Head             *orchestrationcontrol.ResourceHead
	CurrentRevision  *orchestrationcontrol.ResourceRevision
	ResultResourceID int64
	ResultIdentity   orchestrationcontrol.ResourceIdentity
	AppliedAt        time.Time
}

type ApplyMutation struct {
	Head           orchestrationcontrol.ResourceHead
	Revision       orchestrationcontrol.ResourceRevision
	ArtifactDigest string
}

type ApplyBuilder func(LockedApplyState) (ApplyMutation, error)

type Repository interface {
	GetResource(
		context.Context,
		orchestrationcontrol.Scope,
		orchestrationcontrol.ResourceTarget,
	) (orchestrationcontrol.ResourceHead, error)
	ListResources(
		context.Context,
		orchestrationcontrol.Scope,
		ResourceListFilter,
	) (ResourceListPage, error)
	GetRevision(
		context.Context,
		orchestrationcontrol.Scope,
		int64,
		int64,
	) (orchestrationcontrol.ResourceRevision, error)
	ListRevisions(
		context.Context,
		orchestrationcontrol.Scope,
		int64,
		int,
		int,
	) ([]orchestrationcontrol.ResourceRevision, error)
	CreatePlan(context.Context, orchestrationcontrol.Plan) error
	GetPlan(
		context.Context,
		orchestrationcontrol.Scope,
		string,
	) (orchestrationcontrol.Plan, error)
	RunApplyTransaction(
		context.Context,
		orchestrationcontrol.Scope,
		string,
		ApplyBuilder,
	) (orchestrationcontrol.ResourceHead, error)
}
