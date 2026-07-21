package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/google/uuid"
)

type Scope struct {
	OrganizationID   int64        `json:"organizationId"`
	OrganizationSlug slugkit.Slug `json:"organizationSlug"`
	ActorID          int64        `json:"actorId"`
}

type ResourceTarget struct {
	orchestrationresource.TypeMeta
	Namespace slugkit.Slug `json:"namespace"`
	Name      slugkit.Slug `json:"name"`
}

type ResourceIdentity struct {
	ResourceTarget
	UID string `json:"uid"`
}

type ResolvedReference struct {
	orchestrationresource.TypeMeta
	Namespace slugkit.Slug `json:"namespace"`
	Name      slugkit.Slug `json:"name"`
	UID       string       `json:"uid"`
	Revision  int64        `json:"revision"`
	Digest    string       `json:"digest"`
}

type ResourceHead struct {
	ID              int64             `json:"id"`
	OrganizationID  int64             `json:"organizationId"`
	Identity        ResourceIdentity  `json:"identity"`
	DisplayName     string            `json:"displayName"`
	Labels          map[string]string `json:"labels"`
	Status          json.RawMessage   `json:"status"`
	Revision        int64             `json:"revision"`
	Generation      int64             `json:"generation"`
	ResourceVersion int64             `json:"resourceVersion"`
	CreatedByID     int64             `json:"createdById"`
	UpdatedByID     int64             `json:"updatedById"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

func (scope Scope) Validate() error {
	if scope.OrganizationID <= 0 {
		return invalid("scope.organizationId", "must be positive")
	}
	if err := slugkit.Validate(scope.OrganizationSlug.String()); err != nil {
		return invalid("scope.organizationSlug", "must be a valid identifier")
	}
	if scope.ActorID <= 0 {
		return invalid("scope.actorId", "must be positive")
	}
	return nil
}

func (target ResourceTarget) Validate(scope Scope) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if err := target.TypeMeta.Validate(); err != nil {
		return invalid("target.typeMeta", "must satisfy the resource contract")
	}
	if err := slugkit.Validate(target.Namespace.String()); err != nil {
		return invalid("target.namespace", "must be a valid identifier")
	}
	if target.Namespace != scope.OrganizationSlug {
		return invalid("target.namespace", "must equal the authenticated organization slug")
	}
	if err := slugkit.Validate(target.Name.String()); err != nil {
		return invalid("target.name", "must be a valid identifier")
	}
	return nil
}

func (identity ResourceIdentity) Validate(scope Scope) error {
	if err := identity.ResourceTarget.Validate(scope); err != nil {
		return err
	}
	return validateUUID("identity.uid", identity.UID)
}

func (reference ResolvedReference) Validate(scope Scope) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	phaseOne := orchestrationresource.Reference{
		APIVersion: reference.APIVersion,
		Kind:       reference.Kind,
		Namespace:  reference.Namespace,
		Name:       reference.Name,
		UID:        reference.UID,
		Revision:   reference.Revision,
		Digest:     reference.Digest,
	}
	if err := phaseOne.ValidateResolved(scope.OrganizationSlug.String()); err != nil {
		return invalid("resolvedReference", "must satisfy the resource reference contract")
	}
	return validateUUID("resolvedReference.uid", reference.UID)
}

func (reference ResolvedReference) sortKey() string {
	return strings.Join([]string{
		reference.APIVersion,
		reference.Kind,
		reference.Namespace.String(),
		reference.Name.String(),
		reference.UID,
		reference.revisionKey(),
		reference.Digest,
	}, "\x00")
}

func (reference ResolvedReference) duplicateKey() string {
	return strings.Join([]string{
		reference.APIVersion,
		reference.Kind,
		reference.Namespace.String(),
		reference.Name.String(),
		reference.UID,
		reference.revisionKey(),
	}, "\x00")
}

func (reference ResolvedReference) revisionKey() string {
	return fmtInt64(reference.Revision)
}

func (head ResourceHead) Validate(scope Scope) error {
	if head.ID <= 0 {
		return invalid("resourceHead.id", "must be positive")
	}
	if err := validateOrganization(scope, head.OrganizationID); err != nil {
		return err
	}
	if err := head.Identity.Validate(scope); err != nil {
		return err
	}
	metadata := orchestrationresource.Metadata{
		Name:            head.Identity.Name,
		Namespace:       head.Identity.Namespace,
		DisplayName:     head.DisplayName,
		Labels:          head.Labels,
		UID:             head.Identity.UID,
		ResourceVersion: fmtInt64(head.ResourceVersion),
		Generation:      head.Generation,
	}
	if err := metadata.Validate(); err != nil {
		return invalid("resourceHead.metadata", "must satisfy the resource contract")
	}
	canonicalStatus, err := CanonicalJSONObject(head.Status)
	if err != nil || !bytes.Equal(canonicalStatus, head.Status) {
		return corrupt("resourceHead.status", "must be canonical object JSON")
	}
	if err := rejectRawSecretJSON(head.Status); err != nil {
		return corrupt("resourceHead.status", "must not contain raw secrets")
	}
	if err := validateCounters(
		head.Revision,
		head.Generation,
		head.ResourceVersion,
	); err != nil {
		return err
	}
	if head.CreatedByID <= 0 || head.UpdatedByID <= 0 {
		return invalid("resourceHead.actor", "identifiers must be positive")
	}
	if err := validateTimeRange(
		"resourceHead",
		head.CreatedAt,
		head.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

func validateUUID(field, value string) error {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return invalid(field, "must be a canonical UUID")
	}
	return nil
}
