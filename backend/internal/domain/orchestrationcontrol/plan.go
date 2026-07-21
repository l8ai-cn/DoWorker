package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type Plan struct {
	ID                    string              `json:"id"`
	Scope                 Scope               `json:"scope"`
	ActorID               int64               `json:"actorId"`
	Operation             PlanOperation       `json:"operation"`
	Target                ResourceTarget      `json:"target"`
	TargetResourceID      int64               `json:"targetResourceId,omitempty"`
	BaseUID               string              `json:"baseUid,omitempty"`
	BaseResourceVersion   int64               `json:"baseResourceVersion,omitempty"`
	DraftHash             string              `json:"draftHash"`
	PlanHash              string              `json:"planHash"`
	CanonicalManifest     json.RawMessage     `json:"canonicalManifest"`
	ResolvedReferences    []ResolvedReference `json:"resolvedReferences"`
	SemanticChanges       []SemanticChange    `json:"semanticChanges"`
	Issues                []PlanIssue         `json:"issues"`
	ArtifactKind          string              `json:"artifactKind"`
	ArtifactJSON          json.RawMessage     `json:"artifactJson"`
	ArtifactDigest        string              `json:"artifactDigest"`
	OptionsRevision       string              `json:"optionsRevision"`
	CreatedAt             time.Time           `json:"createdAt"`
	ExpiresAt             time.Time           `json:"expiresAt"`
	Status                PlanStatus          `json:"status"`
	ConsumedAt            *time.Time          `json:"consumedAt,omitempty"`
	ConsumedByID          int64               `json:"consumedById,omitempty"`
	ResultResourceID      int64               `json:"resultResourceId,omitempty"`
	ResultIdentity        *ResourceIdentity   `json:"resultIdentity,omitempty"`
	ResultResourceVersion int64               `json:"resultResourceVersion,omitempty"`
	ResultRevision        int64               `json:"resultRevision,omitempty"`
	ResultJSON            json.RawMessage     `json:"resultJson,omitempty"`
}

func (plan Plan) HashInput() PlanHashInput {
	return PlanHashInput{
		Operation:           plan.Operation,
		Scope:               plan.Scope,
		Target:              plan.Target,
		BaseUID:             plan.BaseUID,
		BaseResourceVersion: plan.BaseResourceVersion,
		DraftHash:           plan.DraftHash,
		ResolvedReferences:  plan.ResolvedReferences,
		ArtifactDigest:      plan.ArtifactDigest,
		OptionsRevision:     plan.OptionsRevision,
	}
}

func (plan Plan) Validate() error {
	if err := validateUUID("plan.id", plan.ID); err != nil {
		return err
	}
	if err := plan.Target.Validate(plan.Scope); err != nil {
		return err
	}
	if plan.ActorID <= 0 || plan.ActorID != plan.Scope.ActorID {
		return invalid("plan.actorId", "must equal the authenticated actor")
	}
	if err := plan.Operation.validate(); err != nil {
		return err
	}
	if err := validatePlanBase(plan); err != nil {
		return err
	}
	if err := plan.validatePayload(); err != nil {
		return err
	}
	if err := validatePlanTimes(plan.CreatedAt, plan.ExpiresAt); err != nil {
		return err
	}
	if err := plan.Status.validate(); err != nil {
		return err
	}
	return plan.validateConsumption()
}

func (plan Plan) validatePayload() error {
	if !digestPattern.MatchString(plan.DraftHash) {
		return invalid("plan.draftHash", "must be a lowercase SHA-256 digest")
	}
	if _, err := sortedResolvedReferences(plan.Scope, plan.ResolvedReferences); err != nil {
		return err
	}
	for index := range plan.SemanticChanges {
		if err := plan.SemanticChanges[index].Validate(); err != nil {
			return err
		}
	}
	for index := range plan.Issues {
		if err := plan.Issues[index].Validate(); err != nil {
			return err
		}
	}
	if err := plan.validateManifest(); err != nil {
		return err
	}
	if err := plan.validateArtifact(); err != nil {
		return err
	}
	if err := validateOptionsRevision(plan.OptionsRevision); err != nil {
		return err
	}
	expected, err := ComputePlanHash(plan.HashInput())
	if err != nil {
		return err
	}
	if expected != plan.PlanHash {
		return corrupt("plan.planHash", "must match the versioned hash payload")
	}
	return nil
}

func (plan Plan) validateManifest() error {
	if err := rejectRawSecretJSON(plan.CanonicalManifest); err != nil {
		return invalid("plan.canonicalManifest", "must not contain raw secrets")
	}
	canonical, err := CanonicalJSONObject(plan.CanonicalManifest)
	if err != nil || !bytes.Equal(canonical, plan.CanonicalManifest) {
		return corrupt("plan.canonicalManifest", "must be canonical object JSON")
	}
	var manifest orchestrationresource.Manifest
	if err := json.Unmarshal(plan.CanonicalManifest, &manifest); err != nil {
		return corrupt("plan.canonicalManifest", "must satisfy the resource contract")
	}
	if err := manifest.ValidateSubmission(); err != nil {
		return corrupt("plan.canonicalManifest", "must satisfy the resource contract")
	}
	if manifest.TypeMeta != plan.Target.TypeMeta ||
		manifest.Metadata.Namespace != plan.Target.Namespace ||
		manifest.Metadata.Name != plan.Target.Name {
		return corrupt("plan.canonicalManifest", "must match the authenticated target")
	}
	expected, err := DigestCanonicalJSON(plan.CanonicalManifest)
	if err != nil || expected != plan.DraftHash {
		return corrupt("plan.draftHash", "must match the canonical manifest")
	}
	return nil
}

func (plan Plan) validateArtifact() error {
	if err := (orchestrationresource.TypeMeta{
		APIVersion: orchestrationresource.APIVersionV1Alpha1,
		Kind:       plan.ArtifactKind,
	}).Validate(); err != nil {
		return invalid("plan.artifactKind", "must use the resource kind grammar")
	}
	if err := rejectRawSecretJSON(plan.ArtifactJSON); err != nil {
		return invalid("plan.artifactJson", "must not contain raw secrets")
	}
	canonical, err := CanonicalJSONObject(plan.ArtifactJSON)
	if err != nil || !bytes.Equal(canonical, plan.ArtifactJSON) {
		return corrupt("plan.artifactJson", "must be canonical object JSON")
	}
	expected, err := DigestCanonicalJSON(plan.ArtifactJSON)
	if err != nil || expected != plan.ArtifactDigest {
		return corrupt("plan.artifactDigest", "must match canonical artifact JSON")
	}
	return nil
}

func validatePlanBase(plan Plan) error {
	if err := validateBaseState(
		plan.Operation,
		plan.BaseUID,
		plan.BaseResourceVersion,
	); err != nil {
		return err
	}
	switch plan.Operation {
	case PlanOperationCreate:
		if plan.TargetResourceID != 0 {
			return invalid("plan.targetResourceId", "must be empty for create")
		}
	case PlanOperationUpdate:
		if plan.TargetResourceID <= 0 {
			return invalid("plan.targetResourceId", "must be positive for update")
		}
	}
	return nil
}
