package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type ResourceRevision struct {
	OrganizationID       int64               `json:"organizationId"`
	ResourceID           int64               `json:"resourceId"`
	Identity             ResourceIdentity    `json:"identity"`
	Revision             int64               `json:"revision"`
	Generation           int64               `json:"generation"`
	ResourceVersion      int64               `json:"resourceVersion"`
	CanonicalManifest    json.RawMessage     `json:"canonicalManifest"`
	CanonicalSpec        json.RawMessage     `json:"canonicalSpec"`
	ResolvedReferences   []ResolvedReference `json:"resolvedReferences"`
	Digest               string              `json:"digest"`
	WorkerSpecSnapshotID int64               `json:"workerSpecSnapshotId,omitempty"`
	ActorID              int64               `json:"actorId"`
	CreatedAt            time.Time           `json:"createdAt"`
}

func (revision ResourceRevision) Validate(scope Scope) error {
	if err := validateOrganization(scope, revision.OrganizationID); err != nil {
		return err
	}
	if revision.ResourceID <= 0 {
		return invalid("resourceRevision.resourceId", "must be positive")
	}
	if err := revision.Identity.Validate(scope); err != nil {
		return err
	}
	if err := validateCounters(
		revision.Revision,
		revision.Generation,
		revision.ResourceVersion,
	); err != nil {
		return err
	}
	if revision.ActorID <= 0 {
		return invalid("resourceRevision.actorId", "must be positive")
	}
	if revision.WorkerSpecSnapshotID < 0 {
		return invalid("resourceRevision.workerSpecSnapshotId", "must not be negative")
	}
	if revision.CreatedAt.IsZero() {
		return invalid("resourceRevision.createdAt", "must not be zero")
	}
	if _, err := revision.CreatedAt.MarshalJSON(); err != nil {
		return invalid("resourceRevision.createdAt", "must be encodable")
	}
	if _, err := sortedResolvedReferences(scope, revision.ResolvedReferences); err != nil {
		return err
	}
	return revision.validateDocuments()
}

func (revision ResourceRevision) validateDocuments() error {
	manifestJSON, err := CanonicalJSONObject(revision.CanonicalManifest)
	if err != nil || !bytes.Equal(manifestJSON, revision.CanonicalManifest) {
		return corrupt("resourceRevision.canonicalManifest", "must be canonical object JSON")
	}
	specJSON, err := CanonicalJSONObject(revision.CanonicalSpec)
	if err != nil || !bytes.Equal(specJSON, revision.CanonicalSpec) {
		return corrupt("resourceRevision.canonicalSpec", "must be canonical object JSON")
	}
	if err := rejectRawSecretJSON(revision.CanonicalManifest); err != nil {
		return corrupt("resourceRevision.canonicalManifest", "must not contain raw secrets")
	}

	var manifest orchestrationresource.Manifest
	if err := json.Unmarshal(revision.CanonicalManifest, &manifest); err != nil {
		return corrupt("resourceRevision.canonicalManifest", "must satisfy the resource contract")
	}
	if err := manifest.ValidateStored(); err != nil {
		return corrupt("resourceRevision.canonicalManifest", "must satisfy the resource contract")
	}
	if !revision.matchesManifest(manifest) {
		return corrupt("resourceRevision.canonicalManifest", "must match revision identity and counters")
	}
	manifestSpec, err := CanonicalJSONObject(manifest.Spec)
	if err != nil || !bytes.Equal(manifestSpec, revision.CanonicalSpec) {
		return corrupt("resourceRevision.canonicalSpec", "must match canonical manifest spec")
	}
	expectedDigest, err := DigestCanonicalJSON(revision.CanonicalManifest)
	if err != nil || expectedDigest != revision.Digest {
		return corrupt("resourceRevision.digest", "must match canonical manifest")
	}
	return nil
}

func (revision ResourceRevision) matchesManifest(
	manifest orchestrationresource.Manifest,
) bool {
	return manifest.TypeMeta == revision.Identity.TypeMeta &&
		manifest.Metadata.Namespace == revision.Identity.Namespace &&
		manifest.Metadata.Name == revision.Identity.Name &&
		manifest.Metadata.UID == revision.Identity.UID &&
		manifest.Metadata.Generation == revision.Generation &&
		manifest.Metadata.ResourceVersion == strconv.FormatInt(revision.ResourceVersion, 10)
}

func validateOrganization(scope Scope, organizationID int64) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if organizationID <= 0 || organizationID != scope.OrganizationID {
		return invalid("organizationId", "must equal the authenticated organization")
	}
	return nil
}

func validateCounters(revision, generation, resourceVersion int64) error {
	if revision <= 0 || generation <= 0 || resourceVersion <= 0 {
		return invalid("resource counters", "must be positive")
	}
	if generation > revision {
		return invalid("generation", "must not exceed revision")
	}
	if resourceVersion < revision {
		return invalid("resourceVersion", "must not be less than revision")
	}
	return nil
}
