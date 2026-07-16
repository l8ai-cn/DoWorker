package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

func (service *Service) Plan(
	ctx context.Context,
	request PlanRequest,
) (PlanResult, error) {
	draft, err := service.validateDraft(ctx, request)
	if err != nil {
		return PlanResult{}, err
	}
	result := PlanResult{ValidationResult: draft.result}
	if hasBlockingIssues(result.Issues) {
		return result, nil
	}
	current, err := service.loadCurrentRevision(ctx, request.Scope, draft)
	if err != nil {
		return PlanResult{}, err
	}
	resolved, err := service.resolveReferences(ctx, request.Scope, draft)
	if err != nil {
		return PlanResult{}, err
	}
	output, err := draft.planner.Plan(ctx, TargetPlanInput{
		Scope: request.Scope, Operation: draft.result.Operation,
		Manifest: draft.manifest, TypedSpec: draft.typedSpec,
		Head: draft.head, CurrentRevision: current,
		ResolvedReferences: resolved,
	})
	if errors.Is(err, ErrStaleOptions) {
		result.Issues = []control.PlanIssue{staleOptionsIssue()}
		return result, nil
	}
	if err != nil {
		return PlanResult{}, err
	}
	issues, err := normalizePlanIssues(output.Issues)
	if err != nil {
		return PlanResult{}, err
	}
	result.Issues = issues
	if hasBlockingIssues(issues) {
		return result, nil
	}
	plan, err := service.buildPlan(draft, current, resolved, output, issues)
	if err != nil {
		return PlanResult{}, err
	}
	if err := service.ensureBaseCurrent(ctx, request.Scope, draft); err != nil {
		return PlanResult{}, err
	}
	if err := service.repository.CreatePlan(ctx, plan); err != nil {
		return PlanResult{}, err
	}
	result.Plan = &plan
	return result, nil
}

func (service *Service) buildPlan(
	draft validatedDraft,
	current *control.ResourceRevision,
	resolved []control.ResolvedReference,
	output TargetPlanOutput,
	issues []control.PlanIssue,
) (control.Plan, error) {
	artifact, err := control.CanonicalJSONObject(output.ArtifactJSON)
	if err != nil {
		return control.Plan{}, err
	}
	artifactDigest, err := control.DigestCanonicalJSON(artifact)
	if err != nil {
		return control.Plan{}, err
	}
	draftHash, err := control.DigestCanonicalJSON(draft.result.CanonicalManifest)
	if err != nil {
		return control.Plan{}, err
	}
	changes, err := semanticChanges(current, draft.result.CanonicalManifest)
	if err != nil {
		return control.Plan{}, err
	}
	now := service.clock().UTC()
	plan := control.Plan{
		ID: service.idGenerator(), Scope: draft.scope,
		ActorID:   draft.scope.ActorID,
		Operation: draft.result.Operation, Target: draft.result.Target,
		DraftHash: draftHash, CanonicalManifest: draft.result.CanonicalManifest,
		ResolvedReferences: resolved, SemanticChanges: changes, Issues: issues,
		ArtifactKind: output.ArtifactKind, ArtifactJSON: json.RawMessage(artifact),
		ArtifactDigest: artifactDigest, OptionsRevision: output.OptionsRevision,
		CreatedAt: now, ExpiresAt: now.Add(service.planTTL),
		Status: control.PlanStatusPending,
	}
	if draft.head != nil {
		plan.TargetResourceID = draft.head.ID
		plan.BaseUID = draft.head.Identity.UID
		plan.BaseResourceVersion = draft.head.ResourceVersion
	}
	plan.PlanHash, err = control.ComputePlanHash(plan.HashInput())
	if err != nil {
		return control.Plan{}, err
	}
	if err := plan.Validate(); err != nil {
		return control.Plan{}, err
	}
	return plan, nil
}
